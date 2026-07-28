// paced_wire_concurrency_race_test.go — CHECKS BY CODE the ownership invariant the
// PacedWire doc comment claims: inflight/nextGen are touched by EXACTLY
// ONE goroutine (the wire's own — driveOneCycle/ReviseInFlightGeometry, folded into
// edgeMover.run in production), while the SOURCE and DESTINATION goroutines touch
// only the channel-shaped cross-goroutine surface (Send/RecvTick). No lock guards
// any of this — ownership replaces locking.
//
// This supersedes the old three-goroutine pw.mu contention test (source
// place+step, dest recv, edge geometry-revise all reaching into one mutex): under
// the new model there is only ONE state-owning goroutine, so the property to prove
// is the opposite of the old test's — that touching the wire ONLY through its
// channel surface (Send/RecvTick) from other goroutines, concurrently with the
// owning goroutine driving cycles AND revising geometry on itself, is race-free
// with NO mutex at all.
package wire

import (
	"context"
	"math"
	"sync"
	"testing"
	"time"
)

// TestPacedWireOwnershipUnderRace drives a PacedWire the way production actually
// does post-restructure: one goroutine plays the wire's OWN goroutine role
// (repeatedly driving cycles and revising in-flight geometry — exactly what
// edgeMover.run folds together), a second goroutine only ever calls Send (the
// source node's role), and a third only ever calls RecvTick (the destination
// node's role). Run under `go test -race`, this reports a DATA RACE if the
// ownership split breaks down (e.g. a future change lets a non-owning goroutine
// touch pw.inflight directly).
func TestPacedWireOwnershipUnderRace(t *testing.T) {
	pw := NewPacedWire(0, PulseSpeedWuPerMs)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const window = 300 * time.Millisecond
	deadline := time.Now().Add(window)

	var wg sync.WaitGroup

	// The wire's OWN goroutine: drives cycles and revises in-flight geometry on
	// itself, exactly as edgeMover.run does (this IS the wire's goroutine — see
	// MODEL.md "The network").
	wg.Add(1)
	go func() {
		defer wg.Done()
		clk := NewRealClock()
		i := 0.0
		for time.Now().Before(deadline) {
			i++
			pw.DriveOneCycle(ctx, clk.Tick())
			// Vary geometry (to exercise the race-detector on a live revise)
			// WITHOUT letting arc grow unboundedly: an ever-increasing arc would
			// push every in-flight bead's delivery deadline further into the
			// future every single cycle, so beads could never reach it and
			// inflight would accumulate without bound by this test's own
			// construction -- unrelated to any real bug, but exactly the shape
			// maxInflightBeads (paced_wire.go) now catches. math.Mod bounds the
			// variation to a small, repeating range instead.
			j := math.Mod(i, 50)
			pw.ReviseInFlightGeometry(clk.Tick(), 4+j*0.01, WireSegment{Start: Vec3{}, End: Vec3{X: 1 + j*0.001}})
			time.Sleep(time.Millisecond)
		}
	}()

	// Source goroutine: only ever calls Send — never touches wire-owned state
	// directly, exactly as Out.placeDrivenNoWalker does. A brief per-send sleep
	// keeps this within the model's own stated assumption (wireChanBufferSize's
	// doc comment: "a node fires at most a handful of times between two
	// wire-drive cycles") -- an unthrottled tight loop would enqueue far faster
	// than any wire can ever drain/deliver by construction (each bead takes
	// ~100 ticks to cross at this arc/pulseSpeed), which is a real design bug
	// were it production traffic, not something maxInflightBeads should paper
	// over by being sized larger.
	wg.Add(1)
	go func() {
		defer wg.Done()
		val := 0
		for time.Now().Before(deadline) {
			val++
			pw.Send(val, beadPlacement{InFlightMs: 4, Start: Vec3{}, End: Vec3{X: 1}, Node: "src", Port: "Out"})
			time.Sleep(time.Millisecond)
		}
	}()

	// Dest goroutine: only ever calls RecvTick, exactly as In.PollRecv does.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for time.Now().Before(deadline) {
			pw.RecvTick()
		}
	}()

	wg.Wait()
}
