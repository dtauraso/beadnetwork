package Builder

import (
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const NodesLabel = "Node places"

const (
	OpNodeIdx    = "node-idx"
	OpNodeRef    = "node-ref"
	OpAddNode    = "add-node"
	OpRemoveNode = "remove-node"
	OpConnect    = "connect"
	OpDisconnect = "disconnect"
)

func NodesPill(m Model) Pill {
	var rows [][]Cell
	for _, id := range m.Nodes {
		r := m.Reports[id]
		act := func(op, k string) Act { return Act{Op: op, Node: id, K: k} }
		c := m.Frame.At(r.At)
		rows = append(rows, []Cell{
			Text(id),
			Field("I_r", Num(r.At.R), act(OpNodeIdx, Links.KR), true, true),
			Field("I_φ", Num(r.At.Phi), act(OpNodeIdx, Links.KPhi), true, true),
			Field("I_θ", Num(r.At.Theta), act(OpNodeIdx, Links.KTheta), true, true), Text("C x " + Num(c.X) + " y " + Num(c.Y) + " z " + Num(c.Z)),
			Text("n " + m.NOf(id)),
			Button("remove", act(OpRemoveNode, ""), false),
		})
	}
	rows = append(rows, []Cell{Button("+ add node", Act{Op: OpAddNode}, false)})
	return Pill{Label: NodesLabel, Bottom: true, Groups: []Group{
		{Title: "nodes — the place I, an index on the frame", Rows: rows},
	}}
}

func connectRow(m Model, from, label string) []Cell {
	row := []Cell{Text(label)}
	for _, to := range m.Nodes {
		if _, has := m.Link(from, to); to != from && !has {
			row = append(row, Button(to, Act{Op: OpConnect, Node: from, To: to}, false))
		}
	}
	if len(row) == 1 {
		return nil
	}
	return row
}
