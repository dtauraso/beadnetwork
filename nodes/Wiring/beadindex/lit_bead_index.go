package beadindex

import "math"

// LitBeadIndex maps a bead's progress t (elapsed/ticksToCross, this edge's OWN t) onto the
// index of the chain bead it currently occupies, for a chain of the given STEP count. ok is
// false only when t is outside [0, 1) — off the edge entirely — never because the geometry
// ran out of beads: an index at or past steps is clamped onto the last bead rather than
// reported off-chain.
//
// steps must be the SAME integer chainBeads used to lay the chain out — two different
// lengths for layout vs. lighting is exactly the drift docs/bead-model/bead-lattice.md's one-integer
// model exists to make impossible (LiveBeadProgress.Steps travels with the bead precisely so
// this can never be re-derived from a second source).
//
//	t*steps = (elapsed/ticksToCross)*steps = elapsed/dwell
//
// which is the same for every edge (dwell is the uniform per-bead constant,
// lattice.DwellTicksPerBead), so each index lasts exactly one dwell everywhere.
//
// FLOOR, not round. The lit bead is the last one the traversal has reached, which is what
// floor means; round would instead light the NEAREST, and that ties exactly halfway between
// two beads — not academic here, since two edges reach the same distance via different t
// values and float error would decide the tie differently per edge
// (TestLitBeadIndexSameElapsedLightsSameBead pins this).
func LitBeadIndex(t float64, steps int) (int, bool) {
	if t < 0 || t >= 1 || steps <= 0 {
		return 0, false
	}
	// epsilon: t*steps is a float round-trip (t was itself elapsed/ticksToCross), so a
	// bead sitting EXACTLY on bead i's position can land a hair under it and floor to
	// i-1. A bead's own position is a reachable value, not an edge case, so nudge before
	// flooring. 1e-9 against an integer step index is far below anything visible and far
	// above float noise — the same epsilon the retired arc-length version used, kept
	// because the reasoning (a float round-trip, not the unit it multiplies) is unchanged.
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
