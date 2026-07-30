// paced_wire_rebase_test.go — what ReviseInFlightGeometry must preserve when the edge
// under a bead CHANGES LENGTH. That is the live case: dragging a node makes its edges
// longer or shorter while beads are already crossing them, and a bead 25% of the way
// across must still be 25% of the way across afterwards — it must not jump forward or
// snap back because the arc it was measured against moved.
//
// This file replaces paced_wire_rebase_tolerance_test.go's "no skew" test, which was
// discovered to be a TAUTOLOGY: it revised onto an arc of the SAME length as the old
// one, and in that case placementTick' = now − t·(arc/pulseSpeed) reduces algebraically
// to placementTick — the identity, for any tick argument. It passed with the tolerance
// tightened to 1e-30, and passed with the rebase deliberately fed nowTick+1. A test that
// cannot fail reads exactly like a passing one
// (memory/feedback_check_the_signal_the_check_emits.md).
//
// Changing the arc is what makes the rebase math non-trivial and the assertion able to
// fail. No clock is involved: ticks are passed as literals, so there is nothing timing-
// dependent left here and no bubble is needed.
package wire

import (
	"context"
	"testing"
)

// TestReviseInFlightGeometryPreservesFractionAcrossArcChange: a bead 25% across a
// 40-tick edge stays 25% across after the edge doubles to 80 ticks.
func TestReviseInFlightGeometryPreservesFractionAcrossArcChange(t *testing.T) {
	const oldCrossTicks = 40.0
	const newCrossTicks = 80.0 // the edge DOUBLES — this is what the old test never did.
	const placedAt = 0
	const nowTick = 10 // 10 of 40 ticks elapsed ⇒ the bead is 25% across.
	const wantFraction = 0.25

	pw := NewPacedWire(0, PulseSpeedWuPerTick)
	newArc := newCrossTicks * PulseSpeedWuPerTick

	// Place via the production path (Send + DriveOneCycle), stamped at tick 0.
	if pw.Send(0, beadPlacement{InFlightMs: oldCrossTicks * MsPerTick, Start: Vec3{}, End: Vec3{X: 1}}, placedAt) != SendPlaced {
		t.Fatalf("Send failed")
	}
	pw.DriveOneCycle(context.Background(), placedAt)
	if len(pw.inflight) != 1 {
		t.Fatalf("expected 1 in-flight bead after DriveOneCycle, got %d", len(pw.inflight))
	}

	// Everything runs on the TEST goroutine — no background driver — so reading
	// pw.inflight directly is same-goroutine and race-free.
	pw.ReviseInFlightGeometry(nowTick, newArc, WireSegment{Start: Vec3{}, End: Vec3{X: 2}})

	if len(pw.inflight) != 1 {
		t.Fatalf("expected 1 in-flight bead after revision, got %d", len(pw.inflight))
	}
	b := pw.inflight[0]
	if b.arc != newArc {
		t.Fatalf("arc not updated: got %v, want %v", b.arc, newArc)
	}

	// The fraction is now measured against the NEW crossing time. Preserving 25% over
	// 80 ticks means 20 ticks of it are already covered, so placementTick must be
	// rebased to 10 − 20 = −10. Leaving placementTick alone (the old test's blind spot)
	// would read as 10/80 = 12.5% — the bead visibly snapping backwards.
	gotFraction := (nowTick - b.placementTick) / newCrossTicks
	if !approxEq(gotFraction, wantFraction) {
		t.Fatalf("fraction not preserved across arc change: got %.6f, want %.6f (placementTick=%.1f, want -10)",
			gotFraction, wantFraction, b.placementTick)
	}
	if !approxEq(b.placementTick, -10) {
		t.Fatalf("placementTick not rebased onto the new arc: got %.6f, want -10", b.placementTick)
	}
}
