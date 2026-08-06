// pair_node_self.go — PairNodeSelf, the handle a PAIR-scene node kind (Node1/Node2) uses
// to own its own mover state directly, on its own Update goroutine, instead of through a
// separate nodeMover goroutine (task/pair-node-owns-itself). See MODEL.md and this
// package's node_mover.go for the mover's own doc comments — nothing about what a mover
// IS changes here; only which goroutine drives it, for exactly these two kinds.
//
// THE RING IS UNTOUCHED: every ring node's nodeMover keeps running on its own dedicated
// goroutine (nodeMover.run, launched by mr.start) exactly as before. PairNodeSelf only
// ever wraps a mover whose selfDriven flag a kind explicitly set via
// BuildArgs.ClaimSelfDrive, which mr.start checks to skip launching that second
// goroutine — see selfDriven's own doc comment on nodeMover.
package Wiring

import (
	"context"
)

// PairNodeSelf wraps this node's own *nodeMover so a pair kind's Update loop can drive
// it directly. Every method here must be called ONLY from that node's own goroutine —
// the same single-goroutine-ownership contract nodeMover.run's own doc comment states,
// just satisfied by a different (but still exactly one) caller. Nil-safe throughout:
// BuildArgs.ClaimSelfDrive returns nil on a bare test build with no loader, matching
// every other nil-safe fallback in this package.
type PairNodeSelf struct {
	nm *nodeMover
}

// EmitGeometryOnce sends this node's initial node-geometry frame — the one-time startup
// emit nodeMover.run makes at goroutine start (see its own doc comment), reproduced here
// since this node's own Update loop is that goroutine now.
func (p *PairNodeSelf) EmitGeometryOnce() {
	if p == nil || p.nm == nil {
		return
	}
	if p.nm.tr != nil {
		p.nm.emitGeometry()
	}
}

// Step runs exactly one cycle of this node's own mover work — nodeMover.run's per-cycle
// body (drain every dedicated inbound channel to empty, drive this node's own outgoing
// wires one cycle on the given tick, retry any pending sends, write this node's dedicated
// stream frame), called from the OWNING kind's own goroutine instead of a second one.
// There is no pacing sleep here: the caller's own Update loop already paces itself on its
// own clock (per-goroutine-clock.md), and driving this mover an extra time per caller
// cycle is exactly what "one goroutine, one clock reading" requires — a second sleep here
// would just double-pace the same node.
func (p *PairNodeSelf) Step(ctx context.Context, tick int64) {
	if p == nil || p.nm == nil {
		return
	}
	nm := p.nm
	for {
		progressed := false
		select {
		case msg := <-nm.extIn:
			nm.handle(msg)
			if msg.testDone != nil {
				close(msg.testDone)
			}
			progressed = true
		default:
		}
		for _, ch := range nm.neighborIn {
			select {
			case msg := <-ch:
				nm.handle(msg)
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
	for _, pw := range nm.outWires {
		pw.DriveOneCycle(ctx, tick)
	}
	nm.flushPending()
	nm.writeStreamFrame(nil)
}

// SetTiltIndex applies this node's own new top/normal/bottom tilt-vector index triple
// directly to its own mover state — the direct-call replacement for the removed
// moveMsgKindTiltIndexSync message-to-self (see that constant's retirement note in
// move_msg.go). Same effect as that message's old handle() branch: persist to this
// node's OWN position.json, re-emit, and — PAIR TAB ONLY — reposition this node along
// its own fixed ray per repositionForTiltIndex's model (unchanged; see its own doc
// comment for the exact D formula and the Node2-only/Node1-anchor rule).
func (p *PairNodeSelf) SetTiltIndex(theta, phi, normalTheta, normalPhi, bottomTheta, bottomPhi int32) {
	if p == nil || p.nm == nil {
		return
	}
	nm := p.nm
	nm.topTiltVectorThetaIdx = theta
	nm.topTiltVectorPhiIdx = phi
	nm.normalThetaIdx = normalTheta
	nm.normalPhiIdx = normalPhi
	nm.bottomThetaIdx = bottomTheta
	nm.bottomPhiIdx = bottomPhi
	nm.persistTiltVectorAngle()
	if nm.tr != nil {
		nm.emitGeometry()
	}
	nm.repositionForTiltIndex(theta)
}

// SetReceivedVector applies this node's own last-received vector-channel direction
// directly to its own mover state — the direct-call replacement for the removed
// moveMsgKindReceivedVectorSync message-to-self. Same effect as that message's old
// handle() branch: re-emit so the third drawn arrow picks up the change; nothing here is
// persisted (a channel arrival is transient session state).
func (p *PairNodeSelf) SetReceivedVector(theta, phi int32, set bool) {
	if p == nil || p.nm == nil {
		return
	}
	nm := p.nm
	nm.receivedVectorThetaIdx = theta
	nm.receivedVectorPhiIdx = phi
	nm.receivedVectorSet = set
	if nm.tr != nil {
		nm.emitGeometry()
	}
}

// NodeSelfDriven reports whether node id's own mover is driven by that node's own kind
// goroutine (task/pair-node-owns-itself, ClaimSelfDrive) rather than a separate nodeMover
// goroutine started by mr.start. Exposed for verification: the model's whole point — one
// goroutine, not two, for the same node id — is otherwise invisible from outside this
// package (package main's own headless tests are the only place every kind, Node1/Node2
// included, is registered — see kind_registry_parity_test.go's own doc comment).
func (md *MoveDispatch) NodeSelfDriven(id string) bool {
	nm, ok := md.mr.nodeMovers[id]
	return ok && nm.selfDriven
}

// NodeQuantOffset returns node id's own current quantized polar offset triple
// (iTheta, iPhi, iR), for the same external-verification reason as NodeSelfDriven — e.g.
// confirming a real reload lands on the same offset a live edit just persisted.
func (md *MoveDispatch) NodeQuantOffset(id string) (iTheta, iPhi, iR int, ok bool) {
	nm, exists := md.mr.nodeMovers[id]
	if !exists {
		return 0, 0, 0, false
	}
	return nm.quantOffset.iTheta, nm.quantOffset.iPhi, nm.quantOffset.iR, true
}

// ClearOutBeads empties every one of this node's own outgoing wires directly — the
// direct-call replacement for the removed moveMsgKindBeadClear message-to-self. This
// node's own mover already drives those wires (Step, above), so it may clear them
// itself with no message needed at all.
func (p *PairNodeSelf) ClearOutBeads() {
	if p == nil || p.nm == nil {
		return
	}
	for _, pw := range p.nm.outWires {
		pw.ClearInFlight()
	}
}
