package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
)

const subLineH = 15

type fitCard struct {
	l       *Board.Laid
	clip    uint8
	x, y, w float32
	box     int
	cy      float32
}

type fitItem struct {
	w    float32
	draw func(x, y float32)
}

func openFitCard(l *Board.Laid, clip uint8, x, y, w float32, title, sub string) *fitCard {
	c := &fitCard{l: l, clip: clip, x: x, y: y, w: w, box: len(l.Marks)}
	l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: x, Y: y, W: w, H: cheadH}, Clip: clip, Fill: Board.White, Edge: Board.Guide, Radius: radius6, Stroke: 1, Alpha: 1})
	head := len(l.Marks)
	l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: x + 1, Y: y + 1, W: w - 2, H: cheadH - 1}, Clip: clip, Fill: Board.Card, Radius: radius6 - 1, Corner: 2, Alpha: 1})
	h := Board.RowH()
	hy, avail := y+(cheadH-h)/2, w-2*cpad
	tw := l.Text(x+cpad, hy, h, clip, title, Board.Px, Board.Ink, true, Board.Left) + Board.Gap
	headH := float32(cheadH)
	if tw+Board.W(sub, Board.SmallPx) <= avail {
		l.Text(x+cpad+tw, hy, h, clip, sub, Board.SmallPx, Board.Mute, false, Board.Left)
	} else {
		ly := hy + h
		for _, s := range wrap(sub, avail, Board.SmallPx) {
			l.Text(x+cpad, ly, subLineH, clip, s, Board.SmallPx, Board.Mute, false, Board.Left)
			ly += subLineH
		}
		headH = ly - y + (cheadH-h)/2
	}
	l.Marks[head].R.H = headH - 1
	c.cy = y + headH + Board.Gap
	return c
}

func (c *fitCard) row(label string, items []fitItem) {
	h := Board.RowH()
	cx, right := c.x+cpad, c.x+c.w-cpad
	start := cx
	if label != "" {
		c.l.Text(cx, c.cy, h, c.clip, label, Board.SmallPx, Board.Mute, false, Board.Left)
		start = max(cx+netHeadW, cx+Board.W(label, Board.SmallPx)+Board.Gap)
		if len(items) > 0 && start+items[0].w > right {
			c.cy += h + Board.Gap
			start = cx
		}
	}
	f := flow{x0: start, right: right, x: start, y: c.cy, h: h, gap: Board.Gap}
	for _, it := range items {
		ix, iy := f.at(it.w)
		it.draw(ix, iy)
	}
	c.cy = f.bottom() + Board.Gap
}

func (c *fitCard) close() float32 {
	bottom := c.cy - Board.Gap + cpad
	c.l.Marks[c.box].R.H = bottom - c.y
	return bottom + colGap
}

func (c *fitCard) chip(ch Board.Chip) fitItem {
	ch.Px = Board.BtnPx
	return fitItem{w: Board.W(ch.Label, Board.BtnPx) + 14, draw: func(x, y float32) { c.l.Chip(x, y, c.clip, ch) }}
}

func (c *fitCard) field(name string, f Board.Field) fitItem {
	nw := Board.W(name, Board.SmallPx) + 3
	h := Board.RowH()
	return fitItem{w: nw + Board.FieldW(f), draw: func(x, y float32) {
		c.l.Text(x, y, h, c.clip, name, Board.SmallPx, Board.Mute, false, Board.Left)
		c.l.Field(x+nw, y, c.clip, f)
	}}
}
