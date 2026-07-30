// paced_wire_inflight_bound_test.go — pins PacedWire.inflight's declared
// maximum (maxInflightBeads, paced_wire.go) and its fail-loud behavior at the
// bound (docs/planning/visual-editor/session-log.md Step 3, "inflight"). Per
// memory/feedback_check_the_signal_the_check_emits.md, the bound must be
// exceeded once deliberately to confirm it panics and names its own cause —
// this file does that, then confirms the normal (draining) path never trips
// it and that the backing slice is reset to nil once fully drained.
//
// Every test here exercises only this ONE PacedWire's own goroutine (the test
// goroutine itself, calling drainPlacements/DriveOneCycle synchronously) — no
// second goroutine, no cross-goroutine delivery/timing (docs/testing-shape.md).
package wire

import (
	"context"
	"strings"
	"testing"
)

// TestInflightBoundPanicsWhenDestinationStalls deliberately places beads onto
// inCh but never drives a DriveOneCycle with a draining destination — outCh is
// never read, so once beads reach their delivery deadline they cannot hand
// off and inflight can only grow. This asserts drainPlacements panics exactly
// at maxInflightBeads+1, naming its own cause.
func TestInflightBoundPanicsWhenDestinationStalls(t *testing.T) {
	pw := NewPacedWire(1, PulseSpeedWuPerMs)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("drainPlacements did not panic after exceeding maxInflightBeads (%d); "+
				"inflight grew unbounded instead", maxInflightBeads)
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("recovered panic value = %v (%T), want a string", r, r)
		}
		t.Logf("panic message (verbatim): %s", msg)
		// The message must name BOTH ways this bound is reached, not just one. A
		// stalled destination is the obvious cause, but a source placing faster than
		// the wire can carry reaches it too — naming only the first would send a
		// reader hunting a non-draining consumer that is working fine.
		for _, want := range []string{
			"inflight exceeded",
			"draining outCh",
			"placing faster",
		} {
			if !strings.Contains(msg, want) {
				t.Fatalf("panic message %q does not name its own cause (missing %q)", msg, want)
			}
		}
		if len(pw.inflight) != maxInflightBeads+1 {
			t.Fatalf("inflight len at panic = %d, want %d (bound + the one that tripped it)",
				len(pw.inflight), maxInflightBeads+1)
		}
	}()

	// Place one bead per call to drainPlacements at tick 0, never advancing the
	// clock and never reading outCh — nothing is ever delivered, so every
	// placement stays in inflight forever, exactly the "destination not
	// draining outCh" shape the bound exists to catch.
	for i := range maxInflightBeads + 1 {
		select {
		case pw.inCh <- placeRequest{val: i, bp: beadPlacement{InFlightMs: 1}, placementTick: 0}:
		default:
			t.Fatalf("inCh reported full before reaching maxInflightBeads (at %d); "+
				"wireChanBufferSize must be >= maxInflightBeads+1 for this test to exercise "+
				"the inflight bound rather than the inCh bound", i)
		}
		pw.drainPlacements()
	}
	t.Fatalf("unreachable: drainPlacements should have panicked by iteration %d", maxInflightBeads)
}

// TestInflightBoundNeverTripsWithNormalDelivery drives many cycles with the
// production shape (Send then DriveOneCycle every cycle, outCh drained every
// cycle by Recv) and asserts it never panics and inflight stays small,
// confirming the bound does not false-fire on ordinary traffic.
func TestInflightBoundNeverTripsWithNormalDelivery(t *testing.T) {
	pw := NewPacedWire(1, PulseSpeedWuPerMs)
	ctx := context.Background()

	for tick := int64(0); tick < 500; tick++ {
		if got := pw.Send(int(tick), beadPlacement{InFlightMs: 1}, tick); got != SendPlaced {
			t.Fatalf("tick %d: Send = %v, want SendPlaced", tick, got)
		}
		pw.DriveOneCycle(ctx, tick)
		for {
			if _, ok := pw.Recv(); !ok {
				break
			}
		}
		if len(pw.inflight) > maxInflightBeads {
			t.Fatalf("tick %d: inflight = %d, exceeds maxInflightBeads (%d) under normal delivery",
				tick, len(pw.inflight), maxInflightBeads)
		}
	}
}

// TestInflightResetsToNilWhenDrained places one bead, delivers it (draining
// outCh so the FIFO head can hand off), and asserts pw.inflight is reset to
// nil rather than merely re-sliced to length 0 — inflight[1:] never re-slices
// the backing array, so a bare re-slice would leave the array referenced
// (and still growable) forever even at zero length.
func TestInflightResetsToNilWhenDrained(t *testing.T) {
	pw := NewPacedWire(0, PulseSpeedWuPerMs)
	ctx := context.Background()

	if got := pw.Send(1, beadPlacement{InFlightMs: 0}, 1); got != SendPlaced {
		t.Fatalf("Send = %v, want SendPlaced", got)
	}
	// Send stamps placementTick from the tick passed here (1), and stepAll
	// (via DriveOneCycle) only advances a bead once nowTick > its own
	// placementTick — so placement and delivery need two distinct ticks.
	pw.DriveOneCycle(ctx, 1)
	pw.DriveOneCycle(ctx, 2)

	if _, ok := pw.Recv(); !ok {
		t.Fatalf("Recv() ok = false, want the bead placed above to have been delivered")
	}

	if pw.inflight != nil {
		t.Fatalf("pw.inflight = %#v after fully draining, want nil (not merely len 0), "+
			"so the backing array is released instead of retained and regrown", pw.inflight)
	}
}
