package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
)

const dialGap, ctrlPadR = 12, 4

func gridded(a Axis) bool { return a.Ruler == nil && (a.Ends == "" || a.DialSize == 0) }

func tagW(a Axis) float32 {
	w := Board.W(a.IncTag, Board.Px)
	if w > 0 {
		w += 6
	}
	return w
}

func ctrlW(a Axis, tag float32) float32 {
	inc := fieldW(a.Inc.Field)
	if a.Inc.Act.Op == "" {
		inc = segW(a.Inc)
	}
	return max(tag+max(segW(a.Turn), inc), Board.W(a.Label, 11))
}

func (in *ins) grid(axes []Axis) {
	for _, a := range axes {
		if gridded(a) {
			in.slot, in.tag = max(in.slot, a.DialSize), max(in.tag, tagW(a))
		} else if a.Ruler == nil {
			in.sum = max(in.sum, min(a.DialSize, in.w()))
		}
	}
	for _, a := range axes {
		if gridded(a) {
			in.ctrl = max(in.ctrl, ctrlW(a, in.tag))
		}
	}
	if in.slot > 0 && in.w()-dialGap-in.ctrl < in.slot {
		in.slot = -1
	}
}

func (in *ins) dial(a Axis, dr Rect) {
	if dr.W < 70 {
		a.Dial.Labels = false
	}
	in.l.Dial(dr, in.clip, a.Dial)
	axis := Sheet.DialPhi
	if a.Dial.Axis == Board.Theta {
		axis = Sheet.DialTheta
	}
	if a.DialAct.Op != "" {
		in.l.Spot(dr, in.clip, Sheet.Cell{Act: a.DialAct, Dial: Sheet.Dial{Axis: axis}}, Sheet.PartBody)
	}
}

func (in *ins) left() float32 {
	right := in.r.X + in.r.W - ctrlPadR
	x := right - max(in.ctrl, in.sum)
	if in.ctrl > 0 && in.slot > 0 {
		x -= dialGap + in.slot
	}
	if x >= right {
		return in.x
	}
	return max(in.x, x)
}

func (in *ins) axis(a Axis) {
	if !gridded(a) {
		in.summary(a)
		return
	}
	l, h := in.l, Board.RowH()
	x, y, bottom := in.x, in.y, in.y
	if in.slot < 0 && a.DialSize > 0 {
		size := min(a.DialSize, in.w())
		in.dial(a, Rect{X: in.x + (in.w()-size)/2, Y: y, W: size, H: size})
		y += size + 6
	}
	right := in.r.X + in.r.W - ctrlPadR
	x = max(x, right-in.ctrl)
	if in.slot > 0 && a.DialSize > 0 {
		in.dial(a, Rect{X: x - dialGap - (in.slot+a.DialSize)/2, Y: y, W: a.DialSize, H: a.DialSize})
		bottom = y + a.DialSize
	}
	l.Weighted(x, y, headTH, in.clip, a.Label, 11, Board.Mute, Board.Semi, Board.Left)
	y += headTH
	l.Stepper(right-segW(a.Turn), y, in.clip, a.Turn)
	y += h + 4
	if a.Inc.Steps.Op != "" {
		w := segW(a.Inc)
		if a.Inc.Act.Op != "" {
			w = fieldW(a.Inc.Field)
			l.Field(right-w, y, in.clip, a.Inc.Field)
		} else {
			l.Stepper(right-w, y, in.clip, a.Inc)
		}
		l.Text(right-w-tagW(a), y, h, in.clip, a.IncTag, Board.Px, Board.Mute, false, Board.Left)
		y += h + 4
	}
	in.y = y
	for _, s := range a.Lines {
		in.lines(x, s, 11, Board.Mute)
	}
	in.y = max(in.y, bottom) + 8
}
