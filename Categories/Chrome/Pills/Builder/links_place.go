package Builder

import (
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

func placeRows(t *tree, m Model, id string) {
	fold := Act{Op: OpPlaceFold, Node: id}
	r := m.Col(id)
	on := m.PlaceOpen[id]
	row := linkRow{id: fold, depth: 2, open: openOf(on), fold: fold, name: Text("place")}
	if !on {
		row.detail = []Cell{Text("I r " + Num(r.At.R) + " · φ " + Num(r.At.Phi) + " · θ " + Num(r.At.Theta)).Dimmed(true)}
	}
	t.add(row)
	if !on {
		return
	}
	act := func(k string) Act { return Act{Op: OpNodeIdx, Node: id, K: k} }
	head := func(s string) Cell { return Text(s).Inked(inkHead) }
	c := m.Frame.At(r.At)
	t.put(Group{Card: true, Depth: 3, Rows: [][]Cell{
		{Text(""), head("r"), head("φ"), head("θ")},
		{Text("I").Dimmed(true),
			Field("", Num(r.At.R), act(Links.KR), true, true).Outline(false),
			Field("", Num(r.At.Phi), act(Links.KPhi), true, true).Outline(false),
			Field("", Num(r.At.Theta), act(Links.KTheta), true, true).Outline(false)},
		{Text(""), head("x"), head("y"), head("z")},
		{Text("C").Dimmed(true), Text(Num(c.X)), Text(Num(c.Y)), Text(Num(c.Z))},
		{Text("n").Dimmed(true), Text(m.NOf(id))},
	}})
}
