package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
)

const (
	ipadX      = 10
	ipadY      = 8
	inspW      = 260
	inspMinW   = 120
	colNarrowW = 320
	subH       = 15
	headTH     = 15
	rulerH     = 30
)

type Paint = Board.Paint

type ins struct {
	l    *Board.Laid
	r    Rect
	clip uint8
	x, y float32
	slot float32
	tag  float32
	ctrl float32
	sum  float32
}

func (in *ins) w() float32 { return in.r.W - 2*ipadX }

func inspMin(d Inspector) float32 {
	in := &ins{r: Rect{W: inspW}}
	in.grid(d.Axes)
	return max(inspMinW, in.ctrl+ipadX+ctrlPadR)
}

func (in *ins) lines(x float32, s string, px float32, ink Paint) {
	for _, t := range wrap(s, in.r.X+in.r.W-ipadX-x, px) {
		in.l.Text(x, in.y, subH, in.clip, t, px, ink, false, Board.Left)
		in.y += subH
	}
}

func inspector(l *Board.Laid, r Rect, sy float32, d Inspector) float32 {
	clip := l.Clip(r)
	l.Box(r, clip, Board.White, Board.None)
	l.Line(Rect{X: r.X + 0.5, Y: r.Y, W: 0, H: r.H}, clip, Board.Guide, 1, false)
	in := &ins{l: l, r: r, clip: clip, x: r.X + ipadX, y: r.Y + ipadY - sy}
	top := r.Y + ipadY
	if d.Title == "" {
		in.lines(in.x, d.Hint, 12, Board.Mute)
		return 0
	}
	in.grid(d.Axes)
	in.header(d)
	if d.Note != "" {
		in.lines(in.x, d.Note, 12, Board.Mute)
	}
	for _, a := range d.Axes {
		if a.Ruler != nil {
			in.ruler(a)
		} else {
			in.axis(a)
		}
	}
	if d.Hint != "" {
		in.hint(d.Hint)
	}
	return max(0, in.y+sy-top+ipadY-r.H)
}

func (in *ins) header(d Inspector) {
	l, h := in.l, Board.RowH()
	right := in.r.X + in.r.W - ipadX
	start := in.x
	if d.Switch.Act.Op == "" {
		start = max(in.x, min(in.left(), right-Board.W(d.Title, Board.Px)-6-Board.W(d.Sub, Board.Px)))
	}
	x := start + l.Text(start, in.y, h, in.clip, d.Title, Board.Px, Board.Ink, true, Board.Left) + 6
	if x+Board.W(d.Sub, Board.Px) > right {
		x, in.y = in.x, in.y+h
	}
	x += l.Text(x, in.y, h, in.clip, d.Sub, Board.Px, Board.Mute, false, Board.Left) + 6
	if d.Switch.Act.Op != "" {
		sw := Board.W(d.Switch.Label, Board.Px) + 2*Board.PadX
		mw := Board.W(d.Moves, Board.Px) + 6
		if x+sw > right {
			x, in.y = in.x, in.y+h+4
		}
		if right-sw-mw < x {
			in.y += h + 4
			l.Text(in.x, in.y, h, in.clip, d.Moves, Board.Px, Board.Mute, false, Board.Left)
			l.Chip(in.x+mw, in.y, in.clip, d.Switch)
		} else {
			l.Text(right-sw-6, in.y, h, in.clip, d.Moves, Board.Px, Board.Mute, false, Board.Right)
			l.Chip(right-sw, in.y, in.clip, d.Switch)
		}
	}
	in.y += h + 6
}
