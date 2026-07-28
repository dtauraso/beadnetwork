// paced_wire_streams_active_test.go — pins the fix for the confirmed
// unbounded-pending-growth bug (docs/planning/branch-notes/bounds-inventory.md):
// PacedWire.pending must NOT accumulate when no consumer is wired
// (StreamsActive false, the default), and MUST accumulate + drain when one is
// (StreamsActive true). This asserts only what this ONE PacedWire goroutine
// (the test goroutine itself, driving DriveOneCycle synchronously) itself
// decided/emitted — no second goroutine, no cross-goroutine delivery/timing
// (see docs/testing-shape.md and CLAUDE.md's "Testing shape" section).
package wire

import (
	"context"
	"testing"
)

// TestPendingStaysEmptyWithNoStreamConsumer inverts the confirmed repro: with
// StreamsActive left at its default (false — no per-edge fd wired), drive
// several deliveries end to end (Send, then DriveOneCycle every tick until
// delivered) and assert pending never grows. Before the fix, this same
// sequence left every delivery queued forever (streamOut nil -> 40
// deliveries left all 40 queued, per the brief's measured repro).
func TestPendingStaysEmptyWithNoStreamConsumer(t *testing.T) {
	pw := NewPacedWire(1, PulseSpeedWuPerMs)
	ctx := context.Background()

	for i := range 40 {
		if got := pw.Send(i, beadPlacement{InFlightMs: 1, Node: "src", Port: "out"}); got != SendPlaced {
			t.Fatalf("Send(%d) = %v, want SendPlaced", i, got)
		}
	}

	// Drive enough cycles for every bead to fully cross and be delivered.
	for tick := int64(0); tick < 200; tick++ {
		pw.DriveOneCycle(ctx, tick)
	}

	if len(pw.inflight) != 0 {
		t.Fatalf("inflight = %d after driving to completion, want 0 (all delivered)", len(pw.inflight))
	}
	if len(pw.pending) != 0 {
		t.Fatalf("pending = %d with no stream consumer wired (StreamsActive=false), want 0 "+
			"— pending must never accumulate when nothing will ever call DrainPendingEvents "+
			"(the confirmed unbounded-growth bug)", len(pw.pending))
	}
}

// TestPendingAccumulatesAndDrainsWithStreamConsumer is the control case: with
// StreamsActive explicitly set (mirroring stream_wiring.go's setEdgeStreams,
// which sets it before this wire's mover goroutine ever launches), the same
// drive sequence DOES produce pending events, and DrainPendingEvents clears
// them — confirming the gate only suppresses accumulation, not delivery
// tracing, once a real consumer exists.
func TestPendingAccumulatesAndDrainsWithStreamConsumer(t *testing.T) {
	pw := NewPacedWire(1, PulseSpeedWuPerMs)
	pw.StreamsActive = true
	ctx := context.Background()

	if got := pw.Send(1, beadPlacement{InFlightMs: 1, Node: "src", Port: "out"}); got != SendPlaced {
		t.Fatalf("Send = %v, want SendPlaced", got)
	}

	sawPending := false
	for tick := int64(0); tick < 200; tick++ {
		pw.DriveOneCycle(ctx, tick)
		if len(pw.pending) > 0 {
			sawPending = true
			break
		}
	}
	if !sawPending {
		t.Fatalf("pending never accumulated any event with StreamsActive=true; want at least one")
	}

	if drained := pw.drainPendingEvents(); len(drained) == 0 {
		t.Fatalf("drainPendingEvents() returned 0 events, want at least 1")
	}
	if len(pw.pending) != 0 {
		t.Fatalf("pending = %d after drainPendingEvents, want 0 (drain clears the buffer)", len(pw.pending))
	}
}
