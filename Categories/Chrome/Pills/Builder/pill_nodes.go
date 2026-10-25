package Builder

import (
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const NodesLabel = "Node places"

const (
	OpNodeS      = "node-s"
	OpNodeIdx    = "node-idx"
	OpAddNode    = "add-node"
	OpRemoveNode = "remove-node"
	OpConnect    = "connect"
	OpDropLink   = "drop-link"
)

func NodesPill(m Model) Pill {
	var rows [][]Cell
	for _, id := range m.Nodes {
		r := m.Reports[id]
		act := func(op, k string) Act { return Act{Op: op, Node: id, K: k} }
		c := m.Frame.At(r.At)
		rows = append(rows, []Cell{
			Text(id),
			Field("s", strconv.Itoa(m.ScaleOf(id)), act(OpNodeS, ""), true, false),
			Field("I_r", Num(r.At.R), act(OpNodeIdx, Links.KR), true, true),
			Field("I_φ", Num(r.At.Phi), act(OpNodeIdx, Links.KPhi), true, true),
			Field("I_θ", Num(r.At.Theta), act(OpNodeIdx, Links.KTheta), true, true),
			Text("C x " + Num(c.X) + " y " + Num(c.Y) + " z " + Num(c.Z)),
			Button("remove", act(OpRemoveNode, ""), false),
		})
	}
	rows = append(rows, []Cell{Button("+ add node", Act{Op: OpAddNode}, false)})
	return Pill{Label: NodesLabel, Groups: []Group{
		{Title: "nodes — s (ticks a twelfth of a turn) and the place I, an index on the frame", Rows: rows},
		connectGroup(m),
	}}
}

func connectGroup(m Model) Group {
	var rows [][]Cell
	for _, from := range m.Nodes {
		row := []Cell{Text(from + " →")}
		for _, to := range m.Nodes {
			if _, has := m.Link(from, to); to != from && !has {
				row = append(row, Button(to, Act{Op: OpConnect, Node: from, To: to}, false))
			}
		}
		if len(row) > 1 {
			rows = append(rows, row)
		}
	}
	if len(rows) == 0 {
		rows = [][]Cell{{Text("no edge left to add")}}
	}
	return Group{Title: "+ add edge — its start is measured from where the two nodes are", Rows: rows}
}
