// edge_mover_run.go — edgeMover's per-goroutine loop. See edge_mover.go for the actor's
// held state and edge_mover_stream.go for the per-fd frame write this loop calls.

package Wiring

import (
	"context"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// run is the edge's per-goroutine loop. It IS the wire's own goroutine
// (MODEL.md "The network" — PacedWire is an active goroutine, and it is this
// same per-edge goroutine that already existed to revise in-flight geometry,
// not an additional one): every cycle it drains any pending move/speed
// messages without blocking, then drives its dest wire's ONE cycle of bead
// ownership (DriveOneCycle — placement drain, position-step, delivery
// handoff), then paces to the next cycle on its OWN clock copy. This is what
// lets ReviseInFlightGeometry (called from handle, edge_mover.go, on this SAME
// goroutine) touch pw.inflight: there is exactly one goroutine
// on either side of that call.
func (m *edgeMover) run(ctx context.Context) {
	// Copy taken ONCE at this goroutine's start (run IS the goroutine). If no clockSrc was
	// given (bare test construction), keep the inert placeholder newEdgeMover
	// seeded m.clk with.
	if m.clockSrc != nil {
		m.clk = m.clockSrc.Copy()
	}
	// ONE-TIME startup geometry emit, on THIS edge's own mover goroutine — this is now
	// the sole per-owner source of an edge's initial geometry event (replacing the old
	// source-node-Update-loop startup emit builders.go's EmitGeometry closure used to
	// make for each of its outgoing edges; that closure no longer calls tr.Geometry —
	// see its doc comment). m.tr is non-nil in production; bare test construction with
	// a nil tr just skips this, matching recomputeGeometry's own nil-guard elsewhere.
	if m.tr != nil {
		m.recomputeGeometry()
	}
	for {
		// Drain extIn/srcIn/dstIn/speedCh without blocking, so a cycle always reaches
		// the wire-drive step below even with nothing queued. Three dedicated channels,
		// not one shared inbox: extIn (external gesture entries), srcIn (this edge's
		// source node's own goroutine), dstIn (this edge's target node's own goroutine).
		//
		// Drain-until-empty, transitively bounded by each channel's own declared
		// capacity (moverInboxDepth) -- no iteration cap; see
		// nodes/wire/paced_wire_drive.go's drainPlacements doc comment for the full
		// reasoning shared by every drain-until-empty loop in this repo.
	drain:
		for {
			select {
			case <-ctx.Done():
				return
			case sp := <-m.speedCh:
				// Delivery (per-goroutine-clock.md): apply directly to this
				// goroutine's own clk copy — nothing else reaches it.
				if rc, ok := m.clk.(*wire.RealClock); ok {
					rc.SetSpeed(sp)
				}
			case steps := <-m.stepsIn:
				// Delivery (stepsIn's doc comment): fold the source node's freshest
				// published step count into this edgeMover's own cached copy. A bare
				// value update, not a geometry recompute — the next recomputeGeometry
				// (on a move) or ReviseInFlightGeometry call reads m.steps fresh.
				m.steps = steps
			case msg := <-m.extIn:
				m.handle(msg)
				if msg.testDone != nil {
					close(msg.testDone)
				}
			case msg := <-m.srcIn:
				m.handle(msg)
				if msg.testDone != nil {
					close(msg.testDone)
				}
			case msg := <-m.dstIn:
				m.handle(msg)
				if msg.testDone != nil {
					close(msg.testDone)
				}
			default:
				break drain
			}
		}
		if m.dest != nil {
			// The wire is NOT driven here any more: its DriveOneCycle now runs on the
			// SOURCE NODE's own goroutine (nodeMover.run), which is what "the wire
			// goroutine is removed" means concretely — docs/beads-are-the-edge.md step 3.
			// This loop still writes the edge's own stream frame each cycle, because bead
			// positions may have moved under the node's drive.
			m.writeStreamFrame(m.clk.Tick(), nil)
		}
		if err := m.clk.SleepCycle(ctx); err != nil {
			return
		}
	}
}
