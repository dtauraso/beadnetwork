// move_msg.go — the inter-mover message vocabulary: moveMsgKind* constants and the
// moveMsg type routed between mover goroutines (node_move.go).

package Wiring

// moveMsgKind discriminates moveMsg payloads.
const (
	// The node-move kind ("move", the zero value "") carries no payload and is a
	// no-op in every mover switch, so it has no constant — the switches simply
	// fall through. The remaining kinds each select a distinct payload.
	moveMsgKindAnchor  = "anchor"  // per-port anchor update (drag along the ring)
	moveMsgKindCenter  = "center"  // polar-layout re-propagated world center for one node
	moveMsgKindCenters = "centers" // batched centers for an edge: update both endpoints, recompute ONCE
	// moveMsgKindDrag is a node's own-goroutine drag entry: the drag itself is routed
	// to the dragged node's OWN dedicated extIn channel instead of the stdin reader
	// committing on its behalf. The receiver commits its OWN new position via the owner-goroutine commit
	// path (commitNodeMoveLocal, which publishes its snap SYNCHRONOUSLY via
	// applyCenter). A drag is always a FREE move -- no equal-radii solve, no
	// self-trigger cascade.
	moveMsgKindDrag = "drag"
	// moveMsgKindDragStart arms the dragged node X's OWN bead-actor wake
	// (nodeMover.startBeadDrag, bead_chain.go) — sent from gesture.go's
	// gestPending->gestDragging transition (the one place a drag begins). Sent via the
	// BLOCKING md.sendMove: this must never be dropped, same as drag/center.
	moveMsgKindDragStart = "dragStart"
	// moveMsgKindDragEnd is dragStart's mirror: sent to the dragged node X's own extIn
	// when the gesture FSM's drag concludes (gesture.go's gestPointerUp, on EVERY path a
	// drag can end by — a clean pointer-up after moves, or a pointer-up with no move in
	// between — there is no separate "abandoned" branch, matching
	// BeadWakeGroup.EndDrag's own doc comment). X's own goroutine (handle) responds by
	// calling EndDrag on every one of its own outgoing edges' bead-actor chains
	// (nodeMover.endBeadDrag, bead_chain.go) — clearing every bead's dragging flag with
	// one close per edge, never a per-bead send-loop. Without this kind, "done dragging"
	// would have no sender and a woken bead's mode flag could never clear.
	moveMsgKindDragEnd = "dragEnd"
	// moveMsgKindSelect (nodeMover) / edge-select (edgeMover, same kind) is the
	// gesture goroutine's message telling ONE node/edge to turn its OWN selected bit
	// on or off (Bool). Sent to a node via its extIn (md.sendMove) and to an edge via
	// its extIn (md.sendEdgeSelect) — see setSelectionUI's doc comment.
	moveMsgKindSelect = "select"
	// moveMsgKindHover tells a node to turn its OWN hovered bit on (with the hovered
	// port, if any — Port/IsInput) or off (Bool=false). See setHoverUI's doc comment.
	moveMsgKindHover = "hover"
	// moveMsgKindLatched tells a node to turn its OWN latchedSel bit on or off (Bool).
	// See setSelectionUI's doc comment.
	moveMsgKindLatched = "latched"
	// moveMsgKindNeighborCenter is the PUSH delivery of a moved node's fresh world
	// center to one of its direct neighbors, the neighbor's aimed-port partnerCenter
	// lookup source. Sent from applyCenter (the sole write site of a node's own center)
	// immediately after that node updates its own position, to EVERY node in its own
	// neighborIn key set (one hop only — the receiver never forwards this). The
	// receiver stores the pushed center in its OWN partnerCenters map (nodeMover.handle)
	// and re-emits its own geometry so its aimed ports pick up the fresh partner
	// center — same value, same timing (the FIFO per-destination retry queue delivers
	// this before the existing nil-Center re-emit broadcastToEdgesAndPartners sends
	// right after, so the re-emit always sees the just-pushed center). Reuses the
	// existing SenderID/FromCenter fields.
	moveMsgKindNeighborCenter = "neighborCenter"
	// moveMsgKindTiltVectorAngle tells a node to adjust its OWN vector direction by one
	// TiltVectorAngleStep click (Buffer/layout.go's TopTiltVectorTheta/TopTiltVectorPhi,
	// nodes/Wiring/node_mover.go's topTiltVectorThetaIdx/topTiltVectorPhiIdx): Axis selects "theta" or
	// "phi", Bool is the up(+1)/down(-1) direction — same shape as the distanceGroup
	// arrow-click payload (index + direction, no value on the wire; Go owns the math).
	// Sent to the target node's own extIn via md.sendMove from applyUpdateTiltVector
	// (stdin_reader.go), so the index write + persist + re-emit all run on that node's
	// OWN goroutine.
	moveMsgKindTiltVectorAngle = "tiltVectorAngle"
	// moveMsgKindTiltVectorReset tells a node to return its OWN vector direction to the
	// start position — both indices to 0, the documented default that points the tilt
	// vector at world +y. This is a STOP-AND-RETURN, not a nudge: unlike
	// moveMsgKindTiltVectorAngle, it never places a bead — see applyUpdateTiltVector's
	// (stdin_reader.go) "reset" branch and the RESET button's own doc comment
	// (TiltResetButton.tsx). Sent to the target node's own extIn via md.sendMove from
	// applyUpdateTiltVector for any kind that has NOT claimed BuildArgs.TiltEditIn — the
	// only kinds that have (Node1/Node2) instead route this through their own
	// TiltEditIn/TiltEditMsg.Reset, same split as moveMsgKindTiltVectorAngle.
	moveMsgKindTiltVectorReset = "tiltVectorReset"
	// moveMsgKindTiltIndexSync/moveMsgKindReceivedVectorSync/moveMsgKindBeadClear are
	// GONE (task/pair-node-owns-itself). They used to be the ONE-WAY notifications a
	// pair kind's (Node1/Node2) own goroutine sent to a SEPARATE mover goroutine for
	// the same node id. That separate goroutine no longer exists: a pair node owns its
	// own mover state directly (PairNodeSelf, pair_node_self.go), so what used to be a
	// message to itself is now a plain method call — see
	// PairNodeSelf.SetTiltIndex/SetReceivedVector/ClearOutBeads.
)

// TiltEditMsg is one panel-driven tilt-angle click (TiltVectorAnglePanel), routed to a
// node kind's OWN dedicated channel (BuildArgs.TiltEditIn) instead of its mover, for any
// kind that claims that channel at build time (Node1/Node2 — the only kinds whose own
// goroutine independently owns/decides their tilt index, per the straightening loop's
// firing rule). A kind that never calls TiltEditIn is not registered in
// MoveDispatch.tiltEditIns, so stdin_reader's applyUpdateTiltVector falls back to the old
// mover-owned path (moveMsgKindTiltVectorAngle) for it unchanged.
type TiltEditMsg struct {
	Axis string // "theta" or "phi" — which of the node's own indices to adjust. Ignored when Reset or Start is true.
	Up   bool   // true = +1 step, false = -1 step. Ignored when Reset or Start is true.
	// Start (the START TILT button) begins the vector exchange from whatever angles are
	// currently set: send this node's own outgoingVector on VectorOut and place a bead on
	// Out, exactly what a panel adjust used to do as a side effect ("the kick") — but Start
	// changes NO index itself. It is the opening move, split out of the index-adjust path so
	// a ▲/▼ click moves the tilt by exactly one π/12 step and nothing else. Ignored when
	// Reset is true (Reset wins if both are somehow set).
	Start bool
	// Reset (the RESET button, TiltResetButton.tsx): return BOTH indices to 0 — the
	// documented default, tilt vector pointing at world +y. Unlike an adjust, this places
	// NO bead: it is a stop-and-return, not "the kick" (see package doc comments on
	// Node1/Node2).
	Reset bool
}

// moveMsg is one entry routed to one of a mover's own dedicated channels (there is no
// shared inbox). kind selects the
// payload:
//   - "" or "move": node-move — currently a no-op (polar-layout positions all nodes via "center" messages).
//
// Every PRODUCTION send is fire-and-forget: the sender drops the message onto the
// destination's own channel and returns. No production path observes the receiver finishing — a node does its own
// local work on its own goroutine and drives its own outputs (MODEL.md: no ack, no
// send-gating, no delivery signal).
//
// testDone is the one exception and it is NOT a production mechanism: it exists only so
// a test can block until an async mover has handled a message before asserting (see
// node_move_test.go's `deliver`). It is needed because an edgeMover publishes no
// snapshot a test could safely poll. Production ALWAYS leaves it nil — if you find
// yourself setting it outside a _test.go file, you are reintroducing the ack the model
// forbids; make the receiver's own goroutine do the work instead.
type moveMsg struct {
	Kind   string
	NodeID string
	// Anchor payload (Kind == "anchor"): identify the port whose anchor changed.
	// Port/IsInput name the port on NodeID; AnchorId is the snapped ring-anchor index
	// (Go snaps from the incoming world-space direction; TS never computes the index).
	Port     string
	IsInput  bool
	AnchorId int
	// Center (Kind == "center"): the re-propagated world center for NodeID under the
	// polar layout. Each owning node/edge goroutine writes it onto its held geom
	// and re-emits its own geometry. (RootMove is the decentralized node-to-node
	// message cascade entry; there is no central fan-out step.)
	Center *vec3
	// Centers (Kind == "centers"): batched per-edge re-propagation. Maps node id → new
	// world center for every moved endpoint of THIS edge in a single frame, so an edge
	// whose BOTH endpoints moved updates both and recomputes/emits its segment ONCE
	// instead of once per endpoint message (the node-2 drag duplicate-emit fix).
	Centers map[string]vec3
	// ReachR (Kind == "center"): the re-propagated sphere REACH radius for NodeID (max
	// distance to a surface child under the new centers). The nodeMover writes it onto its
	// held geom so the re-emitted node-geometry streams the correct sphereR during a drag.
	ReachR float64
	// FromCenter (Kind == "neighborCenter"): the SENDER's (SenderID's) fresh committed
	// world center, pushed one hop to every direct neighbor's partnerCenters map.
	FromCenter vec3
	// SenderID (Kind == "neighborCenter"): the id of the mover whose fresh FromCenter
	// this is.
	SenderID string
	// Target (Kind == "drag"): the raw drag target world position for NodeID's
	// owner-goroutine commit. Every node is a free move, so this is committed as-is.
	Target vec3
	// Bool (Kind == "select"/"hover"/"latched"): the on/off payload for that UI bit —
	// each mover owns its own selected/hovered/latchedSel bit, set here by whichever
	// message it receives on its own extIn. Also (Kind == "tiltVectorAngle"): true = up
	// (index+1), false = down (index-1).
	Bool bool
	// Axis (Kind == "tiltVectorAngle"): "theta" or "phi" — which of the node's own
	// topTiltVectorThetaIdx/topTiltVectorPhiIdx to adjust.
	Axis string
	// testDone: see the type comment. Test-only; production leaves it nil.
	testDone chan struct{}
}
