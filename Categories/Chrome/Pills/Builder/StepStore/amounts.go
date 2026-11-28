package StepStore

import (
	"math"
	"os"
	"path/filepath"
	"slices"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links/Store"
)

const (
	DirAmounts    = "amounts"
	FileFrom      = "from.bin"
	FileTo        = "to.bin"
	FileAmtPhi    = "phi.bin"
	FileAmtTheta  = "theta.bin"
	FileAmtR      = "r.bin"
	FileAmtIncPhi = "inc-phi.bin"
	FileAmtIncTh  = "inc-theta.bin"
	FileAmtIncR   = "inc-r.bin"
	FileAmtPicked = "picked.bin"
)

type Edge struct{ From, To string }

type Amount struct {
	D, Inc Links.Ints
	Picked bool
}

func NewAmount() Amount { return Amount{Inc: Links.Ints{Phi: 1, Theta: 1, R: 1}} }

func (a *Amount) Step(k string, sign int) { *a.D.At(k) += sign * *a.Inc.At(k) }

func (a Amount) Presses() int {
	n := 0
	for _, k := range []string{Links.KPhi, Links.KTheta, Links.KR} {
		d, inc := *a.D.At(k), max(1, *a.Inc.At(k))
		n = max(n, int(math.Ceil(math.Abs(float64(d))/float64(inc))))
	}
	return n
}

func AmountsDir(stepDir string) string { return filepath.Join(stepDir, DirAmounts) }

func amountDir(stepDir string, e Edge) string {
	return filepath.Join(AmountsDir(stepDir), e.From+"To"+e.To)
}

func ReadAmounts(stepDir string) map[Edge]Amount {
	out := map[Edge]Amount{}
	entries, _ := os.ReadDir(AmountsDir(stepDir))
	for _, ent := range entries {
		if !ent.IsDir() { // path-resolution-ok: an amount is a directory; skipping strays under amounts/
			continue
		}
		dir := filepath.Join(AmountsDir(stepDir), ent.Name())
		var e Edge
		if !Store.ReadIfExists(filepath.Join(dir, FileFrom), &e.From) || !Store.ReadIfExists(filepath.Join(dir, FileTo), &e.To) {
			continue
		}
		a := NewAmount()
		for name, v := range map[string]*int{FileAmtPhi: &a.D.Phi, FileAmtTheta: &a.D.Theta, FileAmtR: &a.D.R,
			FileAmtIncPhi: &a.Inc.Phi, FileAmtIncTh: &a.Inc.Theta, FileAmtIncR: &a.Inc.R} {
			Store.ReadIfExists(filepath.Join(dir, name), v)
		}
		Store.ReadIfExists(filepath.Join(dir, FileAmtPicked), &a.Picked)
		out[e] = a
	}
	return out
}

func WriteAmount(stepDir string, e Edge, a Amount) error {
	dir := amountDir(stepDir, e)
	for name, v := range map[string]any{FileFrom: e.From, FileTo: e.To,
		FileAmtPhi: a.D.Phi, FileAmtTheta: a.D.Theta, FileAmtR: a.D.R,
		FileAmtIncPhi: a.Inc.Phi, FileAmtIncTh: a.Inc.Theta, FileAmtIncR: a.Inc.R, FileAmtPicked: a.Picked} {
		if err := Store.WriteAtomicIfChanged(filepath.Join(dir, name), v); err != nil {
			return err
		}
	}
	return nil
}

func RemoveAmount(stepDir string, e Edge) error { return os.RemoveAll(amountDir(stepDir, e)) }

func Leaves(root string, parent []int) [][]int {
	var out [][]int
	for _, n := range ReadStepNumbers(PathDir(root, parent)) {
		p := append(slices.Clone(parent), n)
		if kids := Leaves(root, p); len(kids) > 0 {
			out = append(out, kids...)
		} else {
			out = append(out, p)
		}
	}
	return out
}

func MoveAmounts(from, to string) error {
	if _, err := os.Stat(AmountsDir(from)); os.IsNotExist(err) {
		return nil
	}
	if err := os.MkdirAll(to, 0o755); err != nil {
		return err
	}
	return os.Rename(AmountsDir(from), AmountsDir(to))
}
