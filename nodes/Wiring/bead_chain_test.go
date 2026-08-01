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
		snaps := ea.broadcastAndRead(xf1, 1, nil, false)
		if len(snaps) != 1 {
			t.Fatalf("broadcastAndRead returned %d snapshots, want 1", len(snaps))
		}
		got = snaps[0].Position
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
	geom, wake, settle, anim := group.Current()
	stop := make(chan struct{})
	defer close(stop)
	b := wire.NewBead(1.0, 0, geom, wake, settle, anim, make(chan struct{}), stop)
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

	done := make(chan []wire.BeadSnapshot, 1)
	go func() {
		done <- ea.broadcastAndRead(wire.BeadGeometryIn{Center: wire.Vec3{Y: 99}, Aim: wire.Vec3{X: 1}}, 1, nil, false)
	}()
	select {
	case snaps := <-done:
		if snaps[0].Position != seed.Position {
			t.Fatalf("expected the cached seeded position %+v from an unresponsive bead, got %+v", seed.Position, snaps[0].Position)
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

// TestIdleIssuesNoBroadcastWhenNothingChanged is GAP 3: with no drag and no geometry
// change, repeated broadcastAndRead calls must issue ZERO new broadcasts and wake no bead.
// GeomGen (each bead's own count of geometry broadcasts it has actually PROCESSED,
// incremented only inside Bead.run's geom case) is the observable proxy for "was this bead
// woken" — it can only advance if a real BroadcastGeometry closed a fresh chain link for
// this bead to receive, so a flat GeomGen across many calls is direct proof no broadcast
// (and therefore no wake) was issued.
func TestIdleIssuesNoBroadcastWhenNothingChanged(t *testing.T) {
	m := &nodeMover{id: "a"}
	xf := wire.BeadGeometryIn{Center: wire.Vec3{}, Aim: wire.Vec3{X: 1}}
	ea := m.ensureBeadEdgeActors("b", 1, 0, xf)
	defer close(ea.stops[0])

	// First call always broadcasts (this group has never broadcast geometry before —
	// ea.haveGeom starts false) — poll until the bead has caught up to it. SeedGeometry
	// (ensureBeadEdgeActors, called above) already set GeomGen=1 on ea.last[0] BEFORE this
	// bead's goroutine ever processed a real broadcast, so ">= 1" would be trivially true
	// from the seed alone; wait for ">= 2" (seed's 1, plus this call's own real broadcast)
	// so "steady" below reflects an ACTUALLY-OBSERVED broadcast completion, not the seed.
	var snaps []wire.BeadSnapshot
	deadline := time.Now().Add(500 * time.Millisecond)
	for {
		snaps = ea.broadcastAndRead(xf, 1, nil, false)
		if snaps[0].GeomGen >= 2 || time.Now().After(deadline) {
			break
		}
		time.Sleep(2 * time.Millisecond)
	}
	steady := snaps[0].GeomGen
	if steady < 2 {
		t.Fatalf("bead never applied the initial broadcast; GeomGen=%d, want >= 2 (seed + one real broadcast)", steady)
	}

	// Every one of these calls repeats the SAME xf/count, not dragging — GAP 3 requires
	// zero further broadcasts, so GeomGen must not move.
	for i := 0; i < 50; i++ {
		snaps = ea.broadcastAndRead(xf, 1, nil, false)
	}
	if snaps[0].GeomGen != steady {
		t.Fatalf("GeomGen advanced from %d to %d across 50 unchanged, non-dragging calls — idle must cost zero broadcasts (PLAN.md 'idle CPU at zero')", steady, snaps[0].GeomGen)
	}
}

// TestDraggingBroadcastsEveryCall is GAP 4's positive half: while dragging is true, EVERY
// call re-broadcasts geometry — even with xf/count unchanged — so the two modes are
// functionally distinct (not just a flag that changes nothing). GeomGen must strictly
// advance on every single call.
func TestDraggingBroadcastsEveryCall(t *testing.T) {
	m := &nodeMover{id: "a"}
	xf := wire.BeadGeometryIn{Center: wire.Vec3{}, Aim: wire.Vec3{X: 1}}
	ea := m.ensureBeadEdgeActors("b", 1, 0, xf)
	defer close(ea.stops[0])

	prev := 0
	for i := 0; i < 5; i++ {
		var snaps []wire.BeadSnapshot
		deadline := time.Now().Add(500 * time.Millisecond)
		for {
			snaps = ea.broadcastAndRead(xf, 1, nil, true)
			if snaps[0].GeomGen > prev || time.Now().After(deadline) {
				break
			}
			time.Sleep(2 * time.Millisecond)
		}
		if snaps[0].GeomGen <= prev {
			t.Fatalf("call %d: GeomGen did not advance past %d while dragging with unchanged xf/count — dragging must re-broadcast every call (PLAN.md 'machine time')", i, prev)
		}
		prev = snaps[0].GeomGen
	}
}

// TestNodeMoverDraggingFlagSetOncePerDrag: nodeMover.dragging — the flag chainBeads reads
// to gate geometry broadcast — is set by moveMsgKindDragStart and cleared by
// moveMsgKindDragEnd, exactly the flag GAP 4 requires (never toggled per pointer event: a
// drag sends dragStart once and dragEnd once, regardless of how many moveMsgKindDrag
// messages fall in between).
func TestNodeMoverDraggingFlagSetOncePerDrag(t *testing.T) {
	m := &nodeMover{id: "a"}
	if m.dragging {
		t.Fatal("nodeMover.dragging must start false")
	}
	m.handle(moveMsg{Kind: moveMsgKindDragStart, NodeID: "a"})
	if !m.dragging {
		t.Fatal("moveMsgKindDragStart must set nodeMover.dragging")
	}
	// Intervening move messages (the per-pointer-event traffic during a drag) must not
	// touch the flag.
	m.handle(moveMsg{Kind: moveMsgKindDrag, NodeID: "a", Target: vec3{}})
	if !m.dragging {
		t.Fatal("an intervening moveMsgKindDrag cleared nodeMover.dragging — the flag must only change on dragStart/dragEnd")
	}
	m.handle(moveMsg{Kind: moveMsgKindDragEnd, NodeID: "a"})
	if m.dragging {
		t.Fatal("moveMsgKindDragEnd must clear nodeMover.dragging")
	}
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
