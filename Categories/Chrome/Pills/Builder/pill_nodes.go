package Builder

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/LinksBoard"
)

const (
	OpNodeRef    = "node-ref"
	OpConnect    = "connect"
	OpDisconnect = "disconnect"
)

func netData(m Model) *LinksBoard.Net {
	n := &LinksBoard.Net{IDs: m.Nodes}
	for _, id := range m.Nodes {
		c := m.Frame.At(m.Col(id).At)
		n.Nodes = append(n.Nodes, LinksBoard.NetNode{ID: id, X: float32(c.X), Y: float32(c.Y)})
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
