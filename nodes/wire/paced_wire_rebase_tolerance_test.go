// paced_wire_rebase_tolerance_test.go — the single-threaded replacement for the
// polling assertion node_move_test.go used to make across the concurrent edgeMover
// goroutine (via the now-deleted atomic snap/InFlightSegments).
//
// The "no skew with a single clock copy" test that used to lead this file was DELETED:
// it revised onto an arc of the same length, which makes ReviseInFlightGeometry the
// identity, so it could not fail. Its replacement, which varies the arc and therefore
// exercises the rebase math, is paced_wire_rebase_test.go.
package wire

import (
	"context"
	"math"
	"testing"
)

// approxEq is a local copy of nodes/Wiring's wire_test_helpers_test.go helper of the
// same name — trivial enough to duplicate rather than export test-only helpers across
// the package boundary.
func approxEq(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

// TestReviseInFlightGeometryRevisesInFlightSegment is the single-threaded
// replacement for the polling assertion node_move_test.go used to make across
// the concurrent edgeMover goroutine (via the now-deleted atomic snap/
// InFlightSegments). Everything here runs on the TEST goroutine only — no
// background driver — so pw.inflight is read directly, same-goroutine and
// race-free, exactly like TestReviseInFlightGeometryNoSkewWithSingleClockCopy
// above.
func TestReviseInFlightGeometryRevisesInFlightSegment(t *testing.T) {
	const crossTicks = 40 // == steps, at dwell 1.0 (see NewPacedWire below).
	pw := NewPacedWire(0, 1.0)
	steps := crossTicks

	ctx := context.Background()
	startSeg := WireSegment{Start: Vec3{}, End: Vec3{X: 1}}

	if pw.Send(0, beadPlacement{Steps: steps, Start: startSeg.Start, End: startSeg.End}, 0) != SendPlaced {
		t.Fatalf("Send failed")
	}
	// Drain the placement into pw.inflight (the wire's own per-cycle drive).
	pw.DriveOneCycle(ctx, 0)
	if len(pw.inflight) != 1 {
		t.Fatalf("expected 1 in-flight bead after DriveOneCycle, got %d", len(pw.inflight))
	}

	newSeg := WireSegment{Start: Vec3{X: 2}, End: Vec3{X: 3}}
	pw.ReviseInFlightGeometry(0, steps, newSeg)

	if len(pw.inflight) != 1 {
		t.Fatalf("expected 1 in-flight bead after revision, got %d", len(pw.inflight))
	}
	got := pw.inflight[0].seg
	if !approxEq(got.Start.X, newSeg.Start.X) || !approxEq(got.End.X, newSeg.End.X) {
		t.Fatalf("in-flight bead segment = %+v..%+v, want %+v..%+v", got.Start, got.End, newSeg.Start, newSeg.End)
	}
}
