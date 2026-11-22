package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
)

type Net struct {
	Nodes []NetNode
	IDs   []string
	Cells [][]NetCell
}

type NetNode struct {
	ID   string
	X, Y float32
}

type NetField struct {
	Name  string
	Field Board.Field
}

type NetCell struct {
	Self, Linked, Picked bool
	Act                  Act
}

const netHeadW = 104

func network(l *Board.Laid, r Rect, sy float32, n Net, menus *[]placed) float32 {
	clip := l.Clip(r)
	l.Add(Board.Mark{Kind: Board.KBox, R: r, Clip: clip, Fill: Board.Back, Alpha: 1})
	w := max(160, r.W-2*colGap)
	x, y := r.X+colGap, r.Y+colGap-sy
	if len(n.Nodes) > 0 {
		gh := min(150, max(90, w*0.3))
		c := openFitCard(l, clip, x, y, w, "network", "seen along z · bright arrows are picked")
		graph(l, clip, Rect{X: x + cpad, Y: c.cy, W: w - 2*cpad, H: gh}, n)
		c.cy += gh + Board.Gap
		y = c.close()
	}
	if len(n.IDs) > 1 {
		c := openFitCard(l, clip, x, y, w, "links", "each node, then the nodes it feeds · click one to unlink")
		for i, row := range n.Cells {
			var items []fitItem
			for j, cell := range row {
				if cell.Linked {
					items = append(items, c.chip(Board.Chip{Label: n.IDs[j], Act: cell.Act, Lit: true}))
				}
			}
			for j, cell := range row {
				if !cell.Linked && !cell.Self {
					items = append(items, c.faded(Board.Chip{Label: "＋" + n.IDs[j], Act: cell.Act}))
				}
			}
			c.lead(n.IDs[i]+"  →", items)
		}
		y = c.close()
	}
	return max(0, y+sy-(r.Y+r.H))
}
