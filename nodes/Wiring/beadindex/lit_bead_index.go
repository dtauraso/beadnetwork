package beadindex

import "math"

func LitBeadIndex(t float64, steps int) (int, bool) {
	if t < 0 || t >= 1 || steps <= 0 {
		return 0, false
	}

	const eps = 1e-9
	idx := int(math.Floor(t*float64(steps) + eps))
	if idx < 0 {
		idx = 0
	}
	if idx >= steps {
		idx = steps - 1
	}
	return idx, true
}

func BeadPlacementOffset(base, step float64, i int) float64 {
	return base + float64(i)*step
}

func PulsePlacementOffset(base, step, t float64, steps int) float64 {
	lastSlot := float64(steps - 1)
	if lastSlot < 0 {
		lastSlot = 0
	}
	return base + t*lastSlot*step
}
