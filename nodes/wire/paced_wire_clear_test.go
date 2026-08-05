// paced_wire_clear_test.go — pins PacedWire.ClearInFlight: the wire is left as empty as a
// freshly built one, and nothing it was carrying is delivered afterwards.
//
// Every test here exercises only this ONE PacedWire's own goroutine (the test goroutine
// itself, calling Send/DriveOneCycle/ClearInFlight synchronously) — no second goroutine,
// no cross-goroutine delivery or timing (docs/testing-shape.md).
package wire

import (
	"context"
	"testing"
)

// A cleared wire delivers nothing. Both populations have to go: beads already crossing
// (drained onto inflight by an earlier drive) AND beads placed but not yet drained off
// inCh — a clear that took only the first would let the queued ones cross a moment later,
// which is exactly the "the reset didn't take" symptom this exists to prevent.
func TestClearInFlightDeliversNothingAfterwards(t *testing.T) {
	ctx := context.Background()
	pw := NewPacedWire(4, 1.0)

	pw.Send(1, beadPlacement{Steps: 4}, 0)
	// Two drives: the first drains this bead onto inflight, the second steps it. Four
	// steps at dwell 1 means it is still crossing, not yet delivered.
	pw.DriveOneCycle(ctx, 1)
	pw.DriveOneCycle(ctx, 2)
	if len(pw.inflight) != 1 {
		t.Fatalf("setup: want 1 bead still crossing before the clear, got %d", len(pw.inflight))
	}
	pw.Send(1, beadPlacement{Steps: 4}, 2)
	// ...and this second one is still queued on inCh, undrained.

	pw.ClearInFlight()

	// Drive well past every deadline: had either population survived, it would deliver here.
	for tick := int64(3); tick < 20; tick++ {
		pw.DriveOneCycle(ctx, tick)
		if _, ok := pw.Recv(); ok {
			t.Fatalf("a bead was delivered at tick %d after ClearInFlight; the wire was not emptied", tick)
		}
	}
	if len(pw.inflight) != 0 {
		t.Fatalf("inflight = %d beads after ClearInFlight, want 0", len(pw.inflight))
	}
}

// ClearInFlight does NOT touch what has already been handed off to outCh. Those beads
// belong to the DESTINATION node now — it drains its own In on its own goroutine — so a
// source clearing its wire must not reach across that ownership line and swallow them.
func TestClearInFlightLeavesAlreadyDeliveredBeads(t *testing.T) {
	ctx := context.Background()
	pw := NewPacedWire(1, 1.0)

	pw.Send(7, beadPlacement{Steps: 1}, 0)
	// Drive past its one-step deadline so it is handed off to outCh — i.e. it now belongs
	// to the destination, not to this wire.
	for tick := int64(1); tick < 6; tick++ {
		pw.DriveOneCycle(ctx, tick)
	}
	if len(pw.inflight) != 0 {
		t.Fatal("setup: the bead should have been delivered onto outCh before the clear")
	}

	pw.ClearInFlight()

	v, ok := pw.Recv()
	if !ok {
		t.Fatal("ClearInFlight swallowed a bead already delivered to the destination's end")
	}
	if v != 7 {
		t.Fatalf("delivered value = %d, want 7", v)
	}
}
