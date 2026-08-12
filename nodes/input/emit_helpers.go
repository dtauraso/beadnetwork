// emit_helpers.go — the pure computation Update/updateFeedbackRing (node.go) lean on:
// popEnd's double-buffer arithmetic and cadenceTicks' tick-count derivation. Neither
// touches a Node field — popEnd works on the caller's own local slice pointers, and
// cadenceTicks takes the already-read step count as a plain int. Lifted out of node.go
// per the primitive landing rule's narrowed clause: node.go stays this kind's home (the
// struct, the builder/RegisterBuilder call, and the loop/decision bodies that read/write
// this node's own fields), this file holds only the arithmetic those bodies call.
package input

import "github.com/dtauraso/wirefold/nodes/wire/lattice"

// popEnd reads and removes the END element of working, refilling from backup
// when working empties. working/backup are the double-buffer: each is a fresh
// copy of init, and end-popping [1,0] yields 0 then 1. Returns the popped value.
// Caller guarantees len(working) > 0 (refill keeps it non-empty when init != nil).
func popEnd(working, backup *[]int, init []int) int {
	v := (*working)[len(*working)-1]
	*working = (*working)[:len(*working)-1]
	if len(*working) == 0 {
		// Refill: the top row (backup) slides down to become the new working
		// row; a fresh top row appears.
		*working = *backup
		*backup = append([]int(nil), init...)
	}
	return v
}

// cadenceTicks is the pure tick-count derivation inputCadenceTicks (node.go) wraps: the
// CROSSING TIME of an edge of the given step count, steps * DwellTicksPerBead
// (docs/bead-model/bead-lattice.md "Timing"), floored at 1 so a zero/degenerate step count
// still fires every cycle rather than never.
func cadenceTicks(steps int) int64 {
	c := int64(float64(steps) * lattice.DwellTicksPerBead)
	if c < 1 {
		return 1
	}
	return c
}
