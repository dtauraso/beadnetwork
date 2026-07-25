// node_move.go — decentralized node-move handling.
//
// A node-move is NOT handled by a central coordinator. Instead the loader wires each
// mover's dedicated inbound channels: there is no shared many-to-one inbox — every pair of movers that talk (a node and its
// incident edge, or two nodes joined by an edge) gets its OWN directed channel, plus every
// mover gets one dedicated "external" channel for the stdin/gesture goroutine's rare direct
// entries. The stdin reader's whole job for a move is to write each entry onto the ONE
// channel that entry addresses. No recompute, no topology logic lives in the reader.
//
// Two kinds of mover own the work, each in its own goroutine (MODEL.md: each node
// and each wire is a goroutine; geometry emission is per-goroutine —
// memory/feedback_per_goroutine_bridge.md):
//
//   - nodeMover: owns ONE node's geometry. On a move for itself it updates its own
//     held position and re-emits its own node-geometry (emitNodeGeometry).
//   - edgeMover: owns ONE edge. It holds BOTH endpoint nodeGeoms (set at load). On
//     a move of either endpoint it updates the matching endpoint, recomputes its
//     OWN segment + arc (segmentBetweenPorts/arcLengthBetweenPorts), writes them
//     onto the source Out (next placement reads them), revises any in-flight bead's
//     remaining travel on the dest wire (ReviseInFlightGeometry, fraction-preserving),
//     and emits its OWN edge geometry (tr.Geometry).
//
// This reproduces, per-goroutine, exactly what the old central applyNodeMove did in
// one stdin goroutine: same node-geometry emit, same per-edge segment/arc recompute,
// same in-flight revision, same edge-geometry emit.

package Wiring

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"

	T "github.com/dtauraso/wirefold/Trace"
)

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
)

// centerSnap is an immutable snapshot of a node's position published by the nodeMover
// via an atomic.Pointer so readers on other goroutines (stdin reader, etc.) can observe
// the current center without touching the mover's live geom.
type centerSnap struct {
	c     vec3  // world center — cartesian, published for the emit/input boundaries only
	p     polar // scene polar (r,θ,φ about the scene center) — the polar source of truth for geometry math
	reach float64
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
// node_move_test.go's `deliver`). It is needed because an edgeMover publishes no atomic
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
	// the receiver repositions itself relative to.
	SenderID string
	// SnapC (Kind == "neighborSetC"): the new quantized edge length (whole ticks of
	// the receiver's own step constant) to write onto the receiver's own LocalPolar
	// record to SenderID.
	SnapC int
	// DeltaA/DeltaB/DeltaC (Kind == "neighborSetC"): the DRAGGED node's (SenderID's)
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

// MoveDispatch is the pure registry built at load that owns every mover and wires their
// dedicated channels together — there
// is no shared dispatch map anymore; md.mr.nodeMovers/md.mr.edgeMovers themselves are the
// directories a mover's resolveDest closure and the external-entry helpers below look up.
// It also retains the per-edge source Outs so out-of-package test/verifier callers can
// read an edge's loaded geometry (EdgeOut) without going through a central coordinator.
type MoveDispatch struct {
	// mr owns the nodeMover/edgeMover directories + edgeOut (mover_registry.go).
	// MoveDispatch's public Bind/Start/EdgeOut methods are thin delegators so the
	// external API is unchanged.
	mr moverRegistry
	// gs owns every node/edge's load-time seed geometry (geom_seeds.go). MoveDispatch's
	// public NodeSeeds/EdgeSeeds methods are thin delegators so the external API is
	// unchanged.
	gs geomSeeds
	// tr is the trace sink (retained for trace emission; diagnostic breadcrumbs removed).
	tr *T.Trace
	// persist groups the six debounced disk persisters (move_persist.go), each nil until
	// armed by EnableViewpointPersist / EnableEditPersist after the startup seed. Grouped
	// the same way ui.vp/ui.ov/ui.gest are, so a bare test-constructed MoveDispatch only
	// has to reason about one zero-value sub-struct instead of six loose nilable fields.
	persist persisters
	// sw owns the fd-wiring for the per-node interior stream and the dedicated VIEW
	// stream (stream_wiring.go). MoveDispatch's public SetEdgeStreams/SetNodeStreams
	// methods stay as thin delegators so the external API is unchanged; view_stream.go's
	// emitViewFrame reads md.sw.viewOut/viewBuildFrame/viewTick directly (it also reads
	// md.ui.vp/md.ui.ov/md.ui.sceneSphere/md.ui.abcDragCount, which are NOT part of this extraction).
	sw streamWiring
	// ui owns the camera/overlay/gesture/selection/abc-drag UI state (ui_state.go).
	// MoveDispatch's public setSelectionUI/setHoverUI/sendEdgeSelect/resetAbcDrag stay as
	// thin delegators so the external API is unchanged.
	ui uiState
	// lq owns the quantized double-link local-polar move math (quantized_move.go):
	// quantizedLayout (gates the quantized absolute-scene-polar snap — every node is a
	// root, measured/derived about the scene center only) and layoutHolders (resolves a
	// node id to the *LayoutHolder embedded in that node's built struct — the ONLY
	// route from the drag path (RootMove) to each node's own LayoutHolder; MoveDispatch
	// does not own or copy LocalPolars itself, it just routes the update to the owning
	// node). MoveDispatch's public RootMove stays a thin delegator; the several
	// package-private quantized_move.go methods also stay thin delegators so their
	// existing in-package call sites (tests, node_move.go, gesture.go) are unchanged.
	lq layoutQuantizer
	// msgTap is a TEST-ONLY observability seam: when non-nil, sendMove invokes it with
	// every (destID, msg) it routes, BEFORE the send. nil in production — production code
	// never calls SetMsgTap, so msgTap.Load() is always nil there (one atomic load, no
	// lock, no allocation). This exists only so tests can assert the message trace
	// between movers (e.g. the neighborSetC single-hop propagation) is exactly what's
	// expected. It is pure observation — it never authors domain state or changes
	// routing. Stored as an atomic.Pointer (not a plain field) because sendMove is the
	// one chokepoint every mover goroutine calls concurrently.
	msgTap atomic.Pointer[func(destID string, msg moveMsg)]
	// ctx is the process-lifetime context, captured in Start. sendMove (the bare
	// blocking directory send, used by external entry points like RootMove with
	// no owning mover goroutine to thread a ctx from) selects on ctx.Done() so a
	// send into a full inbox aborts on shutdown instead of leaking the calling
	// goroutine forever. nil in tests that build a bare MoveDispatch without
	// calling Start — sendMove treats a nil ctx as "no cancellation available"
	// and falls back to the plain blocking send (matches prior test behavior).
	ctx context.Context

	// rowTables owns the four row-identity tables (row_tables.go). MoveDispatch's
	// public Lookup*/…RowFor methods below are thin delegators to it so the external
	// API is unchanged.
	rt rowTables
}

// LookupNodeRow resolves a numeric buffer NODE-ROW index to its node id via the row table
// built at load. ok=false for an out-of-range row. This is the node analogue of
// LookupPortRow/LookupEdgeRow: a numeric node-row hit (the node InstancedMesh instanceId
// == its buffer node row) resolves back to the node id here in Go, so the numeric buffer
// carries no node id strings and the webview forwards only the row.
func (md *MoveDispatch) LookupNodeRow(row int) (nodeID string, ok bool) {
	return md.rt.lookupNodeRow(row)
}

// NodeRowFor resolves nodeID to its buffer NODE-ROW index via the row table built at
// load — the REVERSE of LookupNodeRow. Used by a dedicated nodeMover goroutine
// (memory/feedback_no_single_writer_bridge.md) to resolve its own layout-link's dst node
// row for its per-node stream frame. ok=false when nodeID is not a registered node.
func (md *MoveDispatch) NodeRowFor(nodeID string) (int32, bool) {
	return md.rt.nodeRowFor(nodeID)
}

// LookupEdgeRow resolves a numeric buffer EDGE-ROW index to its edge label via the row
// table built at load. ok=false for an out-of-range row. This is the row→edge resolution
// the gesture FSM uses to mark the Go-owned edge selection — the numeric buffer carries
// no edge label strings.
func (md *MoveDispatch) LookupEdgeRow(row int) (label string, ok bool) {
	return md.rt.lookupEdgeRow(row)
}

// LayoutLinkPairs returns every LAYOUT double-link pair (id, to), one per unordered pair
// (id is always the alphabetically-first side — mirrors loader.go's emitLayoutLinks own
// de-dup rule), by walking each nodeMover's own layoutLinkTos (seeded once at load,
// static since — see its doc comment in node_mover.go). This is the SAME set
// emitLayoutLinks streams via tr.LayoutLink for the old central accumulator's LayoutLink BLOCK
// (still the sole source of that block; unaffected by this method), reconstructed here
// so main.go can ALSO emit each pair once as a view-owner VIEW-frame event (Step C,
// memory/feedback_no_single_writer_bridge.md — LayoutLink is load-time-once, like SceneSphere, so it has no
// live per-goroutine owner to decentralize onto; this seed-once emission is its
// decentralized counterpart). Order is not guaranteed to match emitLayoutLinks' — the
// .probe log is a multiset of events, per memory/feedback_no_single_writer_bridge.md's own doc comment.
func (md *MoveDispatch) LayoutLinkPairs() [][2]string {
	var out [][2]string
	for id, nm := range md.mr.nodeMovers {
		for _, to := range nm.layoutLinkTos {
			out = append(out, [2]string{id, to})
		}
	}
	return out
}

// EdgeRowForPair resolves the buffer edge-row index of the bead edge connecting node ids
// a/b (in either direction) via the edge-endpoint table built at load — safe to call from
// any goroutine (the table is read-only after construction). Used by a dedicated nodeMover
// goroutine to resolve its own layout-link's edge row for its per-node stream frame.
// ok=false when no such edge exists.
func (md *MoveDispatch) EdgeRowForPair(a, b string) (int32, bool) {
	return md.rt.edgeRowForPair(a, b)
}

// LookupPortRow resolves a numeric buffer PORT-ROW index to its (node, port, isInput)
// identity via the port-row table built at load. ok=false for an out-of-range row. This is
// the row→(node,port) resolution the gesture FSM uses for wiring/handhold — the numeric
// buffer carries no port strings.
func (md *MoveDispatch) LookupPortRow(row int) (node, port string, isInput, ok bool) {
	return md.rt.lookupPortRow(row)
}

// PortRowFor resolves (node, port, isInput) to its buffer PORT-ROW index via the port-row
// table built at load — the REVERSE of LookupPortRow. Used by a dedicated edgeMover
// goroutine to resolve its own SrcPortRow/DstPortRow for its per-edge stream frame.
// ok=false when no port matches.
func (md *MoveDispatch) PortRowFor(node, port string, isInput bool) (int32, bool) {
	return md.rt.portRowFor(node, port, isInput)
}

// SetMsgTap installs (or clears, with nil) the test-only message-trace hook. Test-only —
// production code never calls this.
func (md *MoveDispatch) SetMsgTap(tap func(destID string, msg moveMsg)) {
	if tap == nil {
		md.msgTap.Store(nil)
		return
	}
	md.msgTap.Store(&tap)
}

// SetEdgeStreams wires every edgeMover to ITS OWN dedicated fd — the per-edge stream
// (memory/feedback_no_single_writer_bridge.md): fd = baseFd + row, where row is the
// STABLE edge-seed order (md.gs.edgeSeeds, the same spec order the Edge
// block uses — see main.go's md.EdgeSeeds() seed loop). portRowFor/buildFrame are
// injected funcs (not a Buffer import) so this package stays Buffer-independent,
// matching PortRowResolver/EdgeRowResolver's existing pattern: portRowFor resolves
// (node,port,isInput) to a buffer PORT-ROW index (mirroring the old central accumulator's PortRowFor), and
// buildFrame packs the combined per-edge frame bytes (Buffer.BuildEdgeStreamFrame).
// Edge selection is NOT injected: each edgeMover owns its OWN selected bit, set via a
// moveMsgKindSelect message on its extIn (md.sendEdgeSelect), not a lookup. Call once at
// startup after LoadTopology, before Start — mirrors SetPortRowResolver/
// SetEdgeRowResolver's call site in main.go. A missing edgeMover for a seed row (should
// not happen) is skipped rather than panicking.
func (md *MoveDispatch) SetEdgeStreams(
	baseFd int,
	portRowFor func(node, port string, isInput bool) (int32, bool),
	nodeRowFor func(id string) (int32, bool),
	buildFrame func(tick uint32, srcPortRow, dstPortRow int32, selected uint8, label string, beadVal []int32, beadX, beadY, beadZ []float32, events []RowEvent) []byte,
) {
	md.sw.setEdgeStreams(md.gs.edgeSeeds, md.mr.edgeMovers, baseFd, portRowFor, nodeRowFor, buildFrame)
}

// SetNodeStreams wires every nodeMover to ITS OWN dedicated node-fd (geometry+ports+
// label), AND wires the interiorOuts directory + buildInteriorFrame func every node's own
// Update-loop closures (builders.go's injectClosures) look up for its own dedicated
// interior-fd — the two emitting goroutines per node (memory/feedback_no_single_writer_bridge.md).
// nodeBase/interiorBase are the two fd ranges' base fds; row is the STABLE node-seed
// order (md.gs.nodeSeeds, the same spec order the Node block uses — see
// main.go's md.NodeSeeds() seed loop). nodeRowFor/edgeRowForPair/buildFrame/
// buildInteriorFrame are injected funcs (not a Buffer import), matching SetEdgeStreams'
// existing pattern. Selection/hover/abc-drag/kind are NOT injected lookups: each nodeMover
// owns its OWN selected/hovered/latchedSel/gotDragMsg/dragDelta*/kindID fields, set via
// moveMsgKindSelect/Hover/Latched/AbcReset messages (or, for kindID, once here at
// construction — a node's kind never changes after load, so there is no lookup to
// perform on every emit). Call once at startup after LoadTopology, before Start — mirrors
// SetEdgeStreams' call site in main.go. A missing nodeMover for a seed row (should not
// happen) is skipped rather than panicking.
func (md *MoveDispatch) SetNodeStreams(
	nodeBase, interiorBase int,
	nodeRowFor func(id string) (int32, bool),
	edgeRowForPair func(a, b string) (int32, bool),
	buildFrame func(tick uint32, nodeRow int32, cx, cy, cz, radius, sphereR float32, vrx, vry, vrz, frx, fry, frz float32, selected, kindID, hovered, latchedSel, gotDragMsg uint8, dragDeltaA, dragDeltaB, dragDeltaC int32, label string, portNames []string, portDX, portDY, portDZ, portPX, portPY, portPZ []float32, portIsInput, portHovered []uint8, dstNodeRows, edgeRows []int32, events []RowEvent) []byte,
	buildInteriorFrame func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []RowEvent) []byte,
	kindIDFor func(kind string) uint8,
) {
	md.sw.setNodeStreams(md.gs.nodeSeeds, md.mr.nodeMovers, nodeBase, interiorBase, nodeRowFor, edgeRowForPair, buildFrame, buildInteriorFrame, kindIDFor)
}

// NodeGeomSeed is one node's load-time seed geometry, exported in spec order and consumed
// by main.go's pre-launch tr.NodeGeometry loop (see the row-seeding comment in main.go).
// Ports are the SAME aimed-vs-static port geometry (buildPortGeoms/aimedPortPosDir,
// builders.go) the node's own live emit later produces — computed here from the same
// load-time geoms map, since every node's center is already known at load (buildPartnerCenterFn
// resolves partner centers straight off geoms, no goroutine needed). main.go copies these
// fields into the tr.NodeGeometry call (which additionally resolves the numeric KindID,
// since that table lives in Buffer).
type NodeGeomSeed struct {
	ID, Label, Kind              string
	CX, CY, CZ, Radius, SphereR  float64
	Ports                        []T.PortGeom
	VRX, VRY, VRZ, FRX, FRY, FRZ float64
}

// EdgeGeomSeed is one edge's load-time topology AND its real segment endpoints — the same
// edgeSegment(srcGeom, dstGeom, srcH, dstH) computation the edge's own live recomputeGeometry
// (node_move.go) uses, evaluated here against the load-time geoms so the seed row is never a
// degenerate 0,0,0→0,0,0 segment.
type EdgeGeomSeed struct {
	Label, SrcNode, DstNode string
	// SrcPort/DstPort are the edge's source (output) and dest (input) port NAMES —
	// the edgeMover's own dedicated stream frame resolves these to buffer PORT-ROW
	// indices (see Trace.Geometry).
	SrcPort, DstPort       string
	SX, SY, SZ, EX, EY, EZ float64
}

// NodeSeeds returns every node's load-time seed geometry in SPEC ORDER (see
// geomSeeds.nodeSeeds). Call after LoadTopology returns, before launching any node
// goroutine, and stream each entry via tr.NodeGeometry (main.go). Thin delegator to
// md.gs (geom_seeds.go).
func (md *MoveDispatch) NodeSeeds() []NodeGeomSeed { return md.gs.nodeSeedsFn() }

// loadTimeCenters returns the node-id → LOAD-TIME world center map. Thin delegator to
// md.gs (geom_seeds.go).
func (md *MoveDispatch) loadTimeCenters() map[string]vec3 { return md.gs.loadTimeCenters() }

// EdgeSeeds returns every edge's load-time seed topology (with real endpoint geometry) in
// SPEC ORDER. Call alongside NodeSeeds; stream each entry via tr.Geometry (main.go). Thin
// delegator to md.gs (geom_seeds.go).
func (md *MoveDispatch) EdgeSeeds() []EdgeGeomSeed { return md.gs.edgeSeedsFn() }

// newMoveDispatch builds the registry from per-node geometry and per-edge endpoints.
// It creates one nodeMover per node and one edgeMover per edge, registering each under
// its key (node id / edge id) in md.mr.nodeMovers/md.mr.edgeMovers, and wires the dedicated
// directed channels between adjacent movers. Outs and dest wires are bound later by Bind once node
// construction has populated them. nodeOrder/edgeOrder are the
// SPEC order (deterministic directory-sorted order, not map iteration order) used to
// build md.gs.nodeSeeds/edgeSeeds for buffer row seeding.
//
// speedSinks, when non-nil, is the loader's build-wide accumulator
// (buildCtx.speedSinks): each nodeMover AND each edgeMover created below gets its own
// fresh buffered-1 speed channel (per-goroutine-clock.md "Delivery" — every
// clock-owning goroutine must not be left behind), and that channel's SEND end is
// appended here.
// nil in test call sites that construct a MoveDispatch directly with no
// loader — those edgeMovers then simply have no speed channel to poll.
func newMoveDispatch(geoms map[string]nodeGeom, edgeEndpoints map[string]EdgeEndpoints, tr *T.Trace, nodeOrder, edgeOrder []string, clk Clock, speedSinks *[]chan float64) *MoveDispatch {
	// nil order (test call sites that don't care about seed order) falls back to sorted
	// map keys — still deterministic, just not necessarily spec order.
	if nodeOrder == nil {
		nodeOrder = make([]string, 0, len(geoms))
		for id := range geoms {
			nodeOrder = append(nodeOrder, id)
		}
		sort.Strings(nodeOrder)
	}
	if edgeOrder == nil {
		edgeOrder = make([]string, 0, len(edgeEndpoints))
		for label := range edgeEndpoints {
			edgeOrder = append(edgeOrder, label)
		}
		sort.Strings(edgeOrder)
	}
	md := &MoveDispatch{
		tr: tr,
	}
	md.mr.nodeMovers = map[string]*nodeMover{}
	md.mr.edgeMovers = map[string]*edgeMover{}
	md.mr.edgeOut = map[string]*Out{}
	md.lq.layoutHolders = map[string]*LayoutHolder{}
	md.ui.ov = defaultOverlayState()
	// Static partner-center lookup for the seed pass: every node's center is already known
	// off the load-time geoms map (no goroutine/atomic-snap needed), so this is the SAME
	// buildPartnerCenterFn the dynamic movers use below, just closed over geoms directly
	// instead of md.mr.nodeMovers' atomic snaps. Kept per-node (not shared) to match
	// buildPartnerCenterFn's (nodeID, edgeEndpoints, centerOf) shape.
	seedPartnerCenter := func(nodeID string) partnerCenterFn {
		return buildPartnerCenterFn(nodeID, edgeEndpoints, func(otherID string) vec3 {
			if g, ok := geoms[otherID]; ok {
				return nodeWorldPos(g)
			}
			return vec3{}
		})
	}
	md.gs.nodeSeeds = make([]NodeGeomSeed, 0, len(nodeOrder))
	for _, id := range nodeOrder {
		g, ok := geoms[id]
		if !ok {
			continue
		}
		label := g.Label
		if label == "" {
			label = id
		}
		var cx, cy, cz float64
		if g.HasPos {
			c := nodeWorldPos(g)
			cx, cy, cz = c.X, c.Y, c.Z
		}
		ports := buildPortGeoms(g, aimedPortPosDir(g, seedPartnerCenter(id)))
		md.gs.nodeSeeds = append(md.gs.nodeSeeds, NodeGeomSeed{
			ID: id, Label: label, Kind: g.Kind,
			CX: cx, CY: cy, CZ: cz,
			Radius: nodeRadius(g.Kind), SphereR: effectiveRadius(g),
			Ports: ports,
			VRX:   verticalRingNormalX, VRY: verticalRingNormalY, VRZ: verticalRingNormalZ,
			FRX: flatRingNormalX, FRY: flatRingNormalY, FRZ: flatRingNormalZ,
		})
	}
	md.gs.edgeSeeds = make([]EdgeGeomSeed, 0, len(edgeOrder))
	for _, label := range edgeOrder {
		ep, ok := edgeEndpoints[label]
		if !ok {
			continue
		}
		// Real endpoint geometry: the same edgeSegment computation recomputeGeometry
		// (below) uses on every live move, evaluated once here against the load-time
		// geoms so the seed row is never a degenerate 0,0,0->0,0,0 segment.
		var sx, sy, sz, ex, ey, ez float64
		if srcG, ok := geoms[ep.Source]; ok {
			if dstG, ok := geoms[ep.Target]; ok {
				seg := edgeSegment(srcG, dstG, ep.SourceHandle, ep.TargetHandle)
				sx, sy, sz = seg.Start.X, seg.Start.Y, seg.Start.Z
				ex, ey, ez = seg.End.X, seg.End.Y, seg.End.Z
			}
		}
		md.gs.edgeSeeds = append(md.gs.edgeSeeds, EdgeGeomSeed{
			Label: label, SrcNode: ep.Source, DstNode: ep.Target,
			SrcPort: ep.SourceHandle, DstPort: ep.TargetHandle,
			SX: sx, SY: sy, SZ: sz, EX: ex, EY: ey, EZ: ez,
		})
	}
	for id, g := range geoms {
		nm := newNodeMover(id, g, tr, clk)
		if speedSinks != nil {
			nodeSpeedCh := make(chan float64, 1)
			nm.speedCh = nodeSpeedCh
			*speedSinks = append(*speedSinks, nodeSpeedCh)
		}
		// resolveDest resolves the ONE dedicated directed channel FROM this node
		// (selfID, captured below) TO destID: another node's own neighborIn[selfID]
		// slot, or an incident edge's srcIn/dstIn depending on which endpoint this
		// node is. There is no shared dispatch map to look up — md.mr.nodeMovers/md.mr.edgeMovers are the
		// read-only directories, safe to read from any goroutine once construction
		// finishes.
		selfID := id
		nm.resolveDest = func(destID string) (chan moveMsg, bool) {
			if em, ok := md.mr.edgeMovers[destID]; ok {
				switch selfID {
				case em.srcID:
					return em.srcIn, true
				case em.dstID:
					return em.dstIn, true
				}
				return nil, false
			}
			if other, ok := md.mr.nodeMovers[destID]; ok {
				if ch, ok := other.neighborIn[selfID]; ok {
					return ch, true
				}
			}
			return nil, false
		}
		nm.sendMove = md.enqueueFor(nm)
		nm.centerOf = md.centerOfNode
		ownMover := nm
		nm.commitLocal = func(_ string, newPos vec3) { md.commitNodeMoveLocal(ownMover, newPos) }
		nm.neighborSetC = md.neighborSetCRequantize
		// Go 1.22+ loop semantics give each iteration its own id, so this closure safely
		// captures THIS iteration's id (no shared-variable capture bug).
		nm.layoutHolderFn = func() *LayoutHolder { return md.lq.layoutHolders[id] }
		md.mr.nodeMovers[id] = nm
	}
	for edgeID, ep := range edgeEndpoints {
		em := newEdgeMover(ep, edgeID, geoms[ep.Source], geoms[ep.Target], tr, clk)
		if speedSinks != nil {
			edgeSpeedCh := make(chan float64, 1)
			em.speedCh = edgeSpeedCh
			*speedSinks = append(*speedSinks, edgeSpeedCh)
		}
		md.mr.edgeMovers[edgeID] = em
		// This edge's two nodes each get a dedicated channel TO this edge (already
		// created above, srcIn/dstIn) — and each other's own dedicated channel for
		// node-to-node traffic (neighborIn, the plain-neighbor/partner-reemit fan):
		// two directed channels per ordered pair, never a shared inbox.
		if srcNM, ok := md.mr.nodeMovers[ep.Source]; ok {
			if dstNM, ok := md.mr.nodeMovers[ep.Target]; ok {
				if _, exists := dstNM.neighborIn[ep.Source]; !exists {
					dstNM.neighborIn[ep.Source] = make(chan moveMsg, 8)
				}
				if _, exists := srcNM.neighborIn[ep.Target]; !exists {
					srcNM.neighborIn[ep.Target] = make(chan moveMsg, 8)
				}
			}
		}
	}
	// Wire each nodeMover's aimed-port lookup: for (port,isInput) on nodeID, find its one
	// edge (edgeEndpoints) and read the partner's CURRENT center off the partner
	// nodeMover's atomic snap (md.mr.nodeMovers is a read-only map after this point — only the
	// individual *nodeMover.snap is read cross-goroutine, via the existing atomic pattern).
	for id, nm := range md.mr.nodeMovers {
		nm.partnerCenter = buildPartnerCenterFn(id, edgeEndpoints, func(otherID string) vec3 {
			other, ok := md.mr.nodeMovers[otherID]
			if !ok {
				return vec3{}
			}
			if s := other.snap.Load(); s != nil {
				return s.c
			}
			return vec3{}
		})
	}
	// Give every nodeMover the ids of its OWN incident edges, so a lock-driven move can
	// notify its edges via sendMove (resolveDest's per-pair channel lookup) — no cached
	// channel slice.
	for id, nm := range md.mr.nodeMovers {
		for edgeID, em := range md.mr.edgeMovers {
			if em.srcID == id || em.dstID == id {
				nm.edgeIDs = append(nm.edgeIDs, edgeID)
			}
		}
	}
	// Row-identity tables: built ONCE here, from nodeSeeds/edgeSeeds (already in stable
	// spec order above) — see buildRowTables' doc comment for why this is a load-time
	// constant, not a discovery log.
	md.rt.buildRowTables(md.gs.nodeSeeds, md.gs.edgeSeeds)
	return md
}

// Bind wires the per-edge source Outs and dest wires into each edgeMover. Thin delegator
// to md.mr (mover_registry.go).
func (md *MoveDispatch) Bind(outSink map[string]*Out, slotReg SlotRegistry) {
	md.mr.bind(outSink, slotReg)
}

// Start launches every mover's goroutine. Thin delegator to md.mr (mover_registry.go);
// md.ctx is set here (not part of moverRegistry — see sendMove/enqueueFor's doc
// comments for why sendMove needs it threaded through).
func (md *MoveDispatch) Start(ctx context.Context) *sync.WaitGroup {
	md.ctx = ctx
	return md.mr.start(ctx)
}

// EdgeOut returns the source *Out bound to the given edge label, or nil if unknown.
// Thin delegator to md.mr (mover_registry.go).
func (md *MoveDispatch) EdgeOut(edgeID string) *Out {
	return md.mr.edgeOutFor(edgeID)
}

// centerOfNode returns the current world center for a node id. Thin delegator to md.mr
// (mover_registry.go).
func (md *MoveDispatch) centerOfNode(id string) (vec3, bool) {
	return md.mr.centerOfNode(id)
}

// sendMove routes one moveMsg to a node's dedicated external-entry channel (extIn).
// Thin delegator to md.mr (mover_registry.go); md.msgTap/md.ctx are threaded through
// (not part of moverRegistry).
func (md *MoveDispatch) sendMove(id string, msg moveMsg) {
	md.mr.sendMove(&md.msgTap, md.ctx, id, msg)
}

// enqueueFor returns nm's own non-blocking send function. Thin delegator to md.mr
// (mover_registry.go); md.msgTap is threaded through (not part of moverRegistry).
func (md *MoveDispatch) enqueueFor(nm *nodeMover) func(id string, msg moveMsg) {
	return md.mr.enqueueFor(&md.msgTap, nm)
}

// NOTE (neighborSetC drop history): moveMsgKindNeighborSetC (requantizeLocalPolars'
// per-neighbor fan, quantized_move.go) used to route through a non-blocking
// sendMoveLossy — "receiver is mid-cascade and will self-requantize on its own next
// commit, so a full-inbox drop is safe" was the reasoning, and blocking was avoided to
// dodge a mutual-adjacency deadlock (two nodes each trying to set-c the other while
// both inboxes are full and neither is draining). Measuring it under the SAME
// concurrent mutually-adjacent drag flood TestMutuallyAdjacentDragFloodNoDeadlock
// drives (TestNeighborSetCDropReachability) showed sendMoveLossy dropping ~98% of
// NeighborSetC sends (9417/9600 in one run) — the drop path was not a rare backstop,
// it was silently discarding almost every message. NeighborSetC is now routed through
// the SENDING node's own retry queue (nm.sendMove in requantizeLocalPolars,
// see nodeMover.pending), the same decoupling every other per-commit fan in this file
// already uses (fanEdgesAndPartners) — it gets the same deadlock-avoidance property (the
// send never blocks the handler goroutine) without ever dropping: an item that can't be
// delivered right now stays with the sender and is retried on the sender's own next
// loop cycle instead of being handed to a separate sender goroutine or dropped.
// TestMutuallyAdjacentDragFloodNoDeadlock still passes with this change, so the deadlock
// risk sendMoveLossy was guarding against does not require a lossy send.

// setSelectionUI sets the Go-owned selection (node XOR edge, exclusive). Thin delegator
// to md.ui (ui_state.go).
func (md *MoveDispatch) setSelectionUI(node, edge string) {
	md.ui.setSelectionUI(md.mr.edgeMovers, md.ctx, md.sendMove, node, edge)
}

// setHoverUI sets the Go-owned hover state and MESSAGES the affected node(s). Thin
// delegator to md.ui (ui_state.go).
func (md *MoveDispatch) setHoverUI(node, port string, isInput bool) {
	md.ui.setHoverUI(md.sendMove, node, port, isInput)
}

// resetAbcDrag re-scopes the recipient SET to the drag about to start. Thin delegator to
// md.ui (ui_state.go).
func (md *MoveDispatch) resetAbcDrag() {
	md.ui.resetAbcDrag(md.mr.nodeMovers, md.sendMove)
}

// NodeKind returns the kind string for the given node id, or "" if unknown.
// Used by applyEdit to resolve the node's kind when snapping a port-anchor
// world-space direction to the nearest ring-anchor index. Called from the
// gesture/stdin-reader goroutine (gesture.go:164, :653), which is NOT the
// nodeMover's own goroutine — this is the ONE genuine cross-goroutine read of
// nm.geom.
//
// Kind lives on nm.geom's embedded nodeIdentity (port_geometry.go), a type carrying
// only the fields the loader sets once at construction and that no handler
// (applyCenter, setPortAnchorId, emitGeometry) ever writes again — grepped clean of
// any write to nodeIdentity's fields outside the load-time literal. That split makes
// this safe by CONSTRUCTION rather than by coincidence of which byte ranges a
// particular access happens to touch: identity fields are not merely
// unwritten-in-practice today, they are not reachable from any writer's
// field-assignment at all, in a different embedded struct from the mutable
// ScenePolar/HasPos/ReachR/Inputs/Outputs applyCenter and setPortAnchorId do write.
// TestNodeKindConcurrentWithApplyCenterUnderRace exercises this concurrently under
// -race as a regression check on the split holding.
func (md *MoveDispatch) NodeKind(nodeID string) string {
	if nm, ok := md.mr.nodeMovers[nodeID]; ok {
		return nm.geom.Kind
	}
	return ""
}

// Quantized double-link local-polar move math (quantized_move.go): thin delegators to
// md.lq so their existing in-package call sites (tests, node_move.go, gesture.go) are
// unchanged.
func (md *MoveDispatch) heldCenters() map[string]vec3 { return md.lq.heldCenters(md) }
func (md *MoveDispatch) requantizePoleTraced(lh *LayoutHolder, updates map[string]vec3) dir {
	return md.lq.requantizePoleTraced(lh, updates)
}
func (md *MoveDispatch) neighborSetCRequantize(selfID, fromID string, selfCenter, fromCenter vec3, deltaA, deltaB, deltaC int) {
	md.lq.neighborSetCRequantize(md, selfID, fromID, selfCenter, fromCenter, deltaA, deltaB, deltaC)
}
func (md *MoveDispatch) commitNodeMoveLocal(nm *nodeMover, newPos vec3) {
	md.lq.commitNodeMoveLocal(md, nm, newPos)
}

// RootMove handles a node-drag under the flat absolute scene-polar layout. Thin
// delegator to md.lq (quantized_move.go).
func (md *MoveDispatch) RootMove(nodeID string, target vec3) bool {
	return md.lq.RootMove(md, nodeID, target)
}

// Overlay-visibility API (MoveDispatch delegators), the overlayState methods, the
// overlayToggles table, defaultOverlayState, and the stdinGuideVisPayload mapper are all
// GENERATED into overlay_gen.go from OVERLAY_FLAG_NAMES (tools/gen-node-defs).
