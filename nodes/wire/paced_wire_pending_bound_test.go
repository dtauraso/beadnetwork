// paced_wire_pending_bound_test.go — pins PacedWire.pending's declared maximum
// (maxPendingEvents, paced_wire.go) and its fail-loud behavior at the bound
// (docs/planning/visual-editor/session-log.md Step 1). Per
// memory/feedback_check_the_signal_the_check_emits.md, the bound must be
// exceeded once deliberately to confirm it panics and names its own cause —
// this file does that, then confirms the normal drained path never trips it.
//
// Both tests exercise only this ONE PacedWire's own goroutine (the test
// goroutine itself, calling appendPending/DriveOneCycle synchronously) — no
// second goroutine, no cross-goroutine delivery/timing (docs/testing-shape.md).
package wire

import (
	"context"
	"strings"
	"testing"
)

// TestPendingBoundPanicsWhenDrainStops deliberately never drains pending (the
// same shape as the confirmed bug: a stream consumer wired but the per-cycle
// drain not running) and asserts appendPending panics exactly at
// maxPendingEvents+1, naming its own cause.
func TestPendingBoundPanicsWhenDrainStops(t *testing.T) {
	pw := NewPacedWire(1, 1.0)
	pw.StreamsActive = true

	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("appendPending did not panic after exceeding maxPendingEvents (%d); "+
				"pending grew unbounded instead", maxPendingEvents)
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("recovered panic value = %v (%T), want a string", r, r)
		}
		t.Logf("panic message (verbatim): %s", msg)
		if !strings.Contains(msg, "pending exceeded") || !strings.Contains(msg, "not running") {
			t.Fatalf("panic message %q does not name its own cause "+
				"(want it to mention \"pending exceeded\" and \"not running\")", msg)
		}
		if len(pw.pending) != maxPendingEvents+1 {
			t.Fatalf("pending len at panic = %d, want %d (bound + the one that tripped it)",
				len(pw.pending), maxPendingEvents+1)
		}
	}()

	for i := 0; i <= maxPendingEvents; i++ {
		pw.appendPending(pendingWireEvent{kind: "arrive", value: i})
	}
	t.Fatalf("unreachable: appendPending should have panicked by iteration %d", maxPendingEvents)
}

// TestPendingBoundNeverTripsWithNormalDrain drives many cycles with the
// production shape — DriveOneCycle then drainPendingEvents every cycle, the
// same back-to-back pairing edgeMover.run uses — and asserts it never panics
// and pending never approaches the bound, confirming the bound does not
// false-fire on ordinary traffic.
func TestPendingBoundNeverTripsWithNormalDrain(t *testing.T) {
	pw := NewPacedWire(1, 1.0)
	pw.StreamsActive = true
	ctx := context.Background()

	for i := range 50 {
		if got := pw.Send(i, beadPlacement{Steps: 1, Node: "src", Port: "out"}, 0); got != SendPlaced {
			t.Fatalf("Send(%d) = %v, want SendPlaced", i, got)
		}
	}

	for tick := int64(0); tick < 500; tick++ {
		pw.DriveOneCycle(ctx, tick)
		pw.drainPendingEvents()
		if len(pw.pending) > 0 {
			t.Fatalf("tick %d: pending = %d right after drain, want 0", tick, len(pw.pending))
		}
	}
}
