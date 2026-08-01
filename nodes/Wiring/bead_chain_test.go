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

// TestBeadGoroutineLifetimeFollowsChainLength: growing a chain via ensureBeadEdgeActors
// starts one goroutine per added bead; shrinking it back down closes each removed bead's
// own stop channel, and every removed bead's goroutine actually exits — no leak per
// removed bead. Bead CRUD adds/removes AT THE CHAIN END (MODEL.md), which
// ensureBeadEdgeActors implements directly.
func TestBeadGoroutineLifetimeFollowsChainLength(t *testing.T) {
	m := &nodeMover{id: "a"}

	baseline := beadRunGoroutineCount(t)

	m.ensureBeadEdgeActors("b", 10, 0)
	runtime.Gosched()
	time.Sleep(20 * time.Millisecond)
	grown := beadRunGoroutineCount(t)
	if grown != baseline+10 {
		t.Fatalf("after growing to 10 beads: got %d Bead.run goroutines (baseline %d), want %d", grown, baseline, baseline+10)
	}

	m.ensureBeadEdgeActors("b", 3, 0)
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

	m.ensureBeadEdgeActors("b", 0, 0)
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

// TestBroadcastAndReadAppliesPosition: chainBeads' entire read path — one
// BroadcastGeometry hop, then a read-back per bead — must return each bead's OWN
// resulting position, computed by that bead's own goroutine (Bead.applyTransform), never
// derived here. offsetR for bead 0 at selfTorusR=0 is exactly wire.BeadTorusOuterR (the
// same fixed formula ensureBeadEdgeActors bakes into every bead at construction), so
// Center{}+Aim{X:1}*offsetR is a value only the bead itself could have produced.
func TestBroadcastAndReadAppliesPosition(t *testing.T) {
	m := &nodeMover{id: "a"}
	ea := m.ensureBeadEdgeActors("b", 1, 0)
	defer close(ea.stops[0])

	positions := ea.broadcastAndRead(wire.BeadGeometryIn{Center: wire.Vec3{}, Aim: wire.Vec3{X: 1}})
	if len(positions) != 1 {
		t.Fatalf("broadcastAndRead returned %d positions, want 1", len(positions))
	}
	want := wire.Vec3{X: wire.BeadTorusOuterR}
	if positions[0] != want {
		t.Fatalf("bead position = %+v, want %+v (offsetR=BeadTorusOuterR along aim X=1)", positions[0], want)
	}

	// A second broadcast with a different aim must also be read back correctly — this
	// is the "no fallback, no stale mirror" property: every call re-derives from the
	// bead's own current state, never from a cached value chainBeads computed itself.
	positions2 := ea.broadcastAndRead(wire.BeadGeometryIn{Center: wire.Vec3{Y: 5}, Aim: wire.Vec3{Y: 1}})
	want2 := wire.Vec3{Y: 5 + wire.BeadTorusOuterR}
	if positions2[0] != want2 {
		t.Fatalf("bead position after second broadcast = %+v, want %+v", positions2[0], want2)
	}
}

// TestStartEndBeadDragTogglesEveryChain: startAllBeadDrags/endAllBeadDrags reach EVERY
// one of this node's own outgoing-edge chains with one StartDrag/EndDrag call each per
// chain (not a per-bead send loop) — the production entry points handle() drives from
// moveMsgKindDragStart/moveMsgKindDragEnd.
func TestStartEndBeadDragTogglesEveryChain(t *testing.T) {
	m := &nodeMover{id: "a"}
	eaB := m.ensureBeadEdgeActors("b", 2, 0)
	eaC := m.ensureBeadEdgeActors("c", 2, 0)
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
