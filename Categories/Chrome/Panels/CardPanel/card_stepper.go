package CardPanel

import (
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
)

const (
	tickValueField = "-000 +00"

	StepArrowW = 14
	StepGap    = 2
)

func stepW(f Field) float32 {
	if !HasTick(f) {
		return 0
	}
	return StepGap + StepArrowW
}

func stepArrows(f Field, cell Rect) (up, down Rect) {
	if !HasTick(f) {
		return Rect{}, Rect{}
	}
	x := cell.X + cell.W + StepGap
	half := cell.H / 2
	return Rect{X: x, Y: cell.Y, W: StepArrowW, H: half}, Rect{X: x, Y: cell.Y + half, W: StepArrowW, H: half}
}

func TickText(start, tick int) string {
	return strconv.Itoa(start) + " +" + strconv.Itoa(tick)
}

func (s State) Shown(f Field) string {
	if HasTick(f) {
		return TickText(s.Value(f), s.Cards[f.Node].Get(TickOf(f)))
	}
	return strconv.Itoa(s.Value(f))
}

func (l Layout) HitStep(x, y float64) (Field, int, bool) {
	if !l.Covers(x, y) {
		return Field{}, 0, false
	}
	for _, fb := range l.Fields {
		if fb.Up.W > 0 && Panel.HitRect(fb.Up, x, y) {
			return fb.Field, 1, true
		}
		if fb.Down.W > 0 && Panel.HitRect(fb.Down, x, y) {
			return fb.Field, -1, true
		}
	}
	return Field{}, 0, false
}
