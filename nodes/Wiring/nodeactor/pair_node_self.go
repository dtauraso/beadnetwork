// pair_node_self.go — PairNodeSelf, the handle a PAIR-scene node kind (PairNode) uses
// to own its own NodeGeometry directly, on its own Update goroutine, instead of through a
// separate NodeMover actor (task/pair-node-owns-itself). See MODEL.md and this package's
// node_geometry.go/node_mover.go for the split's own doc comments — nothing about what a
// node's geometry IS changes here; only which goroutine drives it, for exactly this
// kind.
//
// THE RING IS UNTOUCHED: every ring node still gets a real NodeMover (its own goroutine,
// launched by package Wiring's moverRegistry.start — node_mover.go). PairNodeSelf only
// ever wraps a *NodeGeometry whose id a kind explicitly claimed via
// BuildArgs.ClaimSelfDrive — package Wiring's moverRegistry.finalizeActors then never
// constructs a NodeMover for that id at all, so there is nothing for mr.start to skip.
package nodeactor

import (
	"context"

	T "github.com/dtauraso/wirefold/Trace"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

// PairNodeSelf wraps this node's own *NodeGeometry so a pair kind's Update loop can drive
// it directly. Every method here must be called ONLY from that node's own goroutine —
// the same single-goroutine-ownership contract node_geometry.go's own doc comment states,
// just satisfied by a different (but still exactly one) caller. Nil-safe throughout:
// package Wiring's BuildArgs.ClaimSelfDrive returns nil on a bare test build with no
// loader, matching every other nil-safe fallback in that file; every method on a nil
// *PairNodeSelf is itself a no-op.
type PairNodeSelf struct {
	geom *NodeGeometry
	// speedCh is geom.clk's OWN buffered-1 speed-delivery channel — the exact analogue of
	// NodeMover.speedCh for a ring node (package Wiring's moverRegistry.finalizeActors).
	// Without this, geom.clk was copied ONCE at build time (ClaimSelfDrive) and never
	// touched again: the kind's own SEPARATE clock (PairNode's own n.Clock, polled via its
	// own SpeedCh in the kind's Update loop) paced wire delivery correctly, but geom.clk —
	// the clock writeStreamFrame's frame tick and chainBeads' animation actually read —
	// stayed frozen at whatever speed it had at load, so a pair node's RENDERED bead motion
	// never reflected SceneTab.ClockDivisor or a live slider change at all. Polled once per
	// Step cycle below, mirroring NodeMover.Run's "ApplySpeedNonBlocking every cycle" shape
	// exactly.
	speedCh <-chan float64
}

// NewPairNodeSelf wraps geom (and its optional speed channel) as a PairNodeSelf — the
// exported door package Wiring's BuildArgs.ClaimSelfDrive (build_args_selfdrive.go) uses
// now that PairNodeSelf's fields are unexported across the package boundary
// (docs/planning/movedispatch-decomposition.md §20; ClaimSelfDrive used to build this
// value with a bare struct literal, `&PairNodeSelf{geom: ng, speedCh: speedCh}`, while both
// types lived in the same package). speedCh may be nil (no speed sink wired — matches the
// old literal's zero-value default in that branch).
func NewPairNodeSelf(geom *NodeGeometry, speedCh <-chan float64) *PairNodeSelf {
	return &PairNodeSelf{geom: geom, speedCh: speedCh}
}

// Breadcrumb emits a DEBUG breadcrumb on this node's own stream (.claude/rules/
// go-debugging.md) — the pair kinds have no *T.Trace of their own, so this is how they
// reach the one their geometry already holds. Diagnostic only: read back with
// `tools/probe-merge.sh --debug`, never gated on the probe.trace setting, and a no-op when
// no stream is wired (a headless or bare test build).
//
// Keep call sites SPARSE. These are for control events — an arrival, a decision, a user
// edit — not a per-tick firehose; a per-tick breadcrumb once grew a probe log past a
// gigabyte in this repo.
func (p *PairNodeSelf) Breadcrumb(label, value string) {
	if p == nil || p.geom == nil || p.geom.tr == nil {
		return
	}
	// The in-process TEST sink. Production has no sink wired here, so this call alone
	// reaches nothing on a live run (Trace.Breadcrumb's own doc comment) — which is what
	// silently discarded every pair breadcrumb until now.
	p.geom.tr.Breadcrumb(label, p.geom.id, "", value)
	// The PRODUCTION path: a structured KindBreadcrumb event on this node's OWN dedicated
	// stream frame, the same shape NodeGeometry's drag.commit breadcrumb uses. An unknown
	// label is dropped rather than sent as a bad id — the decode side indexes
	// T.BreadcrumbLabels by this number.
	id, ok := T.BreadcrumbLabelID(label)
	if !ok {
		return
	}
	p.geom.writeStreamFrame([]wire.RowEvent{{
		Kind: T.KindBreadcrumb, Label: id, Debug: 1,
		NodeRow: p.geom.stream.nodeRow, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Text: value,
	}})
}

// EmitGeometryOnce sends this node's initial node-geometry frame — the one-time startup
// emit a ring's NodeMover.Run makes at goroutine start (see its own doc comment),
// reproduced here since this node's own Update loop is that goroutine now.
func (p *PairNodeSelf) EmitGeometryOnce() {
	if p == nil || p.geom == nil {
		return
	}
	if p.geom.tr != nil {
		p.geom.emitGeometry()
	}
}

// Step runs exactly one cycle of this node's own geometry work — the same per-cycle body
// NodeMover.Run drives for a ring node (drain every dedicated inbound channel to empty,
// drive this node's own outgoing wires one cycle on the given tick, retry any pending
// sends, write this node's dedicated stream frame), called from the OWNING kind's own
// goroutine instead of a NodeMover actor. There is no pacing sleep here: the caller's own
// Update loop already paces itself on its own clock (per-goroutine-clock.md), and driving
// this geometry an extra time per caller cycle is exactly what "one goroutine, one clock
// reading" requires — a second sleep here would just double-pace the same node.
func (p *PairNodeSelf) Step(ctx context.Context, tick int64) {
	if p == nil || p.geom == nil {
		return
	}
	g := p.geom
	clock.ApplySpeedNonBlocking(g.clocks.clk, p.speedCh)
	for {
		progressed := false
		select {
		case msg := <-g.msg.extIn:
			g.handle(msg)
			if msg.TestDone != nil {
				close(msg.TestDone)
			}
			progressed = true
		default:
		}
		for _, ch := range g.msg.neighborIn {
			select {
			case msg := <-ch:
				g.handle(msg)
				if msg.TestDone != nil {
					close(msg.TestDone)
				}
				progressed = true
			default:
			}
		}
		if !progressed {
			break
		}
	}
	for _, pw := range g.outs.outWires {
		pw.DriveOneCycle(ctx, tick)
	}
	g.flushPending()
	g.writeStreamFrame(nil)
}

// SetTiltIndex applies this node's own new top/normal/bottom tilt-vector index triple
// directly to its own geometry state — the direct-call replacement for the removed
// movemsg.KindTiltIndexSync message-to-self (see that constant's retirement note in
// move_msg.go). Same effect as that message's old handle() branch: persist to this
// node's OWN position.json and re-emit.
//
// A TILT NO LONGER MOVES THE NODE. The pair's separation used to be a second reading of
// this same index — the other node slid along its own ray to
// (|theta| + torus steps) × BeadStepR from this one on every change — so the edge grew and
// shrank as the exchange turned. That is removed: the angle is now only an angle, and the
// separation stays wherever a drag last put it. What went with it: the moving node also
// changed the edge's bead-step count, which changed the crossing time, which (now that the
// exchange is clock-paced) fed back into how fast the tilts turned. One index driving an
// angle, a distance, a bead count and a pace was more than one number should carry.
func (p *PairNodeSelf) SetTiltIndex(theta, normalTheta, bottomTheta int32) {
	if p == nil || p.geom == nil {
		return
	}
	g := p.geom
	g.tilt.topTiltVectorThetaIdx = theta
	g.tilt.normalThetaIdx = normalTheta
	g.tilt.bottomThetaIdx = bottomTheta
	g.persistTiltVectorAngle()
	if g.tr != nil {
		g.emitGeometry()
	}
}

// SetRoundsToParallel reports this node's own rounds-to-rest count to its own geometry,
// so the Node block's RoundsToParallel column carries it. Same one-way shape as
// SetTiltIndex: the node's own goroutine counts, the geometry only mirrors.
//
// It is called ONCE, at the moment this node's rule first comes to rest after the exchange
// opened — not per round. A per-round call would make the column climb for as long as the
// scene stays open, because the exchange keeps circulating after both ends settle
// (stepFromVector replies to every arrival whether or not it moved), and the number a
// reader wants is how far the tilt had to travel, not how long they have been watching.
func (p *PairNodeSelf) SetRoundsToParallel(rounds, msgs int32) {
	if p == nil || p.geom == nil {
		return
	}
	g := p.geom
	g.readout.roundsToParallel = rounds
	g.readout.msgsToParallel = msgs
	if g.tr != nil {
		g.emitGeometry()
	}
}

// SetReceivedVector applies this node's own last-received vector-channel direction
// directly to its own geometry state — the direct-call replacement for the removed
// movemsg.KindReceivedVectorSync message-to-self. Same effect as that message's old
// handle() branch: re-emit so the third drawn arrow picks up the change; nothing here is
// persisted (a channel arrival is transient session state).
func (p *PairNodeSelf) SetReceivedVector(theta int32, set bool) {
	if p == nil || p.geom == nil {
		return
	}
	g := p.geom
	g.tilt.receivedVectorThetaIdx = theta
	g.tilt.receivedVectorSet = set
	if g.tr != nil {
		g.emitGeometry()
	}
}

// SetLatticePoints applies this node's own new lattice point count directly to its own
// geometry state — the direct-call replacement for a message-to-self, same one-way shape
// as SetTiltIndex above. It exists because the angle a tilt-vector INDEX draws depends on
// how many points the lattice has (2π / points per step, node_geometry_stream.go's
// writeStreamFrame): the geometry converts index → angle every frame, but it does not
// itself decide the point count — that is a scene setting PairNode's own goroutine owns
// (Node.adoptLattice) — so it has to be told. Re-emits so the drawn angles pick up the
// new step on the very next frame.
func (p *PairNodeSelf) SetLatticePoints(points int32) {
	if p == nil || p.geom == nil {
		return
	}
	g := p.geom
	g.tilt.latticePoints = points
	if g.tr != nil {
		g.emitGeometry()
	}
}

// ClearOutBeads empties every one of this node's own outgoing wires directly — the
// direct-call replacement for the removed movemsg.KindBeadClear message-to-self. This
// node's own geometry already drives those wires (Step, above), so it may clear them
// itself with no message needed at all.
func (p *PairNodeSelf) ClearOutBeads() {
	if p == nil || p.geom == nil {
		return
	}
	for _, pw := range p.geom.outs.outWires {
		pw.ClearInFlight()
	}
}
