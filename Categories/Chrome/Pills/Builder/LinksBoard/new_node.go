package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
)

type NewNode struct {
	Add     Act
	Centers []Board.Chip
	At      []NetField
	Drops   []Board.Chip
}

func newNode(l *Board.Laid, r Rect, sy float32, n NewNode) float32 {
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
