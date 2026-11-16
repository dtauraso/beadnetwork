package Builder

import (
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/LinksBoard"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const (
	OpNodeIdx    = "node-idx"
	OpNodeRef    = "node-ref"
	OpAddNode    = "add-node"
	OpRemoveNode = "remove-node"
	OpConnect    = "connect"
	OpDisconnect = "disconnect"
)

func netData(m Model) *LinksBoard.Net {
	n := &LinksBoard.Net{Back: Act{Op: OpNetFold}, Add: Act{Op: OpAddNode}, IDs: m.Nodes}
	for _, id := range m.Nodes {
		col := m.Col(id)
		idx := func(k, v, label string) Board.Field {
			return Board.Field{Value: v, Fit: "000.00", Act: Act{Op: OpNodeIdx, Node: id, K: k}, Tip: label}
		}
		ref := func(k string, v int, label string) Board.Field {
			return Board.Field{Value: strconv.Itoa(v), Fit: "000", Act: Act{Op: OpNodeRef, Node: id, K: k}, Tip: label}
		}
		c := m.Frame.At(col.At)
		menu := Act{Op: OpMenu, K: menuNode, Node: id}
		n.Nodes = append(n.Nodes, LinksBoard.NetNode{
			Name: "node " + id, Sub: linkCount(len(m.linksTo(id))) + " in",
			Menu: menu, MenuOpen: m.Menu == menu,
			Drop:  Board.Chip{Label: "Remove node " + id, Act: Act{Op: OpRemoveNode, Node: id}, Tip: "and its links"},
			Place: []Board.Field{idx(Links.KR, Num(col.At.R), "r"), idx(Links.KPhi, Num(col.At.Phi), "φ"), idx(Links.KTheta, Num(col.At.Theta), "θ")},
			C:     "C " + Num(c.X) + " " + Num(c.Y) + " " + Num(c.Z),
			Ref:   []Board.Field{ref(Links.KPhi, col.Ref.Phi, "Δφ"), ref(Links.KTheta, col.Ref.Theta, "Δθ")},
		})
		var row []LinksBoard.NetCell
		for _, to := range m.Nodes {
			_, has := m.Link(id, to)
			cell := LinksBoard.NetCell{Self: to == id, Linked: has, Act: Act{Op: OpConnect, Node: id, To: to}}
			if has {
				cell.Act.Op = OpDisconnect
			}
			row = append(row, cell)
		}
		n.Cells = append(n.Cells, row)
	}
	return n
}
