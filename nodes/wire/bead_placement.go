// bead_placement.go — ONE JOB: the data a bead carries while it crosses a wire.
// The placement a sender hands over (beadPlacement), the in-channel record that
// carries it plus its placement tick (placeRequest), and the in-flight bead the
// wire's owner keeps until delivery (inflightBead). No timing math, no channel
// operations, no driving — those are arrival.go, paced_wire_send.go and
// paced_wire_drive.go.

package wire

// beadPlacement bundles everything one placement needs. Steps times delivery
// (docs/bead-lattice.md "Timing" — ticksToCross = steps * dwell, no length to
// divide); the segment endpoints + source identity drive the per-frame position
// stream. Geometry travels WITH the bead, never stored on the shared wire, so
// each in-flight bead evaluates the exact segment it is drawn on.
// The zero value (empty segment + identity) means "no position stream" — unit
// tests that only exercise delivery pass just Steps.
type beadPlacement struct {
	// Steps is this placement's bead-step count (docs/bead-lattice.md "The
	// count") — copied straight from the sending Out's own Geom().Steps at Send
	// time, so the wire never re-derives it from anything else.
	Steps int
	// Position-stream context. Start/End are this edge's straight-segment endpoints
	// (source OUT-port world pos, dest IN-port world pos). Node/Port are the SOURCE
	// node id + output port — the position trace key, matching the send event so the
	// renderer routes by source+sourceHandle (fan-out).
	Start, End Vec3
	Node, Port string
}

// streams reports whether this placement carries position-stream context. False
// for the bare-delivery placements used by unit tests (empty Node).
func (bp beadPlacement) streams() bool {
	return bp.Node != ""
}

// placeRequest is what Send hands across the wire's in-channel: a bead value,
// the placement geometry it should be timed/drawn against, and the tick to
// stamp it with. The SENDING node's own goroutine reads its own clock ONCE per
// emission and stamps placementTick here (MODEL.md's Clock bullet: "placement
// is decided by the emitting goroutine, at the moment it calls Send, from its
// own clock, once per emission — not re-derived later by whichever goroutine
// happens to drain the wire's in-channel"). The wire itself is a passive
// struct with no goroutine of its own — it is drained by its source node's
// mover (node_mover.go) — so it is no longer the reader. Moving the read to
// the sender is what lets several beads placed in the same emission (e.g. a
// broadcast fan-out) provably share one placementTick: reading a fresh clock
// value per wire in the drain pass could straddle a tick boundary, splitting
// one emission across two ticks.
type placeRequest struct {
	val           int
	bp            beadPlacement
	placementTick int64
}

// inflightBead is one bead traversing the wire. Each bead carries its own
// geometry so a mid-flight geometry edit (node-move) re-derives the remaining
// travel from the NEW step count while preserving the bead's FRACTIONAL progress t.
// Distance is NOT stored: fractional progress t = (clock.Tick() − placementTick)
// / ticksToCross is a pure function of the wire's own single clock reading
// (MODEL.md "Geometry and time").
//
// Every field here is touched by EXACTLY ONE goroutine: this wire's own (folded
// into edgeMover.run via driveOneCycle — see the PacedWire doc comment in
// paced_wire.go).
type inflightBead struct {
	val           int
	placementTick float64     // this wire's own tick reading when placed (fractional after a geometry rebase)
	steps         int         // current bead-step length of this bead's edge (docs/bead-lattice.md)
	seg           WireSegment // current straight-segment endpoints of this bead's edge
	node          string      // source node id — the position/cancel routing key
	port          string      // source output port — the position/cancel routing key
	streams       bool        // whether this bead carries position-stream context
	gen           uint64      // per-bead id (the bead's emitted identity)
	// finalPending is true once a drive cycle has advanced this bead to its
	// delivery deadline (target==deadline) but the handoff to outCh has not yet
	// succeeded (e.g. it was not yet at the FIFO head, or outCh had no room that
	// cycle). Subsequent cycles retry only the (cheap) delivery handoff for such
	// a bead, without re-running the position-advance math.
	finalPending bool
}
