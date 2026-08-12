package wire

import "math"

func (pw *PacedWire) ticksToCross(steps int) float64 {
	if pw.dwell <= 0 {
		return 0
	}
	return float64(steps) * pw.dwell
}

func (pw *PacedWire) NextArrivalTick() (int64, bool) {
	best := int64(0)
	found := false
	for _, b := range pw.inflight {

		at := int64(math.Ceil(b.placementTick + pw.ticksToCross(b.steps)))
		if !found || at < best {
			best, found = at, true
		}
	}
	return best, found
}

func EarliestArrival(wires []*PacedWire) (int64, bool) {
	best := int64(0)
	found := false
	for _, w := range wires {
		if w == nil {
			continue
		}
		if at, ok := w.NextArrivalTick(); ok && (!found || at < best) {
			best, found = at, true
		}
	}
	return best, found
}
