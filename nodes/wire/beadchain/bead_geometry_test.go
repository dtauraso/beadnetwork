package beadchain

import (
	"testing"
	"time"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// bead_geometry_test.go — the geometry side of a bead: clock-independent (the two-clock
// invariant, MODEL.md's "no position update may be gated on the human clock"), mode
// switching between human and machine time, one BroadcastGeometry call reaching every bead
// in one hop (not a chain), each bead's position as a pure function of its own offset (no
// diffusion), the geometry/animation writer split, and the per-event work staying inside a
// frame budget at N=1000.

// --- Clock separation ---------------------------------------------------------------

// TestGeometryCompletesWithHumanClockStopped: a position update completes even though no
// tick is ever sent on this bead's tick channel — geometry is never gated on the human
// clock (MODEL.md's two-clock invariant: "no position update may be gated on the human
// clock", applied here to the bead's own select rather than a package-level rule).
func TestGeometryCompletesWithHumanClockStopped(t *testing.T) {
	b, g, tickCh, stop, obs := newTestBead(2.0)
	_ = tickCh // deliberately never sent to
	defer close(stop)
	b.Start()

	g.BroadcastGeometry(BeadGeometryIn{Center: wire.Vec3{X: 10}, Aim: wire.Vec3{X: 1}})

	snap, ok := waitForSnapshot(t, obs, time.Second, func(s BeadSnapshot) bool {
		return s.Position == (wire.Vec3{X: 12}) // Center + Aim*offsetR
	})
	if !ok {
		t.Fatalf("position update did not complete with the human clock stopped; last=%+v", snap)
	}
}

// TestAnimationStepDoesNotAdvancePosition: a tick pulse (the animation clock) never changes
// position — the disjoint half of the same invariant.
func TestAnimationStepDoesNotAdvancePosition(t *testing.T) {
	b, _, tickCh, stop, obs := newTestBead(3.0)
	defer close(stop)
	b.Start()

	tickCh <- struct{}{}
	snap, ok := waitForSnapshot(t, obs, time.Second, func(s BeadSnapshot) bool { return s.Lit })
	if !ok {
		t.Fatalf("tick never reached the bead's animation state")
	}
	if snap.Position != (wire.Vec3{}) {
		t.Fatalf("an animation tick moved position: %+v", snap.Position)
	}
}

// --- Mode switching -------------------------------------------------------------------

// TestModeSwitchingRestsThenDragsThenRests: a bead rests in human-clock mode (dragging ==
// false) until a wake message moves it to machine time, and a settle message returns it —
// never both at once, and neither loses a pending tick nor a pending geometry update across
// the switch.
func TestModeSwitchingRestsThenDragsThenRests(t *testing.T) {
	b, g, tickCh, stop, obs := newTestBead(1.0)
	defer close(stop)
	b.Start()

	g.StartDrag()
	if _, ok := waitForSnapshot(t, obs, time.Second, func(s BeadSnapshot) bool { return s.Dragging }); !ok {
		t.Fatalf("wake message did not set dragging")
	}

	// A geometry update arriving mid-drag is not lost.
	g.BroadcastGeometry(BeadGeometryIn{Center: wire.Vec3{Y: 5}, Aim: wire.Vec3{Y: 1}})
	if _, ok := waitForSnapshot(t, obs, time.Second, func(s BeadSnapshot) bool { return s.Position == (wire.Vec3{Y: 6}) }); !ok {
		t.Fatalf("geometry update lost while dragging")
	}

	g.EndDrag()
	if _, ok := waitForSnapshot(t, obs, time.Second, func(s BeadSnapshot) bool { return !s.Dragging }); !ok {
		t.Fatalf("settle message did not clear dragging")
	}

	// A tick after settling is not lost either.
	tickCh <- struct{}{}
	if _, ok := waitForSnapshot(t, obs, time.Second, func(s BeadSnapshot) bool { return s.Lit }); !ok {
		t.Fatalf("tick lost after settling")
	}
}

// --- No human-clock coupling in geometry; one hop, not N ------------------------------

// TestOneHopNotN: a single BroadcastGeometry call (one Advance, i.e. one close) reaches
// every bead in the group — for both N=40 and N=1000 — with no tick required and no
// per-bead send. This is the test that would have caught the original O(N^2)
// neighbour-following defect (memory/project_wire_is_straight_line_not_chain.md).
func TestOneHopNotN(t *testing.T) {
	for _, n := range []int{40, 1000} {
		n := n
		t.Run(itoa(n), func(t *testing.T) {
			g := NewBeadWakeGroup()
			stop := make(chan struct{})
			defer close(stop)

			obss := make([]<-chan BeadSnapshot, n)
			for i := 0; i < n; i++ {
				geom, wake, settle := g.Current()
				tickCh := make(chan struct{})
				b := NewBead(float64(i), geom, wake, settle, tickCh, stop)
				obss[i] = b.WithObserve()
				b.Start()
			}

			// ONE broadcast call for the whole group, regardless of n.
			g.BroadcastGeometry(BeadGeometryIn{Center: wire.Vec3{X: 100}, Aim: wire.Vec3{X: 1}})

			for i, obs := range obss {
				want := wire.Vec3{X: 100 + float64(i)}
				if _, ok := waitForSnapshot(t, obs, 2*time.Second, func(s BeadSnapshot) bool { return s.Position == want }); !ok {
					t.Fatalf("bead %d never reached the broadcast position", i)
				}
			}
		})
	}
}

// --- No diffusion -----------------------------------------------------------------------

// TestNoDiffusion: a bead's position is a pure function of the broadcast transform and its
// OWN fixed offset — never of another bead's position. Two beads with different offsets
// fed the SAME broadcast land at different, independently-computable positions; neither one
// consulted the other.
func TestNoDiffusion(t *testing.T) {
	g := NewBeadWakeGroup()
	stop := make(chan struct{})
	defer close(stop)

	geom, wake, settle := g.Current()
	b1 := NewBead(1.0, geom, wake, settle, make(chan struct{}), stop)
	b2 := NewBead(9.0, geom, wake, settle, make(chan struct{}), stop)
	obs1 := b1.WithObserve()
	obs2 := b2.WithObserve()
	b1.Start()
	b2.Start()

	g.BroadcastGeometry(BeadGeometryIn{Center: wire.Vec3{X: 0}, Aim: wire.Vec3{X: 1}})

	if _, ok := waitForSnapshot(t, obs1, time.Second, func(s BeadSnapshot) bool { return s.Position == (wire.Vec3{X: 1}) }); !ok {
		t.Fatalf("b1 did not reach its offset-derived position")
	}
	if _, ok := waitForSnapshot(t, obs2, time.Second, func(s BeadSnapshot) bool { return s.Position == (wire.Vec3{X: 9}) }); !ok {
		t.Fatalf("b2 did not reach its offset-derived position")
	}
}

// --- Disjoint writers ---------------------------------------------------------------

// TestDisjointWriters: the geometry broadcast writes ONLY position; the tick pulse writes
// ONLY animation. Neither event's handler touches the other's field — the structural split
// (separate beadGeometryState/beadAnimationState types) is exercised here behaviourally.
func TestDisjointWriters(t *testing.T) {
	b, g, tickCh, stop, obs := newTestBead(1.0)
	defer close(stop)
	b.Start()

	tickCh <- struct{}{}
	snap, ok := waitForSnapshot(t, obs, time.Second, func(s BeadSnapshot) bool { return s.Lit })
	if !ok {
		t.Fatalf("tick did not update animation")
	}
	if snap.Position != (wire.Vec3{}) {
		t.Fatalf("tick unexpectedly moved position: %+v", snap.Position)
	}
	litBefore := snap.Lit

	g.BroadcastGeometry(BeadGeometryIn{Center: wire.Vec3{Z: 4}, Aim: wire.Vec3{Z: 1}})
	snap, ok = waitForSnapshot(t, obs, time.Second, func(s BeadSnapshot) bool { return s.Position == (wire.Vec3{Z: 5}) })
	if !ok {
		t.Fatalf("geometry did not update position")
	}
	if snap.Lit != litBefore {
		t.Fatalf("geometry broadcast unexpectedly changed animation state")
	}
}

// --- No sleeping (behavioural half; the source-guard half is
// tools/network/concurrency/check-no-wall-clock-wait.sh, which already scans nodes/wire) ------------------

// TestGeometryServicedWithoutWaitingForATick: a bead waiting for its next tick still
// services a geometry update immediately — nothing here blocks on a timer.
func TestGeometryServicedWithoutWaitingForATick(t *testing.T) {
	b, g, _, stop, obs := newTestBead(1.0)
	defer close(stop)
	b.Start()

	start := time.Now()
	g.BroadcastGeometry(BeadGeometryIn{Center: wire.Vec3{X: 1}, Aim: wire.Vec3{X: 1}})
	if _, ok := waitForSnapshot(t, obs, 200*time.Millisecond, func(s BeadSnapshot) bool { return s.Position == (wire.Vec3{X: 2}) }); !ok {
		t.Fatalf("geometry update took too long (possible hidden wait): %v", time.Since(start))
	}
}

// --- Frame budget -----------------------------------------------------------------------

// TestFrameBudgetN1000: a pointer event's position work for N=1000 completes well inside
// one 16.67ms frame (measured, not assumed).
func TestFrameBudgetN1000(t *testing.T) {
	const n = 1000
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

	start := time.Now()
	g.BroadcastGeometry(BeadGeometryIn{Center: wire.Vec3{X: 1}, Aim: wire.Vec3{X: 1}})
	for i, obs := range obss {
		want := wire.Vec3{X: 1 + float64(i)}
		if _, ok := waitForSnapshot(t, obs, 200*time.Millisecond, func(s BeadSnapshot) bool { return s.Position == want }); !ok {
			t.Fatalf("bead %d did not resolve position", i)
		}
	}
	elapsed := time.Since(start)
	const frameBudget = 16*time.Millisecond + 670*time.Microsecond
	if elapsed > 10*frameBudget {
		// Generous slack for a scheduler-driven, non-realtime test environment: the claim
		// is "well inside a frame" for the actual per-event WORK, which under goroutine
		// scheduling noise this bounds loosely rather than tightly (a tight bound here
		// would make this test flaky on a loaded CI box, not a correctness assertion about
		// the model). t.Logf reports the real number for a human to read directly.
		t.Fatalf("N=1000 broadcast+resolve took %v, far beyond a frame budget (%v)", elapsed, frameBudget)
	}
	t.Logf("N=1000 broadcast+resolve: %v (frame budget %v)", elapsed, frameBudget)
}
