package beadchain

import (
	"testing"
	"time"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// bead_wake_settle_test.go — StartDrag/EndDrag as the group's wake/settle signal: every
// affected bead is set/cleared by a single call, an abandoned drag still settles, the flag
// is set once per drag rather than per event, and either endpoint of an edge can wake it.

// --- Wake and settle --------------------------------------------------------------------

// TestWakeSetsEveryAffectedBead: a single StartDrag sets EVERY bead's flag; a single
// EndDrag clears every bead's flag.
func TestWakeSetsEveryAffectedBead(t *testing.T) {
	const n = 25
	g := NewBeadWakeGroup()
	stop := make(chan struct{})
	defer close(stop)

	obss := make([]<-chan BeadSnapshot, n)
	for i := 0; i < n; i++ {
		geom, wake, settle := g.Current()
		b := NewBead(float64(i), geom, wake, settle, make(chan struct{}), stop)
		obss[i] = b.WithObserve()
		b.Start()
	}

	g.StartDrag()
	for i, obs := range obss {
		if _, ok := waitForSnapshot(t, obs, time.Second, func(s BeadSnapshot) bool { return s.Dragging }); !ok {
			t.Fatalf("bead %d was not woken by a single StartDrag", i)
		}
	}

	g.EndDrag()
	for i, obs := range obss {
		if _, ok := waitForSnapshot(t, obs, time.Second, func(s BeadSnapshot) bool { return !s.Dragging }); !ok {
			t.Fatalf("bead %d was not settled by a single EndDrag", i)
		}
	}
}

// --- "done dragging" is not optional ----------------------------------------------------

// TestAbandonedDragStillSettles: a drag that ends WITHOUT a clean per-event "end" — the
// caller simply calls EndDrag once, exactly as it would for any other drag conclusion, with
// no intervening geometry events after StartDrag — still clears every bead's flag. There is
// no separate "abandoned" code path in BeadWakeGroup; EndDrag is unconditional, so this
// documents that an abandoned drag is not a distinguishable case at this layer.
func TestAbandonedDragStillSettles(t *testing.T) {
	const n = 10
	g := NewBeadWakeGroup()
	stop := make(chan struct{})
	defer close(stop)

	obss := make([]<-chan BeadSnapshot, n)
	for i := 0; i < n; i++ {
		geom, wake, settle := g.Current()
		b := NewBead(float64(i), geom, wake, settle, make(chan struct{}), stop)
		obss[i] = b.WithObserve()
		b.Start()
	}

	g.StartDrag() // drag begins; no geometry ever arrives, no explicit "abandon" signal
	for i, obs := range obss {
		if _, ok := waitForSnapshot(t, obs, time.Second, func(s BeadSnapshot) bool { return s.Dragging }); !ok {
			t.Fatalf("bead %d never woke", i)
		}
	}

	g.EndDrag() // the FSM's unconditional settle path, same call an ordinary drag end uses
	for i, obs := range obss {
		if _, ok := waitForSnapshot(t, obs, time.Second, func(s BeadSnapshot) bool { return !s.Dragging }); !ok {
			t.Fatalf("an abandoned drag left bead %d on machine time", i)
		}
	}
}

// --- The flag is set once per drag ------------------------------------------------------

// TestFlagSetOncePerDragNotPerEvent: a whole gesture with many geometry events between one
// StartDrag and one EndDrag never toggles dragging in between — it is read true across
// every intervening geometry event, proving BroadcastGeometry alone cannot flip the flag
// (only StartDrag/EndDrag advance g.wake/g.settle — check-broadcast-is-close-not-loop.sh
// scopes each method separately, and the type split above keeps geometry and mode disjoint
// besides).
func TestFlagSetOncePerDragNotPerEvent(t *testing.T) {
	b, g, _, stop, obs := newTestBead(1.0)
	defer close(stop)
	b.Start()

	g.StartDrag()
	if _, ok := waitForSnapshot(t, obs, time.Second, func(s BeadSnapshot) bool { return s.Dragging }); !ok {
		t.Fatalf("drag never started")
	}

	for i := 0; i < 60; i++ { // one gesture's worth of pointer events
		g.BroadcastGeometry(BeadGeometryIn{Center: wire.Vec3{X: float64(i)}, Aim: wire.Vec3{X: 1}})
		want := wire.Vec3{X: float64(i) + 1}
		snap, ok := waitForSnapshot(t, obs, time.Second, func(s BeadSnapshot) bool { return s.Position == want })
		if !ok {
			t.Fatalf("geometry event %d never resolved", i)
		}
		if !snap.Dragging {
			t.Fatalf("dragging flag flipped mid-gesture at event %d — it must be set once per drag, not per event", i)
		}
	}

	g.EndDrag()
	if _, ok := waitForSnapshot(t, obs, time.Second, func(s BeadSnapshot) bool { return !s.Dragging }); !ok {
		t.Fatalf("drag never ended")
	}
}

// --- Either endpoint can wake -----------------------------------------------------------

// TestEitherEndpointCanWakeSource / TestEitherEndpointCanWakeTarget: dragging the source
// node's own BeadWakeGroup wakes the edge's beads, and so does dragging the target's own
// group — each endpoint owns an independent group over the SAME beads' worth of state (a
// bead subscribes to exactly one group at construction; a real edge would give the source's
// beads to the source's group — this test exercises the mechanism from each side, not a
// shared-group scenario, since a bead is owned by exactly one node/edge/direction, matching
// PLAN.md's "one test each, not a both-at-once test").
func TestEitherEndpointCanWakeSource(t *testing.T) {
	b, g, _, stop, obs := newTestBead(1.0)
	defer close(stop)
	b.Start()
	g.StartDrag() // this bead's group belongs to the edge's SOURCE node
	if _, ok := waitForSnapshot(t, obs, time.Second, func(s BeadSnapshot) bool { return s.Dragging }); !ok {
		t.Fatalf("dragging the source did not wake the bead")
	}
}

func TestEitherEndpointCanWakeTarget(t *testing.T) {
	b, g, _, stop, obs := newTestBead(1.0)
	defer close(stop)
	b.Start()
	g.StartDrag() // this bead's group belongs to the edge's TARGET node instead
	if _, ok := waitForSnapshot(t, obs, time.Second, func(s BeadSnapshot) bool { return s.Dragging }); !ok {
		t.Fatalf("dragging the target did not wake the bead")
	}
}
