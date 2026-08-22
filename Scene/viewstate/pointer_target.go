package viewstate

import "github.com/dtauraso/wirefold/Chrome/Panels/Panel"

type PointerKind uint8

const (
	PointerNothing PointerKind = iota
	PointerInteractive
	PointerRefusing
)

type PointerTarget struct {
	Rect Panel.Rect
	Kind PointerKind
	Tip  string
}

const (
	TipGapY    = 6
	TipPadX    = 6
	TipPadY    = 3
	TipCharW   = 5.6
	TipHeight  = 18
	TipViewPad = 4
)

func (t PointerTarget) TipRect(viewW float32) (x, y, w, h float32) {
	if t.Tip == "" {
		return 0, 0, 0, 0
	}
	w = float32(len([]rune(t.Tip)))*TipCharW + 2*TipPadX
	h = TipHeight
	x = t.Rect.X
	y = t.Rect.Y + t.Rect.H + TipGapY
	if right := x + w; viewW > 0 && right > viewW-TipViewPad {
		x = viewW - TipViewPad - w
	}
	if x < TipViewPad {
		x = TipViewPad
	}
	return x, y, w, h
}
