// node_mover.go — nodeMover, the RING-ONLY actor: its own goroutine, its own inbox drain,
// its own clock-paced loop, wrapping a *nodeGeometry it owns (node_geometry.go). A PAIR
// node (Node1/Node2, task/pair-node-owns-itself) has NO nodeMover at all — its own kind
// goroutine owns a *nodeGeometry directly (BuildArgs.ClaimSelfDrive, pair_node_self.go) —
// there is nothing here for it to skip launching.
//
// Every field that is genuinely about GEOMETRY (position, wires, retry queue, tilt state,
// persistence, the dedicated stream) lives on nodeGeometry, not here. What stays here is
// ACTOR-ONLY: the pacing clock-poll (speedCh/ApplySpeedNonBlocking) and the goroutine loop
// itself (run). See node_geometry.go's own header comment for the full split.

package Wiring

import (
	"context"
	"path/filepath"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// --- node path construction ---
//
// A node owns every path under its own <root>/nodes/<id>/ directory EXCEPT
// nodes/<id>/edges/ (that subtree belongs to the edgeMover of each edge leaving this
// node — see edge_mover.go's doc comment and .claude/rules/persistence-ownership.md
// "The model"). These are the ONLY functions in the package that build those paths;
// quant_offset_persist.go and scene_anchor_persist.go call them rather than
// constructing the path themselves. safeTreePathComponent (scene_persist.go) is applied
// at each call site before use, same as it always was — node ids and port names are
// spec-authored and must not escape the tree root via ".." or a separator.

// positionFilePath is <root>/nodes/<id>/position.json — a node's exact scene-polar
// position plus its quantized-scalar-triple cache (quant_offset_persist.go).
func positionFilePath(root, id string) string {
	return filepath.Join(root, "nodes", id, "position.json")
}

// pendingSend is one (destination, message) pair this node's own goroutine tried to
// deliver, failed (the target's inbox was momentarily full), and is retrying — see
// nodeGeometry.pending's doc comment. There is no separate sender goroutine: only this
// node's own driving goroutine ever reads or writes it.
type pendingSend struct {
	destID string
	msg    moveMsg
}

// nodeMover is the RING actor: its own goroutine (run, launched by moverRegistry.start)
// and its own speed channel, driving a *nodeGeometry it owns. It carries no geometry
// fields of its own — everything about what a node's geometry IS lives on geom.
type nodeMover struct {
	geom *nodeGeometry
	// speedCh delivers a speed change to THIS node's own clk copy
	// (per-goroutine-clock.md "Delivery"), polled via ApplySpeedNonBlocking every cycle
	// of run's loop. Set once, at construction (newMoveDispatch's finalize pass), from
	// the loader's build-wide speed-sink accumulator; nil in bare test construction,
	// which is fine — ApplySpeedNonBlocking is a no-op on a nil channel. A PAIR node has
	// no nodeMover and therefore no speedCh of its own — its own kind goroutine paces
	// itself already.
	speedCh chan float64
}

// newNodeMover wraps geom in a RING actor. Only called for a node that never claims
// BuildArgs.ClaimSelfDrive (see MoveDispatch's finalizeActors, node_move.go).
func newNodeMover(geom *nodeGeometry) *nodeMover {
	return &nodeMover{geom: geom}
}

// run is the node's per-goroutine move loop. It paces itself on its OWN clock copy the
// same way every other loop in the system does (edgeMover.run, DriveHeld,
// emitRefillSlide): a Clock.Copy() taken once here at goroutine start, ApplySpeedNonBlocking
// polled once per cycle, and SleepCycle(ctx) as the pacing sleep at the bottom of the loop.
//
// Each cycle FIRST drains every one of geom's OWN dedicated inbound channels (extIn + one
// per neighbor) — there is no shared inbox to drain — non-blockingly and acts on whatever
// is there, repeating the drain pass until a full pass finds nothing left (so a backlog on
// any one channel is fully drained before the cycle paces, not throttled to "one message
// per channel per cycle"), THEN retries any pending sends, THEN sleeps one clock cycle.
// Nothing here ever blocks on a receive OR a send: an empty channel just falls through its
// `default`, exactly the "read non-blockingly at the top, act on what's there, pace on the
// clock" shape the design calls for.
func (m *nodeMover) run(ctx context.Context) {
	g := m.geom
	if g.clockSrc != nil {
		g.clk = g.clockSrc.Copy()
	}
	// ONE-TIME startup geometry emit, on THIS node's own goroutine — the sole per-owner
	// source of a node's initial node-geometry event.
	if g.tr != nil {
		g.emitGeometry()
	}
	for {
		wire.ApplySpeedNonBlocking(g.clk, m.speedCh)
		// Drain-until-empty, transitively bounded by each channel's own declared
		// capacity (moverInboxDepth) -- no iteration cap; see
		// nodes/wire/paced_wire.go's drainPlacements doc comment for the full
		// reasoning shared by every drain-until-empty loop in this repo.
		for {
			progressed := false
			select {
			case <-ctx.Done():
				return
			case msg := <-g.extIn:
				g.handle(msg)
				if msg.testDone != nil {
					close(msg.testDone)
				}
				progressed = true
			default:
			}
			for _, ch := range g.neighborIn {
				select {
				case msg := <-ch:
					g.handle(msg)
					if msg.testDone != nil {
						close(msg.testDone)
					}
					progressed = true
				default:
				}
			}
			if !progressed {
				break
			}
		}
		// Drive THIS node's own outgoing wires — placement drain, position step, delivery
		// — on this node's own goroutine and its own clock reading. This is the work
		// edgeMover.run used to do for the wire; the wire has no goroutine of its own
		// (docs/beads-are-the-edge.md step 3). Driving it here is also what makes
		// LiveBeadFractions safe to read below: same goroutine, no shared state.
		//
		// Read the clock ONCE for this whole pass, not once per wire: a per-wire read
		// can straddle a tick boundary mid-loop, splitting one emission's beads across
		// two ticks even though they were placed microseconds apart.
		outTick := g.clk.Tick()
		for _, pw := range g.outWires {
			pw.DriveOneCycle(ctx, outTick)
		}
		// Retry any pending sends every cycle — a destination that was full earlier may
		// have drained since.
		g.flushPending()
		// Selection/hover/drag UI state may have changed even with no geometry change
		// this cycle — write this node's dedicated stream frame every cycle (no-op when
		// streamOut is nil), mirroring edgeMover.run's same every-cycle writeStreamFrame call.
		g.writeStreamFrame(nil)
		if err := g.clk.SleepCycle(ctx); err != nil {
			return
		}
	}
}
