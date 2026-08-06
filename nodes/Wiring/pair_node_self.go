// pair_node_self.go — PairNodeSelf, the handle a PAIR-scene node kind (Node1/Node2) uses
// to own its own nodeGeometry directly, on its own Update goroutine, instead of through a
// separate nodeMover actor (task/pair-node-owns-itself). See MODEL.md and this package's
// node_geometry.go/node_mover.go for the split's own doc comments — nothing about what a
// node's geometry IS changes here; only which goroutine drives it, for exactly these two
// kinds.
//
// THE RING IS UNTOUCHED: every ring node still gets a real nodeMover (its own goroutine,
// launched by mr.start — node_mover.go). PairNodeSelf only ever wraps a *nodeGeometry
// whose id a kind explicitly claimed via BuildArgs.ClaimSelfDrive — mover_registry.go's
// finalizeActors then never constructs a nodeMover for that id at all, so there is nothing
// for mr.start to skip.
package Wiring

import (
	"context"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// PairNodeSelf wraps this node's own *nodeGeometry so a pair kind's Update loop can drive
// it directly. Every method here must be called ONLY from that node's own goroutine —
// the same single-goroutine-ownership contract node_geometry.go's own doc comment states,
// just satisfied by a different (but still exactly one) caller. Nil-safe throughout:
// BuildArgs.ClaimSelfDrive returns nil on a bare test build with no loader, matching
// every other nil-safe fallback in this package.
type PairNodeSelf struct {
	geom *nodeGeometry
	// speedCh is geom.clk's OWN buffered-1 speed-delivery channel — the exact analogue of
	// nodeMover.speedCh for a ring node (mover_registry.go's finalizeActors). Without this,
	// geom.clk was copied ONCE at build time (ClaimSelfDrive) and never touched again: the
	// kind's own SEPARATE clock (Node1/Node2's own n.Clock, polled via its own SpeedCh in
	// the kind's Update loop) paced wire delivery correctly, but geom.clk — the clock
	// writeStreamFrame's frame tick and chainBeads' animation actually read — stayed frozen
	// at whatever speed it had at load, so a pair node's RENDERED bead motion never reflected
	// SceneTab.ClockDivisor or a live slider change at all. Polled once per Step cycle below,
	// mirroring nodeMover.run's "ApplySpeedNonBlocking every cycle" shape exactly.
	speedCh <-chan float64
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
	p.geom.tr.Breadcrumb(label, p.geom.id, "", value)
}

// EmitGeometryOnce sends this node's initial node-geometry frame — the one-time startup
// emit a ring's nodeMover.run makes at goroutine start (see its own doc comment),
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
// nodeMover.run drives for a ring node (drain every dedicated inbound channel to empty,
// drive this node's own outgoing wires one cycle on the given tick, retry any pending
// sends, write this node's dedicated stream frame), called from the OWNING kind's own
// goroutine instead of a nodeMover actor. There is no pacing sleep here: the caller's own
// Update loop already paces itself on its own clock (per-goroutine-clock.md), and driving
// this geometry an extra time per caller cycle is exactly what "one goroutine, one clock
// reading" requires — a second sleep here would just double-pace the same node.
func (p *PairNodeSelf) Step(ctx context.Context, tick int64) {
	if p == nil || p.geom == nil {
		return
	}
	g := p.geom
	wire.ApplySpeedNonBlocking(g.clk, p.speedCh)
	for {
		progressed := false
		select {
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
	for _, pw := range g.outWires {
		pw.DriveOneCycle(ctx, tick)
	}
	g.flushPending()
	g.writeStreamFrame(nil)
}

// SetTiltIndex applies this node's own new top/normal/bottom tilt-vector index triple
// directly to its own geometry state — the direct-call replacement for the removed
// moveMsgKindTiltIndexSync message-to-self (see that constant's retirement note in
// move_msg.go). Same effect as that message's old handle() branch: persist to this
// node's OWN position.json, re-emit, and — PAIR TAB ONLY — reposition this node along
// its own fixed ray per repositionForTiltIndex's model (unchanged; see its own doc
// comment for the exact D formula and the Node2-only/Node1-anchor rule).
func (p *PairNodeSelf) SetTiltIndex(theta, normalTheta, bottomTheta int32) {
	if p == nil || p.geom == nil {
		return
	}
	g := p.geom
	g.topTiltVectorThetaIdx = theta
	g.normalThetaIdx = normalTheta
	g.bottomThetaIdx = bottomTheta
	g.persistTiltVectorAngle()
	if g.tr != nil {
		g.emitGeometry()
	}
	g.repositionForTiltIndex(theta)
}

// SetReceivedVector applies this node's own last-received vector-channel direction
// directly to its own geometry state — the direct-call replacement for the removed
// moveMsgKindReceivedVectorSync message-to-self. Same effect as that message's old
// handle() branch: re-emit so the third drawn arrow picks up the change; nothing here is
// persisted (a channel arrival is transient session state).
func (p *PairNodeSelf) SetReceivedVector(theta int32, set bool) {
	if p == nil || p.geom == nil {
		return
	}
	g := p.geom
	g.receivedVectorThetaIdx = theta
	g.receivedVectorSet = set
	if g.tr != nil {
		g.emitGeometry()
	}
}

// NodeSelfDriven reports whether node id's own geometry is driven by that node's own kind
// goroutine (task/pair-node-owns-itself, ClaimSelfDrive) rather than a separate nodeMover
// goroutine — equivalently, whether id has NO entry in the ring's nodeMover actor
// directory at all (finalizeActors never builds one for a claimed id). Exposed for
// verification: the model's whole point — one goroutine, not two, for the same node id —
// is otherwise invisible from outside this package (package main's own headless tests are
// the only place every kind, Node1/Node2 included, is registered — see
// kind_registry_parity_test.go's own doc comment).
func (md *MoveDispatch) NodeSelfDriven(id string) bool {
	if _, hasGeom := md.mr.nodeGeoms[id]; !hasGeom {
		return false
	}
	return !md.HasNodeMover(id)
}

// HasNodeMover reports whether node id has a real, separate nodeMover actor (a ring
// node) as opposed to no nodeMover at all (a self-driven pair node, or an unknown id).
func (md *MoveDispatch) HasNodeMover(id string) bool {
	_, ok := md.mr.nodeMovers[id]
	return ok
}

// NodeQuantOffset returns node id's own current quantized polar offset triple
// (iTheta, iPhi, iR), for the same external-verification reason as NodeSelfDriven — e.g.
// confirming a real reload lands on the same offset a live edit just persisted.
func (md *MoveDispatch) NodeQuantOffset(id string) (iTheta, iPhi, iR int, ok bool) {
	nm, exists := md.mr.nodeGeoms[id]
	if !exists {
		return 0, 0, 0, false
	}
	return nm.quantOffset.iTheta, nm.quantOffset.iPhi, nm.quantOffset.iR, true
}

// ClearOutBeads empties every one of this node's own outgoing wires directly — the
// direct-call replacement for the removed moveMsgKindBeadClear message-to-self. This
// node's own geometry already drives those wires (Step, above), so it may clear them
// itself with no message needed at all.
func (p *PairNodeSelf) ClearOutBeads() {
	if p == nil || p.geom == nil {
		return
	}
	for _, pw := range p.geom.outWires {
		pw.ClearInFlight()
	}
}
