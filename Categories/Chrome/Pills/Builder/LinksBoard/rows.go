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
	pickH    = 18
	chipPadX = 5
	cardFitW = 240
	radioPx  = 12
	lookGap  = 12

	lookShare = 0.3
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
	l.Text(x, y, nodeH, k.clip, n.Sub, 12, Board.Mute, false, Board.Left)
	l.Spot(Rect{X: k.x, Y: y, W: k.w, H: nodeH}, k.clip, Sheet.Cell{Act: n.Fold}, Sheet.PartBody)
	k.y += nodeH
	if !n.Open {
		return
	}
	for _, lk := range n.Links {
		k.link(lk)
	}
}

func pickDot(l *Board.Laid, clip uint8, x, mid float32, lk Link) {
	edge := Board.Edge
	if lk.Picked {
		edge = Board.Accent
	}
	l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: x, Y: mid - radioPx/2, W: radioPx, H: radioPx}, Clip: clip, Fill: Board.White, Edge: edge, Radius: radioPx / 2, Stroke: 1, Alpha: 1})
	if lk.Picked {
		l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: x + 3, Y: mid - radioPx/2 + 3, W: radioPx - 6, H: radioPx - 6}, Clip: clip, Fill: Board.Accent, Radius: (radioPx - 6) / 2, Alpha: 1})
	}
	if lk.PickOpen.Op == "" {
		l.Spot(Rect{X: x - 4, Y: mid - lrowH/2, W: radioPx + 8, H: lrowH}, clip, Sheet.Cell{Act: lk.Pick}, Sheet.PartBody)
	}
}

func (k *col) look(row Rect, lk Link) {
	l, w := k.l, row.W*lookShare
	r := Rect{X: row.X + row.W - w, Y: row.Y, W: w, H: row.H}
	l.Line(Rect{X: r.X + 0.5, Y: r.Y, W: 0, H: r.H}, k.clip, Board.Guide, 1, false)
	cw := Board.W("look", Board.SmallPx) + 2*chipPadX
	cr := Rect{X: k.x + k.w - cpad - cw, Y: row.Y + (lrowH-pickH)/2, W: cw, H: pickH}
	l.Box(cr, k.clip, Board.White, Board.Edge)
	l.Text(cr.X+cw/2, cr.Y, pickH, k.clip, "look", Board.SmallPx, Board.Mute, false, Board.Center)
	l.Spot(Rect{X: r.X - lookGap, Y: r.Y, W: lookGap, H: r.H}, k.clip, Sheet.Cell{Act: lk.Gap}, Sheet.PartBody)
	l.Spot(r, k.clip, Sheet.Cell{Act: lk.Fold}, Sheet.PartBody)
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
	act := lk.Fold
	if lk.PickOpen.Op != "" {
		act = lk.PickOpen
	}
	l.Spot(row, k.clip, Sheet.Cell{Act: act}, Sheet.PartBody)
	x := k.x + lindent
	mid := k.y + lrowH/2
	if lk.Pick.Op != "" {
		pickDot(l, k.clip, x, mid, lk)
		x += radioPx + 6
	} else {
		x += dotPx + 6
	}
	x += l.Text(x, k.y, lrowH, k.clip, lk.Label, Board.Px, Board.Ink, false, Board.Left) + 6
	l.Dial(Rect{X: x, Y: mid - miniPx/2, W: miniPx, H: miniPx}, k.clip, lk.Dial)
	if lk.Hot && !lk.Open && lk.PickOpen.Op != "" {
		k.look(row, lk)
	} else {
		l.Text(k.x+k.w-cpad, k.y, lrowH, k.clip, lk.Sum, 12, Board.Mute, false, Board.Right)
	}
	k.y += lrowH
	if lk.Card != nil {
		cx, cw := k.x+lindent, k.w-lindent-cpad
		if cw < max(cardFitW, cardW(*lk.Card)) {
			cx, cw = k.x+cpad, k.w-2*cpad
		}
		k.y += card(l, k.clip, cx, k.y, cw, *lk.Card) + 6
	}
}
