package Builder

import (
	"fmt"
	"math"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

type CardStep struct {
	Title string
	Lines []string
	Note  string
	OK    bool
}

type Method struct {
	ID   string
	Name string
	Run  func(m Model, item string) []CardStep
}

var Methods = []Method{
	{ID: MethodAtoB, Name: "a→b card: C_n + d_nb, one per row", Run: methodAtoB},
	{ID: MethodPair, Name: "pair card: C_b = C_a + d_ab, then index and place", Run: methodPair},
}

func f2(v float64) string {
	if math.Abs(v) < 5e-4 {
		v = 0
	}
	return fmt.Sprintf("%.2f", math.Round(v*100)/100)
}

func pt(p Links.Vec3) string { return "(" + f2(p.X) + ", " + f2(p.Y) + ", " + f2(p.Z) + ")" }

func idx(i Links.Place) string {
	return "(I_r " + f2(i.R) + ", I_φ " + f2(i.Phi) + ", I_θ " + f2(i.Theta) + ")"
}

func placeText(i Links.Place) string {
	return "O + polar_to_cart(" + f2(i.R) + "·c_r, " + f2(i.Phi) + "·c_φ, " + f2(i.Theta) + "·c_θ)"
}

type row struct {
	from string
	t    Links.Start
	d    Links.Vec3
	at   Links.Place
	c    Links.Vec3
	b    Links.Vec3
}

func rowsInto(m Model, item string) []row {
	var out []row
	for _, l := range m.allLinks() {
		if l.To != item {
			continue
		}
		s := m.ScaleOf(l.To)
		t := l.Dir(s)
		at := m.Reports[l.From].At
		c := m.Frame.At(at)
		d := m.Frame.Vec(s, t)
		out = append(out, row{from: l.From, t: t, d: d, at: at, c: c, b: c.Add(d)})
	}
	return out
}

func vectorStep(rows []row, item string) CardStep {
	st := CardStep{Title: "local: each row's start as a vector", Note: "from the row's start ticks alone; a vector, not yet a place", OK: len(rows) > 0}
	for _, r := range rows {
		st.Lines = append(st.Lines, fmt.Sprintf("d_%s%s = polar_to_cart(%s·ρ, %s·2π/12s_%s, %s·2π/12s_%s) = %s",
			r.from, item, Num(r.t.R), Num(r.t.Phi), item, Num(r.t.Theta), item, pt(r.d)))
	}
	if len(rows) == 0 {
		st.Lines = []string{item + "'s table has no rows"}
	}
	return st
}

func fromNodes(rows []row) CardStep {
	st := CardStep{Title: "the from node's position, from its index", Note: "each row's vector starts at its from node, so the from node's position is needed", OK: true}
	for _, r := range rows {
		st.Lines = append(st.Lines, fmt.Sprintf("C_%s = %s = %s", r.from, placeText(r.at), pt(r.c)))
	}
	return st
}

func result(m Model, item string, p Links.Vec3, label string) []string {
	miss := p.Sub(m.Frame.At(m.Reports[item].At)).Length()
	lands := " — misses it"
	if miss < 1e-6 {
		lands = " — lands on it"
	}
	return []string{label + " as an index: " + idx(m.Frame.IndexOf(p)), "|" + label + " − " + item + " as set| = " + f2(miss) + lands}
}

func methodAtoB(m Model, item string) []CardStep {
	rows := rowsInto(m, item)
	steps := []CardStep{vectorStep(rows, item)}
	if len(rows) == 0 {
		return steps
	}
	steps = append(steps, fromNodes(rows))
	var lines []string
	for _, r := range rows {
		lines = append(lines, fmt.Sprintf("%s(%s) = C_%s + d_%s%s = %s", item, r.from, r.from, r.from, item, pt(r.b)))
	}
	for _, r := range rows[1:] {
		gap := r.b.Sub(rows[0].b).Length()
		agree := " — the rows disagree"
		if gap < 1e-6 {
			agree = " — one point"
		}
		lines = append(lines, fmt.Sprintf("|%s(%s) − %s(%s)| = %s%s", item, r.from, item, rows[0].from, f2(gap), agree))
	}
	for _, r := range rows {
		lines = append(lines, result(m, item, r.b, item+"("+r.from+")")...)
	}
	return append(steps, CardStep{Title: "global: " + item + " from each row", Lines: lines, OK: true,
		Note: "O→n + n→" + item + ", once per row (the a→b card); nothing is added at O; the index is exact, not rounded"})
}

func methodPair(m Model, item string) []CardStep {
	rows := rowsInto(m, item)
	steps := []CardStep{vectorStep(rows, item)}
	if len(rows) == 0 {
		return steps
	}
	steps = append(steps, fromNodes(rows))
	glob := CardStep{Title: "global: C_" + item + " = C_n + d_n" + item + ", per pair", Note: "in " + item + ", with C_n from the edge n→" + item, OK: true}
	conv := CardStep{Title: "convert: C_" + item + " − O as an index", OK: true, Note: "kept exact — not rounded to whole index steps; C_b = O gives I_b = 0",
		Lines: []string{"I_r = |C_b − O| / c_r,  I_φ = acos((C_b − O)_y / |C_b − O|) / c_φ,  I_θ = atan2((C_b − O)_z, (C_b − O)_x) / c_θ"}}
	place := CardStep{Title: "place " + item + " from the index", Note: "the pair card's C_b", OK: true}
	for _, r := range rows {
		glob.Lines = append(glob.Lines, fmt.Sprintf("pair %s→%s: C_%s = C_%s + d_%s%s = %s", r.from, item, item, r.from, r.from, item, pt(r.b)))
		i := m.Frame.IndexOf(r.b)
		conv.Lines = append(conv.Lines, fmt.Sprintf("pair %s→%s: I_%s = %s", r.from, item, item, idx(i)))
		cb := m.Frame.At(i)
		back := cb.Sub(r.b).Length()
		holds := " — the index lost the point"
		if back < 1e-6 {
			holds = " — the index holds the same point"
		}
		place.Lines = append(place.Lines,
			fmt.Sprintf("pair %s→%s: C_%s = %s = %s", r.from, item, item, placeText(i), pt(cb)),
			fmt.Sprintf("|placed − C_%s + d_%s%s| = %s%s", r.from, r.from, item, f2(back), holds))
		place.Lines = append(place.Lines, result(m, item, cb, "C_"+item)[1:]...)
	}
	return append(steps, glob, conv, place)
}
