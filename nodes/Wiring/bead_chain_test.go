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

// bead_chain_test.go — the production integration's own tests for the bead-actor bridge
// (bead_actor_bridge.go), distinct from nodes/wire's primitive-level tests
// (bead_actor_test.go). These exercise nodeMover.ensureBeadEdgeActors/broadcastAndRead/
// startAllBeadDrags/endAllBeadDrags directly — the real call sites chain_beads.go's
// chainBeads uses, and the ones handle()'s moveMsgKindDragStart/moveMsgKindDragEnd cases
// drive. PLAN.md "THE REPLACEMENT IS NOT DONE UNTIL THE OLD PATH IS DELETED" is enforced
// in SOURCE by tools/check-bead-position-not-central.sh; these tests are the BEHAVIOURAL
// half — a replacement and a mirror render identically, so the source guard is what tells
// them apart, but goroutine lifetime and read-back correctness still need behavioural
// coverage of their own.
//
// broadcastAndRead is NON-BLOCKING by design (its own doc comment in
// bead_actor_bridge.go): a returned position may be one broadcast (or more, under a fast
// drag) stale, and that staleness is deliberate — the node's own goroutine never waits on
// a bead's reply. TestBroadcastAndReadNeverBlocksOnUnresponsiveBead below is the
// regression test for the earlier, WRONG shape (blocking on an exact generation match
// against a lossy latest-wins channel), which could hang forever.

// beadRunGoroutineCount counts live goroutines currently inside wire.(*Bead).run, via the
// same goroutine-profile technique bead_actor_test.go's TestIdleBeadIsBlockedNotRunnable
// uses — the sanctioned way to observe another goroutine's existence without touching its
// owned state from a second goroutine.
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

var zeroXF = wire.BeadGeometryIn{}

// TestBeadGoroutineLifetimeFollowsChainLength: growing a chain via ensureBeadEdgeActors
// starts one goroutine per added bead; shrinking it back down closes each removed bead's
// own stop channel, and every removed bead's goroutine actually exits — no leak per
// removed bead. Bead CRUD adds/removes AT THE CHAIN END (MODEL.md), which
// ensureBeadEdgeActors implements directly.
func TestBeadGoroutineLifetimeFollowsChainLength(t *testing.T) {
	m := &nodeMover{id: "a"}

	baseline := beadRunGoroutineCount(t)

	m.ensureBeadEdgeActors("b", 10, 0, zeroXF)
	runtime.Gosched()
	time.Sleep(20 * time.Millisecond)
	grown := beadRunGoroutineCount(t)
	if grown != baseline+10 {
		t.Fatalf("after growing to 10 beads: got %d Bead.run goroutines (baseline %d), want %d", grown, baseline, baseline+10)
	}

	m.ensureBeadEdgeActors("b", 3, 0, zeroXF)
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

	m.ensureBeadEdgeActors("b", 0, 0, zeroXF)
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

// TestEnsureBeadEdgeActorsSeedsValidInitialPosition: a freshly created bead's position is
// correct from CONSTRUCTION, before its own goroutine has ever serviced a broadcast (or
// even been scheduled) — PLAN.md "a first frame never has nothing to read". Read via the
// group's own cache (ea.last), the same non-blocking path broadcastAndRead uses, with no
// broadcast issued at all.
func TestEnsureBeadEdgeActorsSeedsValidInitialPosition(t *testing.T) {
	m := &nodeMover{id: "a"}
	xf := wire.BeadGeometryIn{Center: wire.Vec3{X: 100}, Aim: wire.Vec3{Z: 1}}
	ea := m.ensureBeadEdgeActors("b", 1, 0, xf)
	defer close(ea.stops[0])

	want := wire.Vec3{X: 100, Z: wire.BeadTorusOuterR}
	if ea.last[0].Position != want {
		t.Fatalf("seeded position = %+v, want %+v (offsetR=BeadTorusOuterR along aim Z=1, from Center X=100)", ea.last[0].Position, want)
	}
}

// TestBroadcastAndReadAppliesPosition: chainBeads' entire read path — one
// BroadcastGeometry hop, then a non-blocking read-back per bead — must eventually return
// each bead's OWN resulting position, computed by that bead's own goroutine
// (Bead.applyTransform), never derived here. The read itself never blocks (see
// TestBroadcastAndReadNeverBlocksOnUnresponsiveBead), so this test polls
// broadcastAndRead — a legitimate TEST-side retry of a non-blocking call, not the
// production code waiting on a bead.
func TestBroadcastAndReadAppliesPosition(t *testing.T) {
	m := &nodeMover{id: "a"}
	xf1 := wire.BeadGeometryIn{Center: wire.Vec3{Y: 5}, Aim: wire.Vec3{Y: 1}}
	ea := m.ensureBeadEdgeActors("b", 1, 0, zeroXF)
	defer close(ea.stops[0])

	want := wire.Vec3{Y: 5 + wire.BeadTorusOuterR}
	deadline := time.Now().Add(500 * time.Millisecond)
	var got wire.Vec3
	for {
		positions := ea.broadcastAndRead(xf1)
		if len(positions) != 1 {
			t.Fatalf("broadcastAndRead returned %d positions, want 1", len(positions))
		}
		got = positions[0]
		if got == want || time.Now().After(deadline) {
			break
		}
		time.Sleep(2 * time.Millisecond)
	}
	if got != want {
		t.Fatalf("bead position after broadcast = %+v, want %+v (offsetR=BeadTorusOuterR along aim Y=1, from Center Y=5)", got, want)
	}
}

// TestBroadcastAndReadNeverBlocksOnUnresponsiveBead is the regression test for the
// defect this file replaces: a prior version blocked in a loop until a bead's observe
// channel produced an EXACT generation match, which could hang forever against a
// buffered-1 LATEST-WINS channel (a bead that has moved on to a later generation, or
// whose goroutine simply hasn't run yet, never satisfies an exact match, and there is no
// timeout on a bare channel receive).
//
// This bead's goroutine is deliberately NEVER STARTED, so its observe channel will NEVER
// produce a value — the strongest possible stand-in for "this bead will not answer
// within this call, whether it is behind, ahead, or simply not scheduled yet".
// broadcastAndRead must still return immediately, using the seeded/cached position,
// rather than waiting for a reply that will never come.
func TestBroadcastAndReadNeverBlocksOnUnresponsiveBead(t *testing.T) {
	group := wire.NewBeadWakeGroup()
	geom, wake, settle := group.Current()
	stop := make(chan struct{})
	defer close(stop)
	b := wire.NewBead(1.0, geom, wake, settle, make(chan struct{}), stop)
	obs := b.WithObserve()
	seed := b.SeedGeometry(wire.BeadGeometryIn{Center: wire.Vec3{}, Aim: wire.Vec3{X: 1}})
	// b.Start() is deliberately NOT called.

	ea := &beadEdgeActors{
		group: group,
		beads: []*wire.Bead{b},
		obs:   []<-chan wire.BeadSnapshot{obs},
		stops: []chan struct{}{stop},
		last:  []wire.BeadSnapshot{seed},
	}

	done := make(chan []wire.Vec3, 1)
	go func() {
		done <- ea.broadcastAndRead(wire.BeadGeometryIn{Center: wire.Vec3{Y: 99}, Aim: wire.Vec3{X: 1}})
	}()
	select {
	case positions := <-done:
		if positions[0] != seed.Position {
			t.Fatalf("expected the cached seeded position %+v from an unresponsive bead, got %+v", seed.Position, positions[0])
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("broadcastAndRead blocked on an unresponsive bead — the read path must never block on a bead's own goroutine (PLAN.md: one broadcast hop, no per-bead round trip)")
	}
}

// TestStartEndBeadDragTogglesEveryChain: startAllBeadDrags/endAllBeadDrags reach EVERY
// one of this node's own outgoing-edge chains with one StartDrag/EndDrag call each per
// chain (not a per-bead send loop) — the production entry points handle() drives from
// moveMsgKindDragStart/moveMsgKindDragEnd.
func TestStartEndBeadDragTogglesEveryChain(t *testing.T) {
	m := &nodeMover{id: "a"}
	eaB := m.ensureBeadEdgeActors("b", 2, 0, zeroXF)
	eaC := m.ensureBeadEdgeActors("c", 2, 0, zeroXF)
	defer func() {
		for _, s := range eaB.stops {
			close(s)
		}
		for _, s := range eaC.stops {
			close(s)
		}
	}()

	m.startAllBeadDrags()
	waitAllDragging(t, eaB, true)
	waitAllDragging(t, eaC, true)

	m.endAllBeadDrags()
	waitAllDragging(t, eaB, false)
	waitAllDragging(t, eaC, false)
}

func waitAllDragging(t *testing.T, ea *beadEdgeActors, want bool) {
	t.Helper()
	last := make([]wire.BeadSnapshot, len(ea.obs))
	valid := make([]bool, len(ea.obs))
	deadline := time.Now().Add(500 * time.Millisecond)
	for {
		for i, obs := range ea.obs {
			select {
			case s := <-obs:
				last[i] = s
				valid[i] = true
			default:
			}
		}
		allMatch := true
		for i := range last {
			if !valid[i] || last[i].Dragging != want {
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
