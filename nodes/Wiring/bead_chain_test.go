package Wiring

import (
	"bytes"
	"runtime"
	"runtime/pprof"
	"strings"
	"testing"
	"time"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// bead_chain_test.go — the production integration's own tests, distinct from
// nodes/wire's primitive-level tests (bead_actor_test.go). These exercise
// nodeMover.reconcileBeadChain/startBeadDrag/endBeadDrag directly, which is the real
// production call site chainBeads uses when m.beadTickFn is set (see beadTickFn's own
// doc comment on why chainBeads itself stays untouched/synchronous when it is nil).

// beadRunGoroutineCount counts live goroutines currently inside wire.(*Bead).run, via the
// same goroutine-profile technique bead_actor_test.go's TestIdleBeadIsBlockedNotRunnable
// already uses — the sanctioned way to observe another goroutine's existence without
// touching its owned state.
func beadRunGoroutineCount(t *testing.T) int {
	t.Helper()
	var buf bytes.Buffer
	if err := pprof.Lookup("goroutine").WriteTo(&buf, 2); err != nil {
		t.Fatalf("goroutine profile: %v", err)
	}
	dump := buf.String()
	count := 0
	for _, sec := range strings.Split(dump, "\n\n") {
		if strings.Contains(sec, "wire.(*Bead).run") {
			count++
		}
	}
	return count
}

// TestBeadGoroutineLifetimeFollowsChainLength: growing a chain via reconcileBeadChain
// starts one goroutine per added bead; shrinking it back down closes each removed bead's
// own stop channel, and every removed bead's goroutine actually exits — no leak per
// removed bead. This is the test CLAUDE.md's task called for explicitly: goroutine count
// returns to baseline after a drag that removes beads (bead_crud.go's own count, recomputed
// live by chain_beads.go's edgeStepCount, shrinking as two nodes move together).
func TestBeadGoroutineLifetimeFollowsChainLength(t *testing.T) {
	m := &nodeGeometry{id: "a", beadTickFn: wire.NewTickChan}
	offsetAt := func(i int) float64 { return float64(i) }

	baseline := beadRunGoroutineCount(t)

	m.reconcileBeadChain("b", 10, offsetAt, wire.Vec3{X: 1})
	runtime.Gosched()
	time.Sleep(20 * time.Millisecond)
	grown := beadRunGoroutineCount(t)
	if grown != baseline+10 {
		t.Fatalf("after growing to 10 beads: got %d Bead.run goroutines (baseline %d), want %d", grown, baseline, baseline+10)
	}

	m.reconcileBeadChain("b", 3, offsetAt, wire.Vec3{X: 1})
	// Closing a bead's stop channel wakes it out of its select and it returns
	// immediately — give the scheduler a moment to actually retire those goroutines
	// before re-counting (same shape as TestIdleBeadIsBlockedNotRunnable's own settle
	// wait).
	deadline := time.Now().Add(500 * time.Millisecond)
	var shrunk int
	for {
		runtime.Gosched()
		shrunk = beadRunGoroutineCount(t)
		if shrunk == baseline+3 || time.Now().After(deadline) {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if shrunk != baseline+3 {
		t.Fatalf("after shrinking to 3 beads: got %d Bead.run goroutines (baseline %d), want %d — a removed bead's goroutine leaked", shrunk, baseline, baseline+3)
	}

	// Back to zero on this edge: every bead this test started is gone.
	m.reconcileBeadChain("b", 0, offsetAt, wire.Vec3{X: 1})
	deadline = time.Now().Add(500 * time.Millisecond)
	var final int
	for {
		runtime.Gosched()
		final = beadRunGoroutineCount(t)
		if final == baseline || time.Now().After(deadline) {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if final != baseline {
		t.Fatalf("after shrinking to 0 beads: got %d Bead.run goroutines, want back to baseline %d", final, baseline)
	}
}

// TestReconcileBeadChainBroadcastsOnAimChangeOnly: a second reconcile call at the SAME
// count and the SAME aim must not re-broadcast (PLAN.md "idle costs nothing" — an
// unchanged chain issues no fresh geometry generation), while a changed aim does.
// Observed indirectly: an unchanged reconcile must not invalidate an already-valid
// snapshot's position (it would, if it re-broadcast a value the bead's own goroutine had
// not yet had a chance to apply — since a fresh broadcast always starts a bead back at a
// stale cached position until its own goroutine services the new generation).
func TestReconcileBeadChainAppliesPosition(t *testing.T) {
	m := &nodeGeometry{id: "a", beadTickFn: wire.NewTickChan}
	offsetAt := func(i int) float64 { return 10 }
	c := m.reconcileBeadChain("b", 1, offsetAt, wire.Vec3{X: 1})
	defer close(c.stops[0])

	deadline := time.Now().Add(500 * time.Millisecond)
	for {
		c = m.reconcileBeadChain("b", 1, offsetAt, wire.Vec3{X: 1})
		if len(c.valid) == 1 && c.valid[0] {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("bead never reported a valid snapshot")
		}
		time.Sleep(2 * time.Millisecond)
	}
	got := c.last[0].Position
	want := wire.Vec3{X: 10}
	if got != want {
		t.Fatalf("bead position = %+v, want %+v (offsetR=10 along aim X=1)", got, want)
	}
}

// TestStartEndBeadDragTogglesEveryChain: startBeadDrag/endBeadDrag reach EVERY one of this
// node's own outgoing-edge chains with one StartDrag/EndDrag call each — not a per-bead
// send loop — mirroring the primitive-level TestWakeSetsEveryAffectedBead but through the
// production entry points handle() drives (moveMsgKindDragStart/moveMsgKindDragEnd).
func TestStartEndBeadDragTogglesEveryChain(t *testing.T) {
	m := &nodeGeometry{id: "a", beadTickFn: wire.NewTickChan}
	offsetAt := func(i int) float64 { return float64(i) }
	cB := m.reconcileBeadChain("b", 2, offsetAt, wire.Vec3{X: 1})
	cC := m.reconcileBeadChain("c", 2, offsetAt, wire.Vec3{Y: 1})
	defer func() {
		for _, s := range cB.stops {
			close(s)
		}
		for _, s := range cC.stops {
			close(s)
		}
	}()

	m.startBeadDrag()
	waitAllDragging(t, cB, true)
	waitAllDragging(t, cC, true)

	m.endBeadDrag()
	waitAllDragging(t, cB, false)
	waitAllDragging(t, cC, false)
}

func waitAllDragging(t *testing.T, c *edgeBeadChain, want bool) {
	t.Helper()
	deadline := time.Now().Add(500 * time.Millisecond)
	for {
		allMatch := true
		for i := range c.snaps {
			select {
			case s := <-c.snaps[i]:
				c.last[i] = s
				c.valid[i] = true
			default:
			}
			if !c.valid[i] || c.last[i].Dragging != want {
				allMatch = false
			}
		}
		if allMatch {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("not every bead reached Dragging=%v in time", want)
		}
		time.Sleep(2 * time.Millisecond)
	}
}
