package Builder

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/LinksBoard"
)

const (
	OpNodeRef    = "node-ref"
	OpRemoveNode = "remove-node"
	OpConnect    = "connect"
	OpDisconnect = "disconnect"
)

func netData(m Model) *LinksBoard.Net {
	n := &LinksBoard.Net{Back: Act{Op: OpNetFold}, IDs: m.Nodes}
	for _, id := range m.Nodes {
		menu := Act{Op: OpMenu, K: menuNode, Node: id}
		c := m.Frame.At(m.Col(id).At)
		n.Nodes = append(n.Nodes, LinksBoard.NetNode{
			ID: id, X: float32(c.X), Y: float32(c.Y),
			Menu: menu, MenuOpen: m.Menu == menu,
			Drop: Board.Chip{Label: "Remove node " + id, Act: Act{Op: OpRemoveNode, Node: id}, Note: "and its links"},
		})
		var row []LinksBoard.NetCell
		for _, to := range m.Nodes {
			l, has := m.Link(id, to)
			cell := LinksBoard.NetCell{Self: to == id, Linked: has, Picked: has && l.Picked, Act: Act{Op: OpConnect, Node: id, To: to}}
			if has {
				cell.Act.Op = OpDisconnect
			}
			row = append(row, cell)
		}
		n.Cells = append(n.Cells, row)
	}
	return n
}
