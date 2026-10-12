package CardPanel

import (
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
)

const (
	StepArrowW = 14
	StepGap    = 2
)

func HasStep(f Field) bool {
	return (f.Vector == VecStart || f.Vector == VecEnd) && (f.Comp == CompPhi || f.Comp == CompTheta)
}

func Step(value, delta int) int {
	return ((value+delta)%TurnTicks + TurnTicks) % TurnTicks
}

func stepW(f Field) float32 {
	if !HasStep(f) {
		return 0
	}
	return StepGap + StepArrowW
}

func stepArrows(f Field, cell Rect) (up, down Rect) {
	if !HasStep(f) {
		return Rect{}, Rect{}
	}
	x := cell.X + cell.W + StepGap
	half := cell.H / 2
	return Rect{X: x, Y: cell.Y, W: StepArrowW, H: half}, Rect{X: x, Y: cell.Y + half, W: StepArrowW, H: half}
}

func (s State) Shown(f Field) string {
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
