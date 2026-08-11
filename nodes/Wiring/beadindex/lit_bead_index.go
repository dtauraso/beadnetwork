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

// BeadPlacementOffset is docs/bead-model/bead-lattice.md "Placement": the distance from the
// source node's own centre to bead i's centre along the chain's aim direction — base (the
// source-node's torus outer radius plus a bead's own torus outer radius, chainBeads' own
// `selfTorusR + lattice.BeadTorusOuterR`) plus i whole bead steps. Pure index arithmetic
// (base + i*step), moved out of chain_beads.go's placement loop alongside LitBeadIndex, the
// sibling progress->index math this file already holds.
func BeadPlacementOffset(base, step float64, i int) float64 {
	return base + float64(i)*step
}

// PulsePlacementOffset is BeadPlacementOffset evaluated at a CONTINUOUS index — t (this
// pulse's own [0,1) traversal fraction) times the chain's own last-slot index (steps-1,
// clamped to 0 for a one-bead chain) — instead of an integer i, so a travelling pulse rides
// through the same points the placeholder beads occupy without being rounded onto one of
// them. steps must be the SAME step count the pulse's own t was computed against
// (LiveBeadProgress.Steps travels with the bead, chain_beads.go's own doc comment on why:
// a node moved mid-flight must not make placement, timing and pacing read three different
// lengths).
func PulsePlacementOffset(base, step, t float64, steps int) float64 {
	lastSlot := float64(steps - 1)
	if lastSlot < 0 {
		lastSlot = 0
	}
	return base + t*lastSlot*step
}
