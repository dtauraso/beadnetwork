package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
)

const (
	nodeH    = 24
	lrowH    = 26
	lindent  = 22
	miniPx   = 20
	dotPx    = 8
	cardFitW = 240
)

func (k *col) node(n Node) {
	l, y := k.l, k.y
	k.rule(y-1, Board.Hair)
	glyph := "▸"
	if n.Open {
		glyph = "▾"
	}
	x := k.x + cpad
	l.Text(x+6, y, nodeH, k.clip, glyph, 11, Board.Mute, false, Board.Center)
	x += 12 + 6
	x += l.Weighted(x, y, nodeH, k.clip, n.Label+" ", 12, Board.Mute, Board.Semi, Board.Left)
	l.Weighted(x, y, nodeH, k.clip, n.Sub, 12, Board.Mute, Board.Semi, Board.Left)
	l.Spot(Rect{X: k.x, Y: y, W: k.w, H: nodeH}, k.clip, Sheet.Cell{Act: n.Fold, Tip: "open or fold this node's links in this step"}, Sheet.PartBody)
	k.y += nodeH
	if !n.Open {
		return
	}
	for _, lk := range n.Links {
		k.link(lk)
	}
}

func (k *col) link(lk Link) {
	l := k.l
	row := Rect{X: k.x + 1, Y: k.y, W: k.w - 2, H: lrowH}
	switch {
	case lk.Open:
		l.Add(Board.Mark{Kind: Board.KBox, R: row, Clip: k.clip, Fill: Board.On, Alpha: 1})
		l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: row.X, Y: row.Y, W: 2, H: lrowH}, Clip: k.clip, Fill: Board.Accent, Alpha: 1})
	case lk.Hot:
		l.Add(Board.Mark{Kind: Board.KBox, R: row, Clip: k.clip, Fill: Board.On, Alpha: 0.6})
	}
	l.Spot(row, k.clip, Sheet.Cell{Act: lk.Fold, Tip: lk.Tip}, Sheet.PartBody)
	x := k.x + lindent
	mid := k.y + lrowH/2
	if lk.Pick.Op != "" {
		fill, edge := Board.None, Board.Mute
		if lk.Picked {
			fill, edge = Board.Accent, Board.Accent
		}
		l.Add(Board.Mark{Kind: Board.KCircle, R: Rect{X: x + 0.75, Y: mid - dotPx/2 + 0.75, W: dotPx - 1.5, H: dotPx - 1.5}, Clip: k.clip, Fill: fill, Edge: edge, Stroke: 1.5, Alpha: 1})
		l.Spot(Rect{X: x - 4, Y: k.y, W: dotPx + 8, H: lrowH}, k.clip, Sheet.Cell{Act: lk.Pick, Tip: "pick: this link moves its target in this step"}, Sheet.PartBody)
	}
	x += dotPx + 6
	x += l.Text(x, k.y, lrowH, k.clip, lk.Label, Board.Px, Board.Ink, false, Board.Left) + 6
	l.Dial(Rect{X: x, Y: mid - miniPx/2, W: miniPx, H: miniPx}, k.clip, lk.Dial)
	l.Text(k.x+k.w-cpad, k.y, lrowH, k.clip, lk.Sum, 12, Board.Mute, false, Board.Right)
	k.y += lrowH
	if lk.Card != nil {
		cx, cw := k.x+lindent, k.w-lindent-cpad
		if cw < cardFitW {
			cx, cw = k.x+cpad, k.w-2*cpad
		}
		k.y += card(l, k.clip, cx, k.y, cw, *lk.Card) + 6
	}
}
