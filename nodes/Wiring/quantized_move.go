// quantized_move.go — the quantized scene-polar move math, owned by layoutQuantizer (god-
// object decomposition): pure move, no logic changes. Held-state snapshots
// (heldCenters/heldEdges), the broadcast to edges/partners on a re-propagated center, the
// owner-goroutine commit (commitNodeMoveLocal), RootMove (the decentralized drag entry),
// and reachRFromPolar.
//
// There is no node-node stored coordinate here (MODEL.md "the polar model" — the deleted
// wire.LocalPolar and its requantize/neighborSetC propagation): a node has ONE polar
// coordinate, about the scene centre only.
//
// RootMove's invariant is load-bearing and MUST stay prominent after this move: it runs
// ONCE PER POINTER-MOVE EVENT, not once per drag (memory/project_rootmove_is_per_pointer_move.md).
//
// Every method below takes md *MoveDispatch explicitly for everything that is NOT part of
// layoutQuantizer's own field (quantizedLayout) — mr/ui/tr/persist/
// centerOfNode/NodeRowFor/sendMove are owned elsewhere. MoveDispatch's
// public RootMove, and its several package-private methods of the same names as below
// (heldCenters, heldEdges, broadcastToEdgesAndPartners, commitNodeMoveLocal), stay thin
// delegators in node_move.go so their existing in-package call sites (tests, node_move.go,
// gesture.go) are unchanged.

package Wiring

import (
	"fmt"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"strings"

	T "github.com/dtauraso/wirefold/Trace"
)

// layoutQuantizer owns the quantized scene-polar move math. See MoveDispatch.lq's doc
// comment (node_move.go) for what it owns and why.
type layoutQuantizer struct {
	// quantizedLayout gates the quantized absolute-scene-polar snap — every node is a
	// root, measured/derived about the scene center only.
	quantizedLayout bool
}

// heldCenters returns a fresh snapshot of every node's current world center, read from
// the dispatch goroutine's own centerMirror (centerOfNode), which each nodeMover keeps
// current by message (drainCenterMirror) — safe to call from the stdin/gesture dispatch
// goroutine, which is every live caller. There is no separate accumulated positions map
// to drain.
func (lq *layoutQuantizer) heldCenters(md *MoveDispatch) map[string]vec3 {
	out := make(map[string]vec3, len(md.mr.nodeGeoms))
	for id := range md.mr.nodeGeoms {
		if c, ok := md.centerOfNode(id); ok {
			out[id] = c
		}
	}
	return out
}

func (lq *layoutQuantizer) heldEdges(md *MoveDispatch) []sphereEdge {
	edges := make([]sphereEdge, 0, len(md.mr.edgeMovers))
	for _, em := range md.mr.edgeMovers {
		edges = append(edges, sphereEdge{Source: em.srcID, Target: em.dstID})
	}
	return edges
}

// broadcastToEdgesAndPartners messages every incident edge's mover (batched per-edge Centers) and
// every aimed-port partner (pure re-emit), for the given already-applied set of moved node
// centers. It never writes the moved node's OWN snap — that responsibility belongs to
// whichever caller applied the moved node's own center via applyCenter directly (every
// live caller is owner-goroutine; the old central "self-send into own inbox" path,
// fanCenters, was removed — it deadlocked/staled when its only caller turned out to run
// on the moved node's own goroutine too. See commitNodeMoveLocal for the applyCenter +
// broadcastToEdgesAndPartners pattern).
func (lq *layoutQuantizer) broadcastToEdgesAndPartners(md *MoveDispatch, newCenters map[string]vec3, enqueue func(id string, msg moveMsg)) {
	// Per-edge: send ONE batched message carrying every moved endpoint of that edge,
	// so an edge whose both endpoints moved this frame recomputes/emits exactly once.
	// enqueue (the sending node's own retry queue — nm.sendMove) appends the
	// message to nm.pending and attempts an immediate non-blocking send on the
	// destination's own directed channel (extIn or the sender's slot in the
	// destination's neighborIn map), retrying next cycle if that channel isn't ready
	// to receive; it never blocks the calling handler goroutine, so this call — made
	// from inside handle via commitLocal — never blocks. Dispatch-existence (does id
	// resolve to a live mover) is checked at send time inside that retry path, matching
	// enqueue's other call sites (m.sendMove), which already tap/enqueue unconditionally
	// regardless of whether id resolves.
	for edgeID, em := range md.mr.edgeMovers {
		eps := map[string]vec3{}
		if c, ok := newCenters[em.srcID]; ok {
			eps[em.srcID] = c
		}
		if c, ok := newCenters[em.dstID]; ok {
			eps[em.dstID] = c
		}
		if len(eps) == 0 {
			continue
		}
		enqueue(edgeID, moveMsg{Kind: moveMsgKindCenters, Centers: eps})
	}

	// Partner re-emit: find every partner node — the OTHER end of any edge incident to a
	// moved node — and ask it to re-emit its OWN geometry with its OWN (unchanged)
	// center. Node geometry no longer depends on a connected partner's position at all
	// (a port carries no geometry, docs/channels-not-ports.md — this used to be how an
	// AIMED port picked up its moved partner's fresh center; that aiming is gone), but
	// the re-emit stays: it is what keeps a downstream watcher's view of this partner
	// current on the SAME cadence a moved node's own re-emit fires, without adding a
	// second, separately-timed signal.
	// partners maps partnerID → the ONE moved node (kept for clarity/observability
	// parity with the prior shape; movedID itself is otherwise unused now that the
	// re-emit carries no cache payload).
	partners := map[string]string{}
	for _, em := range md.mr.edgeMovers {
		if _, moved := newCenters[em.srcID]; moved {
			if _, alsoMoved := newCenters[em.dstID]; !alsoMoved {
				partners[em.dstID] = em.srcID
			}
		}
		if _, moved := newCenters[em.dstID]; moved {
			if _, alsoMoved := newCenters[em.srcID]; !alsoMoved {
				partners[em.srcID] = em.dstID
			}
		}
	}
	for partnerID, movedID := range partners {
		if _, ok := md.mr.nodeGeoms[partnerID]; !ok {
			continue
		}
		// Center is deliberately nil (see the doc comment above): this is a PURE
		// re-emit, not a position write for partnerID itself — a non-nil Center here
		// would be read by nodeMover.handle as "this is YOUR OWN new center" and
		// wrongly move partnerID. nodeMover.handle's nil-Center branch re-emits from
		// the mover's own live geom, so it can never race or clobber a pending position
		// write. Per-target FIFO order (each sender's own retry queue
		// drains in append order onto that target's one directed channel) preserves
		// ordering now that delivery goes through the sender's own
		// nm.pending/flushPending instead of a shared outbox.
		enqueue(partnerID, moveMsg{Kind: moveMsgKindCenter, NodeID: partnerID, Center: nil,
			SenderID: movedID})
	}
}

// touchingBead is the per-neighbour geometry commitNodeMoveLocal hands to beadCrudDecide:
// the touching bead's own SOURCE point, its own CURRENT centre, and the chain's own AXIS
// (AimDir, the live unit direction from nm toward the neighbour) — see dragTouchingBeads.
// AimDir is what the node's post-CRUD position is computed along (beadCrudImpliedCentre);
// it is never the raw drag direction.
type touchingBead struct {
	NeighborID string
	Source     vec3
	Centre     vec3
	AimDir     vec3
}

// dragTouchingBeads reads, for EVERY direct neighbour of the dragged node nm, the ONE bead
// on that edge that directly touches nm — from nm's own live state only (its own kind,
// prevPos, and partnerCenters, the same live-copy map chainBeads/requantizeLocalPolars
// already read). A neighbour with no live centre yet (never pushed an applyCenter)
// contributes no touching bead, same convention chainBeads uses for a target with no live
// partner center.
//
// The touching bead's own centre sits at the SAME fixed offset from nm regardless of which
// node owns the chain — tangency to nm's own torus falls out of the placement formula
// (chain_beads.go), not from where nm's centre happens to be:
//
//	beadCentre = prevPos + aimDir*(selfTorusR + wire.BeadTorusOuterR)
//
// where aimDir is the live unit direction from nm toward the neighbour. What differs by
// ownership is the touching bead's SOURCE point (PLAN.md: "the previous bead's centre
// along its chain, or the chain origin on the neighbour's torus surface when it is the
// only bead") — NEVER the bead's own centre, which would be wrong by one bead:
//
//   - nm is the edge's SOURCE (nm.outTargets contains the neighbour): the touching bead is
//     always chain index 0 — it has no predecessor regardless of the current count, so its
//     source is the chain's own origin, nm's own torus surface point:
//     beadSource = prevPos + aimDir*selfTorusR
//   - nm is the edge's TARGET (an incoming edge): the touching bead is the chain's LAST
//     bead, owned and counted by the neighbour (edgeStepCount, same formula chain_beads.go
//     uses, mirrored here on the live distance). With more than one bead, its predecessor
//     is the bead one step back toward the neighbour:
//     beadSource = beadCentre + aimDir*wire.BeadStepR
//     With exactly one bead, there is no predecessor bead — the chain's own origin is the
//     NEIGHBOUR's torus surface (nm.neighborKinds gives the neighbour's kind, derived from
//     domain adjacency at load — see build.go — so every direct neighbour has an entry):
//     beadSource = neighborCenter - aimDir*nodeTorusOuterR(neighborKind)
func dragTouchingBeads(md *MoveDispatch, nm *nodeGeometry, prevPos vec3) []touchingBead {
	nodeID := nm.id
	selfTorusR := nodeTorusOuterR(nm.selfKind)
	out := make([]touchingBead, 0, len(nm.edgeIDs))
	for _, edgeID := range nm.edgeIDs {
		em, ok := md.mr.edgeMovers[edgeID]
		if !ok {
			continue
		}
		neighborID := em.srcID
		isSource := em.srcID == nodeID
		if isSource {
			neighborID = em.dstID
		}
		neighborCenter, ok := nm.partnerCenters[neighborID]
		if !ok {
			continue
		}
		dist, aimDir, ok := edgeCenterDistAndDir(prevPos, neighborCenter)
		if !ok {
			continue
		}
		beadCentre := prevPos.Add(aimDir.Scale(selfTorusR + wire.BeadTorusOuterR))

		// A touching bead's SOURCE POINT is the previous bead's centre along its chain, or
		// the chain origin on the NEIGHBOUR's torus surface when it is the only bead
		// (MODEL.md's bead-CRUD section) — the SAME rule whichever end of the edge this
		// node happens to own. Which endpoint stores the edge changes nothing about where
		// the beads sit.
		//
		// This used to special-case `isSource` to `prevPos + aimDir*selfTorusR` — a point
		// on THIS node's own torus surface, which is neither of the two the model allows.
		// It broke the rule twice over:
		//   - |third| at rest became selfTorusR (~5 bead lengths, ~44.8 world units) rather
		//     than one bead length, so |third| never fell below the one-bead threshold and
		//     REMOVE could never fire at all; and
		//   - beadVec came out as +aimDir*BeadTorusOuterR, pointing TOWARD the neighbour
		//     instead of away from it, inverting the angle gate — dragging AWAY (which
		//     should open a gap and add a bead) scored > 90 degrees and was blocked, while
		//     dragging toward admitted an add whose implied centre sat ~31 world units
		//     away. Hence: drag far in most directions and nothing happens; drag the other
		//     way and the node jumps.
		neighborKind := nm.neighborKinds[neighborID]
		count := edgeStepCount(dist, neighborKind, nm.selfKind)
		var beadSource vec3
		if count >= 2 {
			beadSource = beadCentre.Add(aimDir.Scale(wire.BeadStepR))
		} else {
			beadSource = neighborCenter.Sub(aimDir.Scale(nodeTorusOuterR(neighborKind)))
		}
		out = append(out, touchingBead{NeighborID: neighborID, Source: beadSource, Centre: beadCentre, AimDir: aimDir})
	}
	return out
}

// commitNodeMoveLocal is the OWNER-GOROUTINE single-node commit path
// (generalized to every node): used when the commit
// originates on nodeID's OWN mover goroutine (its own inbox handler for a
// moveMsgKindDrag). It applies nodeID's OWN new center SYNCHRONOUSLY via
// applyCenter — safe and correct here because applyCenter's doc contract is "called
// only from this nodeMover's own inbox-drain goroutine", which this is. Also fans
// centers to incident edges/partners, persists the per-node quantized-offset
// (nodeMover.quantOffset — never a shared map, so no other mover goroutine's commit
// can race this write even for a different node id), and requantizes nodeID's
// local-polar cascade-links against its (unmoved) neighbors.
func (lq *layoutQuantizer) commitNodeMoveLocal(md *MoveDispatch, nm *nodeGeometry, newPos vec3) {
	nodeID := nm.id
	edges := lq.heldEdges(md)
	// reach[nodeID] only ever needs nodeID's own fresh polar plus its DIRECT
	// neighbors' polar (reachRFromPolar only accumulates reach for an edge's
	// SOURCE, from that edge's Target) — each direct neighbor's last-pushed
	// CARTESIAN center is read from THIS node's OWN partnerCenters map (nm.
	// partnerCenters, kept current by every neighbor's applyCenter push — see its
	// doc comment), resolved via nm.edgeIDs (this node's own incident edges, fixed
	// at construction; every edgeIDs neighbor is by construction a key of
	// nm.neighborIn, the same set partnerCenters is seeded/kept from). scene polar
	// is a pure re-derive off the fixed, write-once md.ui.sceneSphere.Center (never
	// mutated after load), so this stays race-free with no cross-goroutine read at
	// all now (this runs on nm's own goroutine, reading nm's own map).
	polars := map[string]polar{}
	for _, edgeID := range nm.edgeIDs {
		em, ok := md.mr.edgeMovers[edgeID]
		if !ok {
			continue
		}
		neighborID := em.srcID
		if neighborID == nodeID {
			neighborID = em.dstID
		}
		if c, ok := nm.partnerCenters[neighborID]; ok {
			polars[neighborID] = cart2polar(c.Sub(md.ui.sceneSphere.Center))
		}
	}
	// Single cart2polar boundary conversion for this drag target — newPos is mouse-
	// derived cartesian (gesture.go ray/plane unproject); everything downstream
	// (reach, measureScalar, the persist schedule) reuses this one polar value rather
	// than re-deriving it from newPos.
	nodePolar := cart2polar(newPos.Sub(md.ui.sceneSphere.Center))

	// committedPos/committedPolar are what gets DRAWN (applyCenter), FANNED
	// (broadcastToEdgesAndPartners), PERSISTED (persistQuantOffset), and re-quantized
	// against by every neighbor (requantizeLocalPolars) for this commit — ONE position,
	// not the raw drag target for some of those and a quantized point for others
	// (docs/which-lattice-a-node-lives-on.md "Why the drag makes it worst": that split is
	// exactly what made the node glide continuously while its own chain beads jumped one
	// bead distance at a time). Under the quantized scene lattice (lq.quantizedLayout),
	// moving the node is now CRUD on the edge beads that touch it (PLAN.md, bead_crud.go)
	// instead of solving a joint lattice-intersection: EVERY touching bead
	// (dragTouchingBeads) judges the SAME raw mouse target independently
	// (resolveBeadCrudMove/beadCrudDecide) — no solver, no enumeration across neighbours,
	// no selection of one edge over another, and no summing of per-edge results into a
	// displacement. The node's new centre comes from the BEAD OPERATION
	// (beadCrudImpliedCentre), along that edge's own chain axis — NEVER from the raw drag
	// target, which supplies nodeDestination for the third-vector test and the angle gate
	// only (PLAN.md "the node moves the bead's distance ... NOT the drag destination
	// point"). If every touching bead's verdict is "none" (or the node has no touching
	// beads at all, a free node with no incident edges), the raw drag target is used
	// directly for a free node, matching the old solver's N==0 branch; with incident edges
	// and every verdict "none", the node holds prevPos. off/committedPolar below are still
	// measured back OFF committedPos purely as the position.json self-describing CACHE
	// (quant_offset_persist.go's doc comment: "the quantized scalar triple... rides along
	// as a self-describing cache of the drag-time snap cells, NOT the position source")
	// — nothing downstream reconstructs committedPos from off. If quantizedLayout is off,
	// keep the historic behavior: committedPos stays the raw, continuous target, and no
	// offset is measured.
	committedPos := newPos
	committedPolar := nodePolar
	var off quantizedOffset
	if lq.quantizedLayout {
		prevPos := nodeWorldPos(nm.geom)
		beads := dragTouchingBeads(md, nm, prevPos)
		if len(beads) == 0 {
			committedPos = newPos
		} else {
			committedPos, _ = resolveBeadCrudMove(beads, prevPos, newPos, wire.BeadStepR)
		}
		committedPolar = cart2polar(committedPos.Sub(md.ui.sceneSphere.Center))

		// DIAGNOSTIC ONLY (task/log-node2-bead-crud): one breadcrumb per pointer-move
		// commit — node 2 (neighbours 1, 4, 5) can barely be dragged; long drags produce
		// no movement, and beads that should be ADDED to push it the right way are not
		// being added. This packs the whole event PLUS every touching bead's own CRUD
		// arithmetic (why it returned none/add/remove) so the actual numbers, not a
		// theory, explain it. Gated on nm.tr != nil exactly like the neighbor-center-recv/
		// neighbor-setc-recv breadcrumb sites above (node_mover.go) — cheap no-op with no
		// stream wired (headless tests, bare movers).
		if nm.tr != nil {
			dragVector := newPos.Sub(prevPos)
			parts := make([]string, 0, len(beads))
			for _, b := range beads {
				diag := beadCrudDiagnose(b.NeighborID, b.Source, b.Centre, b.AimDir, prevPos, newPos, dragVector, wire.BeadStepR)
				verdictStr := "none"
				switch diag.Verdict {
				case beadCrudAdd:
					verdictStr = "add"
				case beadCrudRemove:
					verdictStr = "remove"
				}
				impliedStr := "none"
				if diag.ImpliedOK {
					impliedStr = fmt.Sprintf("(%.4f,%.4f,%.4f)", diag.Implied.X, diag.Implied.Y, diag.Implied.Z)
				}
				parts = append(parts, fmt.Sprintf(
					"[nbr=%s third=%.4f beadLen=%.4f verdict=%s cosA=%.4f gateBlocked=%v srcDist=%.4f implied=%s]",
					diag.NeighborID, diag.ThirdLen, diag.BeadLen, verdictStr, diag.CosAngle, diag.GateBlocked, diag.SourceDist, impliedStr))
			}
			dragLen := dragVector.Length()
			committedDelta := committedPos.Sub(prevPos).Length()
			value := fmt.Sprintf(
				"node=%s prevPos=(%.4f,%.4f,%.4f) dest=(%.4f,%.4f,%.4f) dragLen=%.4f committed=(%.4f,%.4f,%.4f) committedDelta=%.4f moved=%v beads=%s",
				nodeID, prevPos.X, prevPos.Y, prevPos.Z, newPos.X, newPos.Y, newPos.Z, dragLen,
				committedPos.X, committedPos.Y, committedPos.Z, committedDelta, committedDelta > 1e-9,
				strings.Join(parts, " "))
			nm.tr.Breadcrumb("bead-crud", nodeID, "", value)
			nm.writeStreamFrame([]wire.RowEvent{{
				Kind: T.KindBreadcrumb, Label: T.BreadcrumbBeadCrud, Debug: 1,
				NodeRow: nm.nodeRow, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
				Text: value,
			}})
		}
	}

	polars[nodeID] = committedPolar
	reach := reachRFromPolar(polars, edges)

	nm.applyCenter(committedPos, reach[nodeID])
	lq.broadcastToEdgesAndPartners(md, map[string]vec3{nodeID: committedPos}, nm.sendMove)

	// PERSIST ON EVERY DRAG, both modes. This used to sit inside `if lq.quantizedLayout`,
	// which silently stopped saving the moment a scene chose the continuous drag: the node
	// moved, drew, fanned to its neighbours and looked entirely correct, and the position
	// was gone on the next load. The two modes differ in WHERE the node lands, never in
	// whether that landing is written down.
	//
	// off is the quantized scalar triple, measured here in both modes because position.json
	// carries it as a self-describing CACHE of the drag-time snap cells, not as the position
	// source (quant_offset_persist.go) — the source is committedPolar either way. Measuring
	// it under the continuous drag keeps that cache describing the position actually stored
	// rather than the last one a quantized drag happened to leave behind.
	off = measureScalar(committedPolar, nm.quantOffset)
	nm.quantOffset = off
	nm.persistQuantOffset(off, committedPolar)
}

// RootMove handles a node-drag under the flat absolute scene-polar layout: every node
// is positioned independently about the scene sphere center — there is no reference/
// parent concept, so dragging moves ONLY the dragged node (no cascade). The dragged
// node's COMMITTED world position (commitNodeMoveLocal) is the drag target SNAPPED to
// the scene lattice, moving exactly one bead distance per commit, the same distance its
// own chain beads move.
//
// RootMove is the decentralized drag entry, widened to EVERY node (the generalization
// that came with the quantizedOffsets data-race fix): dragging any node does not commit
// on the stdin reader's own goroutine — it routes a single moveMsgKindDrag to the
// dragged node's OWN inbox and returns. The dragged node's own moveMsgKindDrag handler
// (nodeMover.handle) does the rest, entirely on its own goroutine: commit its own new
// position (commitLocal — fan + persist, no cross-goroutine self-send). There is no
// equal-radii solve, no rule-node cascade, no gate-anchor broadcast, and no per-neighbour
// re-quantize message any more (MODEL.md "the polar model": a node has no stored
// coordinate for a neighbour); a drag never touches any node's position but its own.
// Returns false for an unknown node.
//
// NOTE: RootMove runs ONCE PER POINTER-MOVE EVENT, not once per drag (two bugs — commits
// 338f05da, 154a05bd — came from assuming otherwise; see
// memory/project_rootmove_is_per_pointer_move.md). The drag-log reset is NOT emitted
// here for that reason: the reset belongs at the real drag-start edge (the
// pending→dragging transition in gesture.go), not on every move tick RootMove sees.
func (lq *layoutQuantizer) RootMove(md *MoveDispatch, nodeID string, target vec3) bool {
	if _, ok := md.mr.nodeGeoms[nodeID]; !ok {
		return false
	}
	// Route the drag itself to the dragged node's OWN inbox instead of committing on
	// the stdin reader's goroutine — every node's moveMsgKindDrag handler commits
	// (synchronous local apply, reported over reportCh) on its own goroutine. No
	// central commit call here.
	md.sendMove(nodeID, moveMsg{Kind: moveMsgKindDrag, NodeID: nodeID, Target: target})
	return true
}

// reachRFromPolar computes each node's sphere REACH radius (max distance from a node to any
// node it outputs to) under the given polar positions and edge set. Distance is the spherical
// law-of-cosines distance between the two polar positions (polarDist) — no cartesian, no vector
// subtraction. Called by loader.go buildFromSpec and by RootMove so the fanned "center" message
// carries the new reach radius and the ring stays sized during a drag.
func reachRFromPolar(polars map[string]polar, edges []sphereEdge) map[string]float64 {
	reachR := map[string]float64{}
	for _, e := range edges {
		sp, okS := polars[e.Source]
		tp, okT := polars[e.Target]
		if !okS || !okT {
			continue
		}
		if d := polarDist(sp, tp); d > reachR[e.Source] {
			reachR[e.Source] = d
		}
	}
	return reachR
}
