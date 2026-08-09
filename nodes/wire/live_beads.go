// live_beads.go — the STATELESS PROJECTIONS of a PacedWire's in-flight queue.
//
// Nothing here holds state of its own: each is a pure read of pw.inflight at a given tick,
// computed with the same lerp math advanceBead uses but with no side effects. They are NOT
// part of the wire's readout (wire_readout.go) and hold no copy of the queue — they stay
// methods on PacedWire, under the wire's usual single-goroutine contract: call them only
// from the goroutine that drives this wire, which in production is its SOURCE NODE's own
// mover. This file is a split for readability; the code below is unchanged.
package wire

// LiveBeadRow is one in-flight bead's CURRENT world position + value + id, computed with
// the same lerp math advanceBead uses but with NO side effects — no trace emit, no state
// mutation. Used only by the dedicated per-edge stream (edgeMover.writeStreamFrame,
// node_mover.go) to snapshot this wire's current beads without duplicating tr.Position's
// separate accumulation into a central buffer.
type LiveBeadRow struct {
	Val     int
	X, Y, Z float64
	Gen     uint64
}

// LiveBeadProgress is one in-flight bead's fractional progress and its VALUE. The value is
// what the chain colours with: a lit chain bead takes bead 0's or bead 1's own fill, so
// "which bead is lit" is not enough information on its own.
type LiveBeadProgress struct {
	T   float64 // fractional progress 0..1
	Val int     // bead value (0|1)
	// Steps is the bead's OWN step count — the geometry its t was computed
	// against (ticksToCross = steps*dwell). The caller recovers WHICH BEAD is
	// lit as floor(t*Steps) — no length multiplication anywhere
	// (docs/bead-lattice.md "Timing"): layout laid the chain out on this same
	// integer, so lighting and layout read the same N and cannot disagree.
	Steps int
}

// LiveBeadFractions returns the FRACTIONAL progress t (0..1) and VALUE of every in-flight
// bead on this wire at tick, in FIFO order — the same t advanceBead computes for the moving
// bead's position, exposed as the scalar it always was.
//
// This is what the chain-bead animation needs and ALL it needs (docs/beads-are-the-edge.md):
// a chain is a fixed sequence, so "where has this traversal got to" is one number per bead
// in flight, not a recomputed world position. The lit bead is index = t × count.
//
// Same single-goroutine contract as LiveBeadRows: safe only from the goroutine that drives
// this wire. That goroutine is now the SOURCE NODE's own mover (nodeMover.run drives its
// outgoing wires), which is exactly why the node can light its own chain without reading
// another goroutine's state.
func (pw *PacedWire) LiveBeadFractions(tick int64) []LiveBeadProgress {
	nowTick := float64(tick)
	out := make([]LiveBeadProgress, 0, len(pw.inflight))
	for i := range pw.inflight {
		b := &pw.inflight[i]
		crossTicks := pw.ticksToCross(b.steps)
		if crossTicks <= 0 {
			continue
		}
		target := nowTick
		if nowTick >= b.placementTick+crossTicks {
			target = b.placementTick + crossTicks
		}
		t := (target - b.placementTick) / crossTicks
		if t > 1 {
			t = 1
		}
		if t < 0 {
			t = 0
		}
		out = append(out, LiveBeadProgress{T: t, Val: b.val, Steps: b.steps})
	}
	return out
}

// LiveBeadRows returns every in-flight, position-streaming bead's CURRENT world position
// at tick (this wire's own goroutine's clock reading), in FIFO order. Safe to call ONLY
// from this wire's own goroutine (reads pw.inflight directly — same single-
// goroutine-ownership contract stepAll/ReviseInFlightGeometry rely on). A bead with no
// position stream (bp.streams()==false) is omitted, matching advanceBead's own emit gate.
func (pw *PacedWire) LiveBeadRows(tick int64) []LiveBeadRow {
	nowTick := float64(tick)
	rows := make([]LiveBeadRow, 0, len(pw.inflight))
	for i := range pw.inflight {
		b := &pw.inflight[i]
		if !b.streams {
			continue
		}
		crossTicks := pw.ticksToCross(b.steps)
		target := nowTick
		if crossTicks > 0 && nowTick >= b.placementTick+crossTicks {
			target = b.placementTick + crossTicks
		}
		t := 0.0
		if crossTicks > 0 {
			t = (target - b.placementTick) / crossTicks
		}
		if t < 0 {
			t = 0
		}
		if t > 1 {
			t = 1
		}
		p := lerp(b.seg.Start, b.seg.End, t)
		rows = append(rows, LiveBeadRow{Val: b.val, X: p.X, Y: p.Y, Z: p.Z, Gen: b.gen})
	}
	return rows
}
