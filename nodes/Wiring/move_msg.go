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
	// moveMsgKindNeighborSetC: the plain-neighbor / general edge-length propagation
	// (requantizeLocalPolars' per-neighbor fan). A dragged node X sends EVERY direct
	// domain neighbor M this SINGLE ASSIGNMENT -- the new quantized edge length SnapC
	// (X's own freshly-requantized c to M) and X's fresh FromCenter. M does NOT
	// re-derive its stored bearing to X: it KEEPS its own persisted
	// QuantITheta/QuantIPhi to SenderID exactly as stored, writes ONLY the new c onto
	// that record, and repositions itself at
	// FromCenter - dir(storedTheta,storedPhi about M's own pole)*newR -- sliding along
	// its existing viewing direction to the new distance, X held fixed. One hop only: M
	// never forwards this to its own neighbors, and this never runs any further cascade
	// (see neighborSetCReposition).
	moveMsgKindNeighborSetC = "neighborSetC"
	// moveMsgKindDragStart arms the dragged node X's OWN drag-anchor snapshot: X's
	// per-neighbor LocalPolar triples AT DRAG START, captured once on X's own goroutine
	// (nodeMover.handle), so every subsequent requantizeLocalPolars call during the same
	// drag reports current-minus-ANCHOR (the drag's running total) instead of
	// current-minus-previous-move-event (which is almost always 0 — RootMove runs on
	// every ~8ms pointer-move, far finer than one quantize step). Sent from gesture.go's
	// gestPending->gestDragging transition (the one place a drag begins), the same edge
	// that already emits tr.AbcDragReset() — see that call site's comment for why it
	// must not live in RootMove. Sent via the BLOCKING md.sendMove: a dropped drag-start
	// would silently leave X's anchor either unset (falling back to the
	// lazy-arm-on-first-commit path below, which is still correct but anchors one commit
	// later than the true drag start) or, worse, STALE from a prior drag if this were the
	// second+ drag on the same node — so this must never be dropped, same as drag/center.
	moveMsgKindDragStart = "dragStart"
	// moveMsgKindDragEnd is the mirror of moveMsgKindDragStart, sent once per drag
	// gesture at drag END (gesture.go's gestPointerUp, which captures g.dragNode
	// BEFORE reset() clears it) so the dragged node's own bead-actor groups
	// (chain_beads.go/bead_actor_bridge.go) settle every bead they woke — PLAN.md
	// "'done dragging' is not optional": every drag must end on this path, including
	// one abandoned mid-gesture, so a bead is never left parked on machine time.
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
	// moveMsgKindAbcReset tells a node to clear its OWN abc-drag recipient bit
	// (gotDragMsg/dragDeltaA/B/C) at the start of a new drag. Broadcast to every node
	// mover from resetAbcDrag.
	moveMsgKindAbcReset = "abcReset"
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
	// existing SenderID/FromCenter fields (same shape moveMsgKindNeighborSetC already
	// carries: sender id + sender's fresh center).
	moveMsgKindNeighborCenter = "neighborCenter"
	// moveMsgKindDeltaForward is the delta-forward full-graph-propagation observability
	// message: a node selfID that just picked up a delta triple (DeltaA/B/C, the
	// ORIGINALLY-dragged node's own quantized-triple change) — either as the direct
	// drag-recipient (moveMsgKindNeighborSetC from the dragged node) or as a forward
	// recipient (a moveMsgKindDeltaForward from a neighbor that already relayed it) —
	// forwards the SAME triple to each of its CASCADE-LINK neighbors (every cascade-link
	// neighbor except whichever one it came from), carrying its OWN id as SenderID (the
	// forwarder). The receiver records GotForwardMsg/ForwardDeltaA-C/ForwardFromRow on
	// its own node stream frame (LATEST delta wins, so it stays in sync with a drag that
	// keeps moving) AND does its OWN relay in turn (nodeMover.forwardDelta), onward to
	// every cascade-link neighbor the relay rule selects. The cascade-link set is STORED
	// per-node file data (nodes/<id>/cascade-edges.json, read at load by loader_tree.go)
	// and now covers the FULL node adjacency, cycle-closing links included — so it is NOT
	// loop-free by construction. What makes propagation terminate is the PER-KIND relay
	// rules in forwardDelta and the moveMsgKindDeltaForward handler: PulseLeft and
	// PulseRight are termini, TimeStart and Pulse route by sender kind, TimeEnd stops.
	// There is still no runtime visit-tracking and no once-per-drag guard — every move
	// re-propagates freely and terminates on those rules (measured: 1-6 forwards per
	// single-node drag). See neighborSetCRequantize's forward step and node_mover.go's
	// forwardDelta. Pure observability throughout: no re-quantize, no move.
	moveMsgKindDeltaForward = "deltaForward"
)

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
	// FromCenter (Kind == "neighborSetC"): the SENDER's (SenderID's) fresh committed
	// world center. The receiver repositions itself along its OWN stored bearing to
	// SenderID at the new distance (see neighborSetCReposition) — receiver-computes.
	FromCenter vec3
	// SenderID (Kind == "neighborSetC"): the id of the mover whose fresh FromCenter
	// the receiver repositions itself relative to. (Kind == "deltaForward"): the id of
	// the FORWARDER (the direct drag-recipient one hop back), not the originally-dragged
	// node — the forward recipient resolves this to ForwardFromRow via NodeRowFor.
	SenderID string
	// SnapC (Kind == "neighborSetC"): the new quantized edge length (whole ticks of
	// the receiver's own step constant) to write onto the receiver's own LocalPolar
	// record to SenderID.
	SnapC int
	// DeltaA/DeltaB/DeltaC (Kind == "neighborSetC" or "deltaForward"): the DRAGGED node's
	// own quantized-triple change (newTriple - oldTriple, integer indices) for ITS edge
	// to this receiver, computed ONCE on SenderID's own goroutine in
	// requantizeLocalPolars. Pure observability payload — the receiver reports it on the
	// in-editor drag-log; it never applies it or recomputes its own position from it
	// (that stays exactly the receiver-computes reposition already in place). Zero if
	// SenderID had no prior stored triple to this receiver to subtract from.
	DeltaA, DeltaB, DeltaC int
	// Target (Kind == "drag"): the raw drag target world position for NodeID's
	// owner-goroutine commit. Every node is a free move, so this is committed as-is.
	Target vec3
	// Bool (Kind == "select"/"hover"/"latched"): the on/off payload for that UI bit —
	// each mover owns its own selected/hovered/latchedSel bit, set here by whichever
	// message it receives on its own extIn.
	Bool bool
	// testDone: see the type comment. Test-only; production leaves it nil.
	testDone chan struct{}
}
