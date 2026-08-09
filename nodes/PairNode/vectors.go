package PairNode

// vectors.go — THE DIRECTIONS THIS NODE COMPUTES, AND WHAT IT SAYS ABOUT THEM.
//
// Everything here is a PURE READ of this node's own two ends on its own ring, plus the two
// one-way reports that mirror those reads into this node's own geometry. Nothing in this file
// decides anything: it holds no rule about where a tilt comes to rest (that is machine.go), it
// turns nothing (that is stepFromVector, node.go), and every function is a link follow or a
// field copy — never trig, never a sum
// (memory/feedback_abc_times_constant_not_rederive.md).

import (
	"github.com/dtauraso/wirefold/nodes/Wiring"
)

// bottomTilt is THIS node's own BOTTOM TILT VECTOR: the state a half turn from its own top,
// so it points out of the node's other side and turns with the top as the top turns. A LINK,
// resolved when the ring was built (ring.go) — no arithmetic, and no trig
// (memory/feedback_abc_times_constant_not_rederive.md). There is no φ any more: a half turn
// in θ alone already negates the direction exactly (see Wiring.HalfTurnThetaIdx's own doc
// comment).
func (n *Node) bottomTilt() Wiring.TiltVectorMsg {
	return Wiring.TiltVectorMsg{ThetaIdx: n.bottomState().idx}
}

// coplanarNormal is THIS node's own coplanar normal: ONE QUARTER TURN past this node's own
// tilt vector, always the same way round — Wiring.PerpendicularThetaIdx (6 steps of
// Wiring.CurveParamTiltVectorAngleStep, 90°) ADDED to the tilt index, and nothing else. Both
// ends of a pair run this same unmodified addition — there is no sign difference between
// them. Index arithmetic, never trig (memory/feedback_abc_times_constant_not_rederive.md).
//
// The wrap is a COMPARISON, not a division: a quarter turn added to an index can carry it at
// most one full turn past the top of the circle, so one test against Wiring.FullTurnThetaIdx
// and one subtraction bring it back. No `/`, no `%`, and therefore no sign convention to get
// wrong at negative indices — which is what the arithmetic this replaced needed floor
// division for.
//
// DO NOT ADD A POLE-CROSSING TERM HERE. One was here, adding a further half turn on odd
// multiples of HalfTurnThetaIdx, on the reasoning that θ measured from world +y makes each
// half turn a pole crossing, and that crossing a pole flips φ by 180° so the normal needs its
// own half turn to keep pointing the same drawn way. That reasoning needs a φ to flip, and
// this model has none: task/drop-tilt-vector-phi removed the φ column, and every consumer
// treats an index as a position on a plain circle — the renderer decodes it as
// (sinθ, cosθ, 0), which has no pole and no fold anywhere. What the term actually did was
// point the normal the OTHER way, t + 18 rather than t + 6, over half the index range,
// including the negative indices a ▼ click reaches first.
func (n *Node) coplanarNormal() Wiring.TiltVectorMsg {
	return Wiring.TiltVectorMsg{ThetaIdx: n.topState().quarter.idx}
}

// syncTiltIndex reports THIS node's current tilt index AND its current coplanar-normal
// index (coplanarNormal above) to this node's own geometry in one call — every call site
// that changes TopTiltThetaIdx must also report the normal, since the normal is
// derived from the tilt and the geometry does not derive it itself (see
// Wiring.PairNodeSelf.SetTiltIndex, and moveMsgKindTiltIndexSync's retirement note in
// move_msg.go for what this used to be). nil-safe, same as every other closure call here.
func (n *Node) syncTiltIndex() {
	if n.SyncTiltIndex == nil {
		return
	}
	norm := n.coplanarNormal()
	bottom := n.bottomTilt()
	n.SyncTiltIndex(n.topState().idx, norm.ThetaIdx, bottom.ThetaIdx)
}

// syncReceivedVector reports THIS node's current received-vector state (ReceivedThetaIdx/
// ReceivedThetaIdx/Set) to this node's own geometry — the third-arrow twin of syncTiltIndex.
// Called by every site that changes those fields, below. nil-safe, same as syncTiltIndex.
func (n *Node) syncReceivedVector() {
	if n.SyncReceivedVector == nil {
		return
	}
	n.SyncReceivedVector(n.ReceivedThetaIdx, n.ReceivedSet)
}

// outgoingVector is what THIS node SENDS on VectorOut: its own coplanarNormal. The message
// on the channel IS the direction this node computed and draws, so the arrow the partner
// draws as its received direction coincides with this node's own normal on screen.
//
// Do not rotate it on the way out. A rotation here has to be undone by the receiver's step
// signs (stepFromVector, below) to leave behaviour unchanged, which would spread one
// convention across two call sites of the same rule instead of leaving it in the one place;
// and rotating by a half turn in particular changes nothing at all about where the pair
// comes to rest, since the bottom tilt is the top plus that same half turn — it only swaps
// which of the receiver's two acute tests fires.
func (n *Node) outgoingVector() Wiring.TiltVectorMsg {
	v := n.coplanarNormal()
	// The lattice the index is on travels with it — see TiltVectorMsg.Points. Without it the
	// partner cannot tell a direction picked before a point-count change from one picked
	// after, and the two mean different angles.
	v.Points = n.ringOf().points
	// WHICH MACHINE THIS NODE IS RUNNING travels with every direction it sends. One end reads
	// the gap when the exchange opens and the other has no way to know what it decided; this is
	// how it finds out, on the first reply, without a message of its own. A node still in the
	// setting mode says TiltMachineNone here — that mode's own choice, not a special case for
	// having none — and the other end ignores it (adoptMachine).
	v.Machine = n.Machine.choice()
	return v
}
