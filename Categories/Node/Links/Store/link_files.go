package Store

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const (
	DirLinks = "links"
	DirDrag  = "drag"

	FileSource        = "source.bin"
	FileOrder         = "order.bin"
	FileStartPhi      = "start-phi-tick.bin"
	FileStartTheta    = "start-theta-tick.bin"
	fileStartPhiOld   = "start-phi.bin"
	fileStartThetaOld = "start-theta.bin"
	FileStartR        = "start-r.bin"
	FileDeltaPhi      = "delta-phi.bin"
	FileDeltaTheta    = "delta-theta.bin"
	FileDeltaR        = "delta-r.bin"
	FileIncPhi        = "increment-phi.bin"
	FileIncTheta      = "increment-theta.bin"
	FileIncR          = "increment-r.bin"
	FileLockPhi       = "lock-phi.bin"
	FileLockTheta     = "lock-theta.bin"
	FilePicked        = "picked.bin"
	FileStartS        = "start-s.bin"
	FileDeltaS        = "delta-s.bin"
	FileIncS          = "increment-s.bin"
	FileOnePhi        = "one-phi.bin"
	FileOneTheta      = "one-theta.bin"
	FileOneR          = "one-r.bin"
)

func BaseLinksDir(nodeDir string) string { return filepath.Join(nodeDir, DirLinks) }

func DragLinksDir(nodeDir string) string { return filepath.Join(nodeDir, DirDrag, DirLinks) }

func linkName(from, to string) string { return from + "To" + to }

func ReadLinks(nodeDir, to string) ([]Links.Link, error) {
	if n, ok, err := ReadNode(NodePath(nodeDir)); err != nil || ok {
		return n.Links(), err
	}
	if g, ok, err := ReadOldGrid(GridPath(nodeDir)); err != nil || ok {
		return g.ShownCol().node(to).Links(), err
	}
	if _, err := os.Stat(DragLinksDir(nodeDir)); err == nil { // path-resolution-ok: a scene from before grid.bin keeps its edits under drag/links
		return readLinksIn(DragLinksDir(nodeDir), to)
	}
	return readLinksIn(BaseLinksDir(nodeDir), to)
}

func ReadBaseLinks(nodeDir, to string) ([]Links.Link, error) {
	return readLinksIn(BaseLinksDir(nodeDir), to)
}

func ReadDragLinks(nodeDir, to string) ([]Links.Link, error) {
	return readLinksIn(DragLinksDir(nodeDir), to)
}

func readLinksIn(linksDir, to string) ([]Links.Link, error) {
	entries, err := os.ReadDir(linksDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var all []Links.Link
	for _, e := range entries {
		if !e.IsDir() { // path-resolution-ok: a link is a directory; skipping strays under links/
			continue
		}
		dir := filepath.Join(linksDir, e.Name())
		l := Links.Link{To: to, Increment: Links.Ints{Phi: 1, Theta: 1, R: 1}}
		if !ReadIfExists(filepath.Join(dir, FileSource), &l.From) {
			continue
		}
		ReadIfExists(filepath.Join(dir, FileOrder), &l.Order)
		ReadIfExists(filepath.Join(dir, FileStartR), &l.Start.R)
		ReadIfExists(filepath.Join(dir, FileDeltaPhi), &l.Delta.Phi)
		ReadIfExists(filepath.Join(dir, FileDeltaTheta), &l.Delta.Theta)
		ReadIfExists(filepath.Join(dir, FileDeltaR), &l.Delta.R)
		ReadIfExists(filepath.Join(dir, FileIncPhi), &l.Increment.Phi)
		ReadIfExists(filepath.Join(dir, FileIncTheta), &l.Increment.Theta)
		ReadIfExists(filepath.Join(dir, FileIncR), &l.Increment.R)
		ReadIfExists(filepath.Join(dir, FileLockPhi), &l.Lock.Phi)
		ReadIfExists(filepath.Join(dir, FileLockTheta), &l.Lock.Theta)
		ReadIfExists(filepath.Join(dir, FilePicked), &l.Picked)
		ReadIfExists(filepath.Join(dir, FileStartS), &l.Start.S)
		ReadIfExists(filepath.Join(dir, FileDeltaS), &l.Delta.S)
		ReadIfExists(filepath.Join(dir, FileIncS), &l.Increment.S)
		ReadIfExists(filepath.Join(dir, FileOnePhi), &l.Start.One.Phi)
		ReadIfExists(filepath.Join(dir, FileOneTheta), &l.Start.One.Theta)
		ReadIfExists(filepath.Join(dir, FileOneR), &l.Start.One.R)
		readStartTicks(dir, &l.Start)
		mustBeCounts(dir, l)
		all = append(all, l)
	}
	SortLinks(all)
	return all, nil
}

func readStartTicks(dir string, s *Links.Start) {
	if ReadIfExists(filepath.Join(dir, FileStartPhi), &s.Tick.Phi) && ReadIfExists(filepath.Join(dir, FileStartTheta), &s.Tick.Theta) {
		return
	}
	var phi, theta float64
	ReadIfExists(filepath.Join(dir, fileStartPhiOld), &phi)
	ReadIfExists(filepath.Join(dir, fileStartThetaOld), &theta)
	s.Tick = Links.Ticks{Phi: Links.LegacyTicks(phi, s.S), Theta: Links.LegacyTicks(theta, s.S)}
}

func mustBeCounts(dir string, l Links.Link) {
	const most = 1 << 40
	for _, v := range []int{l.Delta.Phi, l.Delta.Theta, l.Delta.R, l.Increment.Phi, l.Increment.Theta, l.Increment.R, l.Start.S, l.Delta.S, l.Increment.S, l.Start.One.Phi, l.Start.One.Theta, l.Start.One.R} {
		if v > most || v < -most {
			panic(fmt.Sprintf("Store.readLinksIn: %s holds %d in a Δ, nΔ or s file, which is a float64 read as an int64 — Δ and nΔ are whole numbers stored as int64, so this file was written before they were, or by something writing floats into them", dir, v))
		}
	}
}

func SortLinks(ls []Links.Link) {
	sort.SliceStable(ls, func(i, j int) bool {
		if ls[i].Order != ls[j].Order {
			return ls[i].Order < ls[j].Order
		}
		if ls[i].From != ls[j].From {
			return ls[i].From < ls[j].From
		}
		return ls[i].To < ls[j].To
	})
}

func WriteBaseLinks(nodeDir string, links []Links.Link) error {
	return writeLinksIn(BaseLinksDir(nodeDir), links)
}

func WriteLinks(nodeDir string, links []Links.Link) error {
	return writeLinksIn(DragLinksDir(nodeDir), links)
}

func RemoveLinksTo(nodeDir, from, to string) error {
	for _, dir := range []string{BaseLinksDir(nodeDir), DragLinksDir(nodeDir)} {
		if err := os.RemoveAll(filepath.Join(dir, linkName(from, to))); err != nil {
			return err
		}
	}
	return removeFromNode(nodeDir, from)
}

func writeLinksIn(linksDir string, links []Links.Link) error {
	keep := map[string]bool{}
	for _, l := range links {
		name := linkName(l.From, l.To)
		keep[name] = true
		dir := filepath.Join(linksDir, name)
		for file, v := range map[string]any{
			FileSource: l.From, FileOrder: l.Order,
			FileStartPhi: l.Start.Tick.Phi, FileStartTheta: l.Start.Tick.Theta, FileStartR: l.Start.R,
			FileDeltaPhi: l.Delta.Phi, FileDeltaTheta: l.Delta.Theta, FileDeltaR: l.Delta.R,
			FileIncPhi: l.Increment.Phi, FileIncTheta: l.Increment.Theta, FileIncR: l.Increment.R,
			FileLockPhi: l.Lock.Phi, FileLockTheta: l.Lock.Theta,
			FilePicked: l.Picked,
			FileStartS: l.Start.S, FileDeltaS: l.Delta.S, FileIncS: l.Increment.S,
			FileOnePhi: l.Start.One.Phi, FileOneTheta: l.Start.One.Theta, FileOneR: l.Start.One.R,
		} {
			if err := WriteAtomicIfChanged(filepath.Join(dir, file), v); err != nil {
				return err
			}
		}
	}
	if err := os.MkdirAll(linksDir, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(linksDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() && !keep[e.Name()] { // path-resolution-ok: removing a link the node no longer holds
			if err := os.RemoveAll(filepath.Join(linksDir, e.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}
