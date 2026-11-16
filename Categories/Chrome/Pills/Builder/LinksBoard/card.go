package LinksBoard

import (
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
)

const (
	cardPadX  = 8
	cardPadY  = 6
	dialPx    = 36
	col0      = 40
	col1      = 14
	gapX      = 6
	gapY      = 4
	headTxtH  = 14
	footH     = 16
	smallCol0 = 28
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

func segW(s Board.Stepper) float32 {
	v := s.Plain
	if s.Act.Op != "" {
		v = s.Value
	}
	return 34 + max(28, Board.W(v, Board.BtnPx)+10)
}

func card(l *Board.Laid, clip uint8, x, y, w float32, d Card) float32 {
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
		var aw, bw float32
		for _, r := range d.Rows {
			aw, bw = max(aw, segW(r.A)), max(bw, segW(r.B))
		}
		var tw, sw float32
		for _, r := range d.Rows {
			if r.Why == "" {
				tw = max(tw, Board.W(r.Total, 12))
			}
			if r.Start.Steps.Op != "" {
				sw = max(sw, segW(r.Start))
			}
		}
		if sw > 0 {
			sw += gapX
		}
		d0 := float32(col0)
		if d0+col1+sw+aw+bw+tw+4*gapX > cw {
			d0 = smallCol0
		}
		s := cx + d0 + gapX + col1 + gapX
		a := s + sw
		b := a + aw + gapX
		if sw > 0 {
			l.Weighted(s, cy, headTxtH, clip, d.StartHead, 10.5, Board.Mute, Board.Semi, Board.Left)
		}
		t := b + bw + gapX
		for i, s := range [3]float32{a, b, t} {
			l.Weighted(s, cy, headTxtH, clip, d.Heads[i], 10.5, Board.Mute, Board.Semi, Board.Left)
		}
		cy += headTxtH + gapY
		for _, r := range d.Rows {
			cy += row(l, clip, cx, cy, d0, s, a, b, t, r) + gapY
		}
		cy -= gapY
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

func row(l *Board.Laid, clip uint8, x, y, d0, s, a, b, t float32, r CardRow) float32 {
	h := float32(Board.SegH)
	dial := min(float32(dialPx), d0)
	if r.Dial != nil {
		h = max(h, dial)
	}
	mid := y + h/2
	switch {
	case r.Dial != nil:
		dr := Rect{X: x, Y: y, W: dial, H: dial}
		l.Dial(dr, clip, *r.Dial)
		axis := Sheet.DialPhi
		if r.Dial.Axis == Board.Theta {
			axis = Sheet.DialTheta
		}
		if r.DialAct.Op != "" {
			l.Spot(dr, clip, Sheet.Cell{Act: r.DialAct, Dial: Sheet.Dial{Axis: axis}, Tip: "drag around the dial to set Δ · it snaps to nΔ"}, Sheet.PartBody)
		}
	case r.Ruler != nil:
		l.Ruler(Rect{X: x, Y: mid - 7, W: d0, H: 14}, clip, *r.Ruler)
	}
	l.Text(x+d0+gapX, y, h, clip, r.Axis, Board.Px, Board.Ink, true, Board.Left)
	if r.Start.Steps.Op != "" {
		l.Seg(s, mid-Board.SegH/2, clip, r.Start)
	}
	l.Seg(a, mid-Board.SegH/2, clip, r.A)
	if r.Why != "" {
		l.Text(b, y, h, clip, r.Why, 11, Board.Mute, false, Board.Left)
		return h
	}
	if r.B.Steps.Op != "" {
		l.Seg(b, mid-Board.SegH/2, clip, r.B)
	}
	l.Text(t, y, h, clip, r.Total, 12, Board.Mute, false, Board.Left)
	return h
}
