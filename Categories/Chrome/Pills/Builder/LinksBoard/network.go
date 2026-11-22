package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/LinksBoard/Fill"
)

const netHeadW = 104

func network(l *Board.Laid, r Rect, sy float32, n Fill.Net, menus *[]placed) float32 {
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

func newNode(l *Board.Laid, r Rect, sy float32, n Fill.NewNode) float32 {
	clip := l.Clip(r)
	l.Add(Board.Mark{Kind: Board.KBox, R: r, Clip: clip, Fill: Board.Back, Alpha: 1})
	w := max(160, r.W-2*colGap)
	x, y := r.X+colGap, r.Y+colGap-sy
	c := openFitCard(l, clip, x, y, w, "new node", "placed from a center node, as a link from it would be")
	var centers []fitItem
	for _, ch := range n.Centers {
		centers = append(centers, c.chip(ch))
	}
	c.row("center", centers)
	var at []fitItem
	for _, f := range n.At {
		at = append(at, c.field(f.Name, f.Field))
	}
	c.row("angles and r", at)
	c.row("", []fitItem{c.chip(Board.Chip{Label: "＋ add node", Act: n.Add, Lit: true})})
	y = c.close()
	c = openFitCard(l, clip, x, y, w, "remove node", "removes the node and every link to or from it")
	var drops []fitItem
	for _, ch := range n.Drops {
		drops = append(drops, c.chip(ch))
	}
	c.row("", drops)
	y = c.close()
	return max(0, y+sy-(r.Y+r.H))
}
