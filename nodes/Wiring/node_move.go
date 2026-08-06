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
	"fmt"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"sort"
	"strconv"
	"sync"

	T "github.com/dtauraso/wirefold/Trace"
)

// MoveDispatch is the pure registry built at load that owns every mover and wires their
// dedicated channels together — there
// is no shared dispatch map anymore; md.mr.nodeGeoms/md.mr.edgeMovers themselves are the
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
	// md.ui.vp/md.ui.ov/md.ui.sceneSphere, which are NOT part of this extraction).
	sw streamWiring
	// ui owns the camera/overlay/gesture/selection/abc-drag UI state (ui_state.go).
	// MoveDispatchs public setSelectionUI/setHoverUI/sendEdgeSelect stay as
	// thin delegators so the external API is unchanged.
	ui uiState
	// lq owns the quantized scene-polar move math (quantized_move.go): quantizedLayout
	// gates the quantized absolute-scene-polar snap — every node is a root,
	// measured/derived about the scene center only, with no per-neighbour stored
	// coordinate (MODEL.md "the polar model"). MoveDispatch's public RootMove stays a
	// thin delegator; the several package-private quantized_move.go methods also stay
	// thin delegators so their existing in-package call sites (tests, node_move.go,
	// gesture.go) are unchanged.
	lq layoutQuantizer
	// scenes owns tab switching: the anchor to persist the selection against, and the
	// quit func whose call the extension host's looping respawn follows (scene_tabs.go).
	// Zero until EnableSceneSwitch arms it, so a bare test-constructed MoveDispatch can
	// never end a process.
	scenes sceneSwitch
	// tapToInstall is a TEST-ONLY observability seam: when SetMsgTap is called (before
	// Start), this is stashed here so any nodeMover constructed AFTER that call (there
	// are none in practice — all nodeMovers are built once in newMoveDispatch, before
	// SetMsgTap could ever run — but this keeps the two code paths symmetric and cheap)
	// also gets the tap installed at construction. The live seam every mover goroutine
	// actually fires is its OWN nm.tap field (node_mover.go), set once per mover by
	// SetMsgTap below — there is no shared/concurrently-read tap anymore: each mover
	// owns and reads only its own copy, on its own goroutine. nil in production —
	// production code never calls SetMsgTap.
	tapToInstall func(destID string, msg moveMsg)
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

	// tiltEditIns holds, for each node id whose OWN kind claimed BuildArgs.TiltEditIn at
	// build time (Node1/Node2 — the only kinds that own their tilt index independently),
	// that node's dedicated inbound channel for a panel-driven tilt-angle click. Written
	// ONCE per entry, on the single-threaded build path (buildNodes, via
	// BuildArgs.TiltEditIn), before any goroutine runs — never touched again. A node id
	// with no entry here is a kind that still routes tiltVectorAngle straight to its
	// mover (applyUpdateTiltVector's fallback, stdin_reader.go). Read only by
	// sendTiltEdit, called from the stdin-reader goroutine.
	tiltEditIns map[string]chan TiltEditMsg
}

// finalizeActors builds the ring's nodeMover actor directory from md.mr.nodeGeoms, AFTER
// every kind's own build func has run (so every BuildArgs.ClaimSelfDrive call has already
// happened — see moverRegistry.selfDriveClaimed's own doc comment). Thin delegator to
// md.mr (mover_registry.go).
func (md *MoveDispatch) finalizeActors(speedSinks *[]chan float64) {
	md.mr.finalizeActors(speedSinks)
}

// newMoveDispatch builds the registry from per-node geometry and per-edge endpoints.
// It creates one nodeMover per node and one edgeMover per edge, registering each under
// its key (node id / edge id) in md.mr.nodeGeoms/md.mr.edgeMovers, and wires the dedicated
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
// rowCount is the buffer's node-row space (topoSpec.RowCount — the largest node id found,
// not the node count): rows 0..rowCount-1, ROW ID = NODE ID - 1. 0 (test call sites that
// don't pass one) falls back to the number of resolved seeds, i.e. no gaps.
func newMoveDispatch(geoms map[string]nodeGeom, edgeEndpoints map[string]EdgeEndpoints, tr *T.Trace, nodeOrder, edgeOrder []string, clk wire.Clock, speedSinks *[]chan float64, rowCount int) (*MoveDispatch, error) {
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
	md.mr.nodeGeoms = map[string]*nodeGeometry{}
	md.mr.edgeMovers = map[string]*edgeMover{}
	md.mr.edgeOut = map[string]*wire.Out{}
	md.mr.centerMirror = map[string]vec3{}
	md.ui.ov = defaultOverlayState()
	md.ui.speed = 1        // default playback multiplier; LoadSpeed overwrites from view/speed.json if present
	md.ui.clockDivisor = 1 // no scaling until LoadSpeed resolves the loaded scene's own divisor
	md.gs.nodeSeeds = make([]NodeGeomSeed, 0, len(nodeOrder))
	for i, id := range nodeOrder {
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
		// ROW ID = NODE ID - 1 — declared by the id, not by position in nodeOrder. Falls
		// back to positional index only for a non-numeric id, which real (loadTree-built)
		// specs never produce (loud load-time error there); this keeps synthetic-id unit
		// tests that construct a MoveDispatch directly working unchanged.
		row := i
		if n, err := strconv.Atoi(id); err == nil {
			row = n - 1
		}
		md.gs.nodeSeeds = append(md.gs.nodeSeeds, NodeGeomSeed{
			ID: id, Label: label, Kind: g.Kind,
			CX: cx, CY: cy, CZ: cz,
			Radius: nodeRadius(g.Kind), SphereR: effectiveRadius(g),
			VRX: verticalRingNormalX, VRY: verticalRingNormalY, VRZ: verticalRingNormalZ,
			FRX: flatRingNormalX, FRY: flatRingNormalY, FRZ: flatRingNormalZ,
			Row: row,
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
		// geoms so the seed row is never a degenerate 0,0,0->0,0,0 segment. A missed
		// lookup means the edge's source or target node id has no geometry — a
		// malformed topology (most commonly a stale edge file left behind after its
		// target node's directory was deleted by hand; in-edges are not indexed, so
		// nothing else catches this) — and must fail the load loudly, never seed a
		// silent 0,0,0->0,0,0 segment indistinguishable from real data.
		srcG, srcOK := geoms[ep.Source]
		if !srcOK {
			return nil, fmt.Errorf("newMoveDispatch: edge %q references missing source node %q (no geometry loaded for it)", label, ep.Source)
		}
		dstG, dstOK := geoms[ep.Target]
		if !dstOK {
			return nil, fmt.Errorf("newMoveDispatch: edge %q references missing target node %q (no geometry loaded for it)", label, ep.Target)
		}
		seg := edgeSegment(srcG, dstG)
		sx, sy, sz := seg.Start.X, seg.Start.Y, seg.Start.Z
		ex, ey, ez := seg.End.X, seg.End.Y, seg.End.Z
		md.gs.edgeSeeds = append(md.gs.edgeSeeds, EdgeGeomSeed{
			Label: label, SrcNode: ep.Source, DstNode: ep.Target,
			SX: sx, SY: sy, SZ: sz, EX: ex, EY: ey, EZ: ez,
		})
	}
	for id, g := range geoms {
		ng := newNodeGeometry(id, g, tr, clk)
		// resolveDest resolves the ONE dedicated directed channel FROM this node
		// (selfID, captured below) TO destID: another node's own neighborIn[selfID]
		// slot, or an incident edge's srcIn/dstIn depending on which endpoint this
		// node is. There is no shared dispatch map to look up — md.mr.nodeGeoms/md.mr.edgeMovers are the
		// read-only directories, safe to read from any goroutine once construction
		// finishes.
		selfID := id
		ng.resolveDest = func(destID string) (chan moveMsg, bool) {
			if em, ok := md.mr.edgeMovers[destID]; ok {
				switch selfID {
				case em.srcID:
					return em.srcIn, true
				case em.dstID:
					return em.dstIn, true
				}
				return nil, false
			}
			if other, ok := md.mr.nodeGeoms[destID]; ok {
				if ch, ok := other.neighborIn[selfID]; ok {
					return ch, true
				}
			}
			return nil, false
		}
		ng.sendMove = md.enqueueFor(ng)
		ng.tap = md.tapToInstall
		ng.centerOf = md.centerOfNode
		ownGeom := ng
		ng.commitLocal = func(_ string, newPos vec3) { md.commitNodeMoveLocal(ownGeom, newPos) }
		md.mr.nodeGeoms[id] = ng
		// Seed the dispatch goroutine's center mirror from the same load-time geom
		// (single-threaded setup, before md.Start — no driving goroutine is running yet)
		// so the first framing read has every center before any push arrives (mirrors
		// the partnerCenters seed below).
		md.mr.centerMirror[id] = nodeWorldPos(g)
	}
	// A pair that points at each other is a LOAD-TIME fact of the edge set, resolved here
	// (single-threaded setup, before any mover goroutine exists) and copied into each
	// node's own field. Without it a node cannot know whether its target also aims back —
	// its own outTargets say nothing about the other node's — and the two chains would draw
	// along the identical centre line (parallelChainOffset, port_geometry.go).
	hasEdge := make(map[string]bool, len(edgeEndpoints))
	for _, ep := range edgeEndpoints {
		hasEdge[ep.Source+"\x00"+ep.Target] = true
	}
	for _, ep := range edgeEndpoints {
		if !hasEdge[ep.Target+"\x00"+ep.Source] {
			continue
		}
		if nm, ok := md.mr.nodeGeoms[ep.Source]; ok {
			if nm.mutualTargets == nil {
				nm.mutualTargets = map[string]bool{}
			}
			nm.mutualTargets[ep.Target] = true
		}
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
		if srcNM, ok := md.mr.nodeGeoms[ep.Source]; ok {
			if dstNM, ok := md.mr.nodeGeoms[ep.Target]; ok {
				if _, exists := dstNM.neighborIn[ep.Source]; !exists {
					dstNM.neighborIn[ep.Source] = make(chan moveMsg, moverInboxDepth)
				}
				if _, exists := srcNM.neighborIn[ep.Target]; !exists {
					srcNM.neighborIn[ep.Target] = make(chan moveMsg, moverInboxDepth)
				}
			}
		}
	}
	// Seed every nodeMover's own partnerCenters map: quantized_move.go's neighbor-move
	// math (neighborSetCReposition et al.) reads a direct neighbor's CURRENT world center
	// off THIS node's OWN partnerCenters map (owned, written only by this node's own
	// goroutine), kept current thereafter by each neighbor's own moveMsgKindNeighborCenter
	// push (applyCenter) — one hop, no cascade.
	for _, nm := range md.mr.nodeGeoms {
		// Seed partnerCenters at construction (single-threaded setup, before md.Start —
		// no mover goroutine is running yet, so reading a neighbor's geom directly here
		// is safe) with the SAME value the old snap seed used (newNodeMover seeds snap
		// from nodeWorldPos(geom)), so the first emit reproduces today's center exactly.
		// A node's neighbor set is nm.neighborIn's key set (populated above from
		// edgeEndpoints — one dedicated channel per adjacent node, both directions).
		for neighborID := range nm.neighborIn {
			if other, ok := md.mr.nodeGeoms[neighborID]; ok {
				nm.partnerCenters[neighborID] = nodeWorldPos(other.geom)
			}
		}
	}
	// Give every nodeMover the ids of its OWN incident edges, so a lock-driven move can
	// notify its edges via sendMove (resolveDest's per-pair channel lookup) — no cached
	// channel slice.
	for id, nm := range md.mr.nodeGeoms {
		for edgeID, em := range md.mr.edgeMovers {
			if em.srcID == id || em.dstID == id {
				nm.edgeIDs = append(nm.edgeIDs, edgeID)
			}
		}
	}
	// Row-identity tables: built ONCE here, from nodeSeeds/edgeSeeds (each node seed already
	// carries its own absolute Row = id-1) — see buildRowTables' doc comment for why this is
	// a load-time constant, not a discovery log.
	md.rt.buildRowTables(md.gs.nodeSeeds, md.gs.edgeSeeds, rowCount)
	return md, nil
}

// Bind wires the per-edge source Outs and dest wires into each edgeMover. Thin delegator
// to md.mr (mover_registry.go).
func (md *MoveDispatch) Bind(outSink map[string]*wire.Out, slotReg SlotRegistry) {
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
func (md *MoveDispatch) EdgeOut(edgeID string) *wire.Out {
	return md.mr.edgeOutFor(edgeID)
}

// centerOfNode returns the current world center for a node id. Thin delegator to md.mr
// (mover_registry.go).
func (md *MoveDispatch) centerOfNode(id string) (vec3, bool) {
	return md.mr.centerOfNode(id)
}

// sendMove routes one moveMsg to a node's dedicated external-entry channel (extIn).
// Thin delegator to md.mr (mover_registry.go); md.ctx is threaded through (not part of
// moverRegistry). This is the bare external-entry path (RootMove, gesture.go) with no
// owning mover goroutine, so it never fires a tap — see nodeMover.tap's doc comment.
func (md *MoveDispatch) sendMove(id string, msg moveMsg) {
	md.mr.sendMove(md.ctx, id, msg)
}

// sendTiltEdit routes one panel-driven tilt-angle click to node id's OWN dedicated
// tiltEditIns channel and returns true, or returns false when id has no such channel (a
// kind that never called BuildArgs.TiltEditIn — every kind except Node1/Node2 today),
// telling the caller (applyUpdateTiltVector) to fall back to the old mover-owned path
// instead. Same blocking-with-ctx-cancel-escape shape as sendMove/mr.sendMove, for the
// same reason: this is a bare external-entry send with no owning goroutine to thread a
// ctx from.
func (md *MoveDispatch) sendTiltEdit(id string, msg TiltEditMsg) bool {
	ch, ok := md.tiltEditIns[id]
	if !ok {
		return false
	}
	if md.ctx == nil {
		ch <- msg
		return true
	}
	select {
	case ch <- msg:
	case <-md.ctx.Done():
	}
	return true
}

// enqueueFor returns nm's own non-blocking send function. Thin delegator to md.mr
// (mover_registry.go).
func (md *MoveDispatch) enqueueFor(nm *nodeGeometry) func(id string, msg moveMsg) {
	return md.mr.enqueueFor(nm)
}

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
	if nm, ok := md.mr.nodeGeoms[nodeID]; ok {
		return nm.geom.Kind
	}
	return ""
}

// Quantized scene-polar move math (quantized_move.go): thin delegators to md.lq so
// their existing in-package call sites (tests, node_move.go, gesture.go) are unchanged.
func (md *MoveDispatch) heldCenters() map[string]vec3 { return md.lq.heldCenters(md) }
func (md *MoveDispatch) commitNodeMoveLocal(nm *nodeGeometry, newPos vec3) {
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
