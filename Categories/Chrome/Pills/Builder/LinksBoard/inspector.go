package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
)

const (
	ipadX      = 10
	ipadY      = 8
	inspW      = 260
	inspMinW   = 180
	colNarrowW = 320
	subH       = 15
	headTH     = 15
	rulerH     = 30
	minDial    = 48
)

type Paint = Board.Paint

type ins struct {
	l    *Board.Laid
	r    Rect
	clip uint8
	x, y float32
}

func (in *ins) w() float32 { return in.r.W - 2*ipadX }

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
	x := in.x + l.Text(in.x, in.y, h, in.clip, d.Title, Board.Px, Board.Ink, true, Board.Left) + 6
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

func (in *ins) axis(a Axis) {
	l, h := in.l, Board.RowH()
	inc := fieldW(a.Inc.Field)
	if a.Inc.Act.Op == "" {
		inc = segW(a.Inc)
	}
	ctrl := max(segW(a.Turn), Board.W(a.IncTag, Board.Px)+6+inc, Board.W(a.Label, 11))
	size, beside := a.DialSize, true
	if size > 0 && in.w()-10-ctrl < size {
		size = in.w() - 10 - ctrl
		if size < minDial {
			size, beside = min(a.DialSize, in.w()), false
		}
	}
	if a.Ends != "" && a.DialSize > 0 {
		in.summary(a)
		return
	}
	if size < 70 {
		a.Dial.Labels = false
	}
	y0 := in.y
	if a.DialSize > 0 {
		dr := Rect{X: in.x, Y: in.y, W: size, H: size}
		if !beside {
			dr.X = in.x + (in.w()-size)/2
		}
		l.Dial(dr, in.clip, a.Dial)
		axis := Sheet.DialPhi
		if a.Dial.Axis == Board.Theta {
			axis = Sheet.DialTheta
		}
		if a.DialAct.Op != "" {
			l.Spot(dr, in.clip, Sheet.Cell{Act: a.DialAct, Dial: Sheet.Dial{Axis: axis}}, Sheet.PartBody)
		}
	}
	x, y := in.x+size+10, in.y
	if a.DialSize == 0 {
		x = in.x
	} else if !beside {
		x, y = in.x, in.y+size+6
		y0 = y
		size = 0
	}
	l.Weighted(x, y, headTH, in.clip, a.Label, 11, Board.Mute, Board.Semi, Board.Left)
	y += headTH
	l.Stepper(x, y, in.clip, a.Turn)
	y += h + 4
	if a.Inc.Steps.Op != "" {
		tw := l.Text(x, y, h, in.clip, a.IncTag, Board.Px, Board.Mute, false, Board.Left) + 6
		if a.Inc.Act.Op != "" {
			l.Field(x+tw, y, in.clip, a.Inc.Field)
		} else {
			l.Stepper(x+tw, y, in.clip, a.Inc)
		}
		y += h + 4
	}
	in.y = y
	for _, s := range a.Lines {
		in.lines(x, s, 11, Board.Mute)
	}
	in.y = max(in.y, y0+size) + 8
}
