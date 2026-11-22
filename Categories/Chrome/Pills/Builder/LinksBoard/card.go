package LinksBoard

import (
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/LinksBoard/Fill"
)

const (
	cardPadX = 8
	cardPadY = 6
	col1     = 16
	gapX     = 6
	gapY     = 4
	headTxtH = 14
	footH    = 16
)

func wrap(s string, w, px float32) []string {
	var out []string
	line := ""
	for _, word := range strings.Fields(s) {
		next := strings.TrimSpace(line + " " + word)
		if line != "" && Board.W(next, px) > w {
			out = append(out, line)
			next = word
		}
		line = next
	}
	if line != "" {
		out = append(out, line)
	}
	return out
}

func segW(s Board.Stepper) float32 { return Board.SegW(s) }

func headW(s string) float32 { return Board.TextW(s, 10.5, Board.Semi) }

type cardWidths struct{ lw, sw, ow, aw, bw, tw, g float32 }

func cardCols(d Fill.Card) cardWidths {
	var c cardWidths
	for _, r := range d.Rows {
		c.lw = max(c.lw, Board.TextW(r.Axis, Board.Px, 7))
		c.aw, c.bw = max(c.aw, segW(r.A)), max(c.bw, segW(r.B))
		if r.C.Steps.Op != "" {
			c.tw = max(c.tw, segW(r.C))
		} else if r.Why == "" {
			c.tw = max(c.tw, Board.W(r.Total, 12))
		}
		if r.Start.Steps.Op != "" {
			c.sw = max(c.sw, segW(r.Start))
		}
		if r.One.Steps.Op != "" {
			c.ow = max(c.ow, segW(r.One), headW(d.OneHead))
		}
	}
	hs, h0, h1 := headW(d.StartHead), headW(d.Heads[0]), headW(d.Heads[1])
	c.tw = max(c.tw, headW(d.Heads[2]))
	c.g = max(float32(gapX), (h0+h1-c.aw-c.bw)/2+gapX, (h1-c.bw)/2+gapX)
	if c.sw > 0 {
		c.g = max(c.g, (hs+h0-c.sw-c.aw)/2+gapX)
	}
	return c
}

func cardW(d Fill.Card) float32 {
	if len(d.Rows) == 0 {
		return 0
	}
	c := cardCols(d)
	w := 2*cardPadX + c.lw + c.g + c.aw + c.g + c.bw + c.g + c.tw
	if c.sw > 0 {
		w += c.sw + c.g
	}
	rest := c.bw + c.g + c.tw
	if c.ow > 0 {
		w += c.ow + c.g
		rest += c.ow + c.g
	}
	for _, r := range d.Rows {
		if r.Why != "" {
			w = max(w, w-rest+Board.W(r.Why, 11))
		}
	}
	return w
}

func card(l *Board.Laid, clip uint8, x, y, w float32, d Fill.Card) float32 {
	box := len(l.Marks)
	l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: x, Y: y, W: w}, Clip: clip, Fill: Board.White, Edge: Board.Guide, Radius: radius6, Corner: 1, Stroke: 1, Alpha: 1})
	l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: x + 1, Y: y - 0.5, W: w - 2, H: 2}, Clip: clip, Fill: Board.White, Alpha: 1})
	cx, cy, cw := x+cardPadX, y+cardPadY, w-2*cardPadX
	lines := func(s string, px float32, ink Board.Paint, x0 float32) {
		for _, t := range wrap(s, cw-(x0-cx), px) {
			l.Text(x0, cy, footH, clip, t, px, ink, false, Board.Left)
			cy += footH
		}
	}
	if d.Note != "" {
		lines(d.Note, 12, Board.Mute, cx)
	}
	if len(d.Rows) > 0 {
		c := cardCols(d)
		lw, sw, ow, aw, bw, g := c.lw, c.sw, c.ow, c.aw, c.bw, c.g
		gx, gclip, thumb := cx, clip, func() {}
		if most := cardW(d) - 2*cardPadX - cw; most > 0 && d.Scroll.Op != "" {
			gx, gclip, thumb = sideScroll(l, clip, Rect{X: cx, Y: cy, W: cw, H: headTxtH + float32(len(d.Rows))*(Board.SegH+gapY)}, most, d, &cy)
		}
		clip := gclip
		s := gx + lw + g
		a := s
		if sw > 0 {
			a = s + sw + g
			l.Weighted(s+sw/2, cy, headTxtH, clip, d.StartHead, 10.5, Board.Mute, Board.Semi, Board.Center)
		}
		b := a + aw + g
		o := b + bw + g
		t := o
		if ow > 0 {
			t = o + ow + g
			l.Weighted(o+ow/2, cy, headTxtH, clip, d.OneHead, 10.5, Board.Mute, Board.Semi, Board.Center)
		}
		l.Weighted(a+aw/2, cy, headTxtH, clip, d.Heads[0], 10.5, Board.Mute, Board.Semi, Board.Center)
		l.Weighted(b+bw/2, cy, headTxtH, clip, d.Heads[1], 10.5, Board.Mute, Board.Semi, Board.Center)
		l.Weighted(t, cy, headTxtH, clip, d.Heads[2], 10.5, Board.Mute, Board.Semi, Board.Left)
		cy += headTxtH + gapY
		for _, r := range d.Rows {
			cy += row(l, clip, gx, cy, s, o, a, b, t, r) + gapY
		}
		cy -= gapY
		thumb()
	}
	if d.Fixed != "" {
		cy += gapY
		cw0 := Board.W("🔒 r fixed", 10.5) + 10
		l.Box(Rect{X: cx, Y: cy + 1, W: cw0, H: footH - 2}, clip, Board.On, Board.Edge)
		l.Text(cx+cw0/2, cy, footH, clip, "🔒 r fixed", 10.5, Board.Ink, false, Board.Center)
		parts := wrap(d.Fixed, cw-cw0-4, 11)
		if len(parts) > 0 {
			l.Text(cx+cw0+4, cy, footH, clip, parts[0], 11, Board.Mute, false, Board.Left)
		}
		cy += footH
		if len(parts) > 1 {
			lines(strings.Join(parts[1:], " "), 11, Board.Mute, cx)
		}
	}
	if d.Foot != "" {
		cy += gapY
		lines(d.Foot, 11, Board.Mute, cx)
	}
	h := cy + cardPadY - y
	l.Marks[box].R.H = h
	return h
}

func row(l *Board.Laid, clip uint8, x, y, s, o, a, b, t float32, r Fill.CardRow) float32 {
	h := float32(Board.SegH)
	mid := y + h/2
	l.Text(x, y, h, clip, r.Axis, Board.Px, Board.Ink, true, Board.Left)
	if r.Start.Steps.Op != "" {
		l.Seg(s, mid-Board.SegH/2, clip, r.Start)
	}
	if r.One.Steps.Op != "" {
		l.Seg(o, mid-Board.SegH/2, clip, r.One)
	}
	l.Seg(a, mid-Board.SegH/2, clip, r.A)
	if r.Why != "" {
		l.Text(b, y, h, clip, r.Why, 11, Board.Mute, false, Board.Left)
		return h
	}
	if r.B.Steps.Op != "" {
		l.Seg(b, mid-Board.SegH/2, clip, r.B)
	}
	if r.C.Steps.Op != "" {
		l.Seg(t, mid-Board.SegH/2, clip, r.C)
		return h
	}
	l.Text(t, y, h, clip, r.Total, 12, Board.Mute, false, Board.Left)
	return h
}
