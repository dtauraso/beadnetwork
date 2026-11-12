package Builder

import (
	"math"
	"strconv"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const TreeLabel = "Tree"

func n3(v float64) string {
	r := math.Round(v*1000) / 1000
	if r == 0 {
		r = 0
	}
	return strconv.FormatFloat(r, 'f', -1, 64)
}

func ang(phi, theta float64, r *float64) string {
	s := "φ " + n3(phi) + " | θ " + n3(theta)
	if r != nil {
		s += " | r " + n3(*r)
	}
	return s
}

func ticks(t Links.Ints) string {
	return "φ " + strconv.Itoa(t.Phi) + " | θ " + strconv.Itoa(t.Theta) + " | r " + strconv.Itoa(t.R) + " | s " + strconv.Itoa(t.S)
}

func lockText(l Links.Lock) string {
	if l == (Links.Lock{}) {
		return "—"
	}
	return "φ " + strconv.Itoa(l.Phi) + " | θ " + strconv.Itoa(l.Theta) + " | r " + strconv.Itoa(l.R)
}

func names(xs []string) string {
	if len(xs) == 0 {
		return "—"
	}
	return strings.Join(xs, ", ")
}

func kv(k, v string, ref bool) []Cell { return []Cell{Text(k), LitText(v, Act{}, ref && v != "—")} }

func notStored() []Cell { return []Cell{Text("— not stored —")} }

func box(depth int, label string, rows ...[]Cell) Group {
	return Group{Title: strings.Repeat("    ", depth) + label, Rows: rows}
}

func linkNames(ls []Links.Link) []string {
	var out []string
	for _, l := range ls {
		out = append(out, l.Name())
	}
	return out
}

func TreePill(m Model) Pill {
	f := m.Frame
	var seqNames []string
	open := "—"
	for _, q := range m.Seqs {
		seqNames = append(seqNames, q.Name)
		if q.K == m.OpenSeq {
			open = q.Name
		}
	}
	all := m.allLinks()
	var recalc []Links.Link
	for _, l := range all {
		if !l.Picked {
			recalc = append(recalc, l)
		}
	}
	var methods []string
	for _, id := range MethodIDs {
		if m.Cards.Methods[id] {
			methods = append(methods, id)
		}
	}
	groups := []Group{
		box(0, "builder", kv("s", n3(float64(m.S)), false), kv("ρ", n3(f.Rho), false), kv("T_φ", n3(float64(f.Tphi)), false),
			kv("T_θ", n3(float64(f.Ttheta)), false), kv("c_r", n3(f.Cr), false), kv("sequences", names(seqNames), false), kv("open", open, true)),
	}
	if m.Root != "" {
		groups = append(groups, folderGroups(m, ".", 1)...)
	}
	groups = append(groups,
		box(1, "what runs now", kv("edges", names(linkNames(all)), true), kv("nodes", names(m.Nodes), true),
			kv("picked", names(linkNames(m.Picked())), true), notStored(), kv("recalculated", names(linkNames(recalc)), true)),
		box(2, "method cards", kv("item", m.Item(), true), kv("methods", names(methods), false), kv("card step", strconv.Itoa(m.CardShown()), false)),
	)
	for _, l := range all {
		d := l.DirFrom(m.Frame, m.RefOf(l.From))
		groups = append(groups, box(2, "edge "+l.Name(), kv("from", l.From, true), kv("to", l.To, true),
			kv("start", ang(l.Start.Phi, l.Start.Theta, &l.Start.R), false), kv("Δ", ticks(l.Delta), false),
			kv("increment", ticks(l.Increment), false), kv("lock", lockText(l.Lock), false),
			notStored(), kv("start + Δ", ang(d.Phi, d.Theta, &d.R), false)))
	}
	for _, id := range m.Nodes {
		r := m.Reports[id]
		var out, in []string
		for _, l := range all {
			if l.From == id {
				out = append(out, l.Name())
			}
			if l.To == id {
				in = append(in, l.Name())
			}
		}
		c := f.At(r.At)
		groups = append(groups, box(2, "node "+id,
			kv("at on O's grid", "r "+n3(r.At.R)+" | φ "+n3(r.At.Phi)+" | θ "+n3(r.At.Theta), false),
			kv("out", names(out), true), kv("in", names(in), true),
			kv("all lock", lockText(m.StepOf(id).Locks()), false), notStored(),
			kv("C", "x "+n3(c.X)+" | y "+n3(c.Y)+" | z "+n3(c.Z), false)))
	}
	return Pill{Label: TreeLabel, Bottom: true, Groups: groups}
}
