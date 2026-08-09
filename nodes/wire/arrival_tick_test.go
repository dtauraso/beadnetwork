package wire

import "testing"

// arrival_tick_test.go — the arrival tick is KNOWN AT PLACEMENT, and across several wires the
// SOONEST one wins. Both are what let a source node sleep to the moment a traversal completes
// rather than waking every cycle to ask whether it has.
//
// These assert one goroutine's own arithmetic over its own in-flight beads — no delivery, no
// second goroutine, nothing about two goroutines communicating (docs/testing-shape.md).

// place puts a bead in flight directly, the way the wire's own goroutine would after draining
// its inCh — the drain path itself is not what these tests are about.
func place(pw *PacedWire, placementTick float64, steps int) {
	pw.inflight = append(pw.inflight, inflightBead{placementTick: placementTick, steps: steps})
}

func TestNextArrivalTickIsPlacementPlusCrossing(t *testing.T) {
	pw := NewPacedWire(0, 2.0) // dwell 2 ticks per bead step
	if _, ok := pw.NextArrivalTick(); ok {
		t.Fatal("empty wire reports an arrival")
	}
	place(pw, 10, 3) // lands at 10 + 3*2 = 16
	got, ok := pw.NextArrivalTick()
	if !ok || got != 16 {
		t.Fatalf("NextArrivalTick = %d, %v; want 16, true", got, ok)
	}
}

func TestNextArrivalTickCeilsToAWholeTick(t *testing.T) {
	// Delivery happens on the first INTEGER tick at or past the deadline, so a fractional
	// deadline must round UP — waking a fraction early would find nothing delivered and put
	// the poll back that this exists to remove.
	pw := NewPacedWire(0, 0.5)
	place(pw, 10, 3) // 10 + 1.5 = 11.5
	if got, _ := pw.NextArrivalTick(); got != 12 {
		t.Fatalf("NextArrivalTick = %d, want 12 (ceil of 11.5)", got)
	}
}

func TestNextArrivalTickIsTheEarliestInFlight(t *testing.T) {
	pw := NewPacedWire(0, 1.0)
	place(pw, 0, 40) // lands at 40
	place(pw, 5, 10) // lands at 15 — placed later, arrives sooner
	if got, _ := pw.NextArrivalTick(); got != 15 {
		t.Fatalf("NextArrivalTick = %d, want 15 (the earliest, not the first placed)", got)
	}
}

func TestEarliestArrivalAcrossWiresTakesTheShorterEdge(t *testing.T) {
	// A node with n outgoing edges sleeps to the SOONEST arrival among them: the shorter
	// edge is the first thing needing it awake, and the longer one has simply not arrived.
	long := NewPacedWire(0, 1.0)
	short := NewPacedWire(0, 1.0)
	place(long, 0, 30)  // lands at 30
	place(short, 0, 12) // lands at 12
	got, ok := EarliestArrival([]*PacedWire{long, short})
	if !ok || got != 12 {
		t.Fatalf("EarliestArrival = %d, %v; want 12, true", got, ok)
	}
}

func TestEarliestArrivalIgnoresIdleAndNilWires(t *testing.T) {
	idle := NewPacedWire(0, 1.0)
	busy := NewPacedWire(0, 1.0)
	place(busy, 4, 2) // lands at 6
	got, ok := EarliestArrival([]*PacedWire{nil, idle, busy})
	if !ok || got != 6 {
		t.Fatalf("EarliestArrival = %d, %v; want 6, true", got, ok)
	}
	if _, ok := EarliestArrival([]*PacedWire{nil, idle}); ok {
		t.Fatal("all-idle wires report an arrival — a node with nothing in flight has no traversal to wake for")
	}
}
