// quantized_move.go — the quantized cascade-link local-polar move math, owned by
// layoutQuantizer (god-object decomposition): pure move, no logic changes. Held-state
// snapshots (heldCenters/heldEdges), the broadcast to edges/partners on a re-propagated
// center, pole requantizing, the one-hop neighborSetC propagation, the owner-goroutine
// commit (commitNodeMoveLocal), RootMove (the decentralized drag entry),
// requantizeLocalPolars, and reachRFromPolar.
//
// RootMove's invariant is load-bearing and MUST stay prominent after this move: it runs
// ONCE PER POINTER-MOVE EVENT, not once per drag (memory/project_rootmove_is_per_pointer_move.md).
//
// Every method below takes md *MoveDispatch explicitly for everything that is NOT part of
// layoutQuantizer's own two fields (quantizedLayout, layoutHolders) — mr/ui/tr/persist/
// centerOfNode/NodeRowFor/sendMove are owned elsewhere. MoveDispatch's
// public RootMove, and its several package-private methods of the same names as below
// (heldCenters, heldEdges, broadcastToEdgesAndPartners, requantizePoleTraced,
// neighborSetCRequantize, commitNodeMoveLocal, requantizeLocalPolars), stay thin
// delegators in node_move.go so their existing in-package call sites (tests, node_move.go,
// gesture.go) are unchanged.

package Wiring

import (
	"fmt"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"math"

	T "github.com/dtauraso/wirefold/Trace"
)

// layoutQuantizer owns the quantized cascade-link local-polar move math. See
// MoveDispatch.lq's doc comment (node_move.go) for what it owns and why.
type layoutQuantizer struct {
	// quantizedLayout gates the quantized absolute-scene-polar snap — every node is a
	// root, measured/derived about the scene center only.
	quantizedLayout bool
	// layoutHolders resolves a node id to the *LayoutHolder embedded in that node's
	// built struct (reflection-attached by buildNodes the same way LocalPolars itself
	// is attached — see loader.go). This is the ONLY route from the drag path
	// (RootMove) to each node's own LayoutHolder; MoveDispatch does not own or copy
	// LocalPolars itself, it just routes the update to the owning node.
	layoutHolders map[string]*wire.LayoutHolder
}

// heldCenters returns a fresh snapshot of every node's current world center, read from
// the dispatch goroutine's own centerMirror (centerOfNode), which each nodeMover keeps
// current by message (drainCenterMirror) — safe to call from the stdin/gesture dispatch
// goroutine, which is every live caller. There is no separate accumulated positions map
// to drain.
func (lq *layoutQuantizer) heldCenters(md *MoveDispatch) map[string]vec3 {
	out := make(map[string]vec3, len(md.mr.nodeMovers))
	for id := range md.mr.nodeMovers {
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
		if _, ok := md.mr.nodeMovers[partnerID]; !ok {
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

// requantizePoleTraced is the SINGLE site every LOCAL-polar write routes through once a
// node's LayoutHolder exists (this file's several call sites). `updates` carries the FRESH
// offset (vec3, THIS node — nodeID — as origin) for each neighbor whose distance/direction
// just changed — the legitimate cart↔polar boundary entry (dirFromOffset + azimuthFrom).
//
// Every OTHER neighbor already on lh (unchanged this call) is NEVER re-measured against a
// live cartesian center: its direction is RECONSTRUCTED from its own stored indices about
// the OLD pole (lh.Pole(), persisted from the last call) via fromAxisFrame — arithmetic on
// stored ints × step constants, then one boundary trig call to turn that direction back into
// a vector. This is the fixed-increment/stored-index model
// (memory/feedback_abc_times_constant_not_rederive.md,
// docs/demos/polar-drag-3d.html's autoPole/ΔR⁻¹·q block): an unchanged neighbor's world
// position hasn't moved, so its stored indices ARE ground truth and are carried forward,
// adjusted only by the fixed pole increment (rotating_pole.go) when the measurement pole
// tilts. `pole = localPole(dirs)` is recomputed from the WHOLE neighbor set's directions
// (fresh from cartesian, unchanged from stored-index reconstruction) and then PERSISTED on
// lh (SetPole) so the next call's unchanged neighbors reconstruct against the pole THIS
// call actually quantized against.
//
// When the pole doesn't move (the common case — home stays home), an unchanged neighbor's
// re-expressed indices are byte-identical to what's already stored (fromAxisFrame then
// azimuthFrom about the SAME pole is an exact round-trip): the write is skipped, a true
// no-op, not a reproject that happens to land on the same numbers.
func (lq *layoutQuantizer) requantizePoleTraced(lh *wire.LayoutHolder, updates map[string]vec3) dir {
	existing := lh.LocalPolarsSnapshot()
	oldPole := lh.Pole()

	existingByID := make(map[string]wire.LocalPolar, len(existing))
	for _, lp := range existing {
		existingByID[lp.To] = lp
	}

	// Each neighbor's DIRECTION: fresh neighbors from their live cartesian offset (the
	// legitimate boundary entry); unchanged neighbors reconstructed from stored indices
	// about the OLD pole — no md.centerOfNode call for an unchanged neighbor.
	dirs := make(map[string]dir, len(existing)+len(updates))
	freshRadius := make(map[string]float64, len(updates))
	for id, o := range updates {
		d, r := dirFromOffset(o)
		dirs[id] = d
		freshRadius[id] = r
	}
	for _, lp := range existing {
		if _, fresh := updates[lp.To]; fresh {
			continue
		}
		t, p, _ := lp.EffectiveSteps()
		dirs[lp.To] = fromAxisFrame(oldPole, float64(lp.QuantITheta)*t, float64(lp.QuantIPhi)*p)
	}

	dirVecs := make([]vec3, 0, len(dirs))
	for _, d := range dirs {
		dirVecs = append(dirVecs, dirToVec3(d))
	}
	newPole := localPole(dirVecs)

	for id, d := range dirs {
		t, p, rStep := lh.LocalPolarSteps(id)
		c, psi := azimuthFrom(newPole, d)
		iTheta := int(math.Round(c / t))
		iPhi := int(math.Round(psi / p))

		old, hadEntry := existingByID[id]
		_, fresh := updates[id]

		iR := old.QuantIR
		if fresh || !hadEntry {
			iR = int(math.Round(freshRadius[id] / rStep))
		}

		if !fresh && hadEntry &&
			old.QuantITheta == iTheta && old.QuantIPhi == iPhi && old.QuantIR == iR &&
			old.StepTheta == t && old.StepPhi == p && old.StepR == rStep {
			continue // true no-op: pole/indices unchanged, skip the write
		}
		lh.SetLocalPolar(id, iTheta, iPhi, iR, t, p, rStep)
	}
	lh.SetPole(newPole)
	return newPole
}

// neighborSetCRequantize is the OWNER-GOROUTINE half of a neighbor's edge re-quantize
// (moveMsgKindNeighborSetC): the dragged node fromID moved, so selfID's stored local
// polar to fromID no longer matches the live geometry. selfID STAYS PUT — dragging fromID
// moves only fromID — and re-quantizes its OWN edge to fromID from the live offset, with
// theta, phi AND r all fresh (about selfID's rotating pole, via requantizePoleTraced with
// fromID as the single fresh update); selfID's OTHER neighbors are carried forward as
// index x step, not re-derived. There is NO reposition: only fromID moved, so the incident
// fromID-selfID edge redraws from fromID's own commit (broadcastToEdgesAndPartners on fromID's
// side). Single hop — no forward past selfID, no cascade. No-op for an unknown selfID.
// deltaA/deltaB/deltaC are the DRAGGED node fromID's OWN quantized-triple change for its
// edge to selfID (computed once on fromID's goroutine, see requantizeLocalPolars) —
// pure observability payload carried through to the AbcDrag trace event, never applied
// to selfID's own position/quantize math. selfCenter is selfID's OWN current center —
// read by the caller (nodeMover.handle, on selfID's own goroutine, nodeWorldPos(m.geom)).
func (lq *layoutQuantizer) neighborSetCRequantize(md *MoveDispatch, selfID, fromID string, selfCenter, fromCenter vec3, deltaA, deltaB, deltaC int) {
	lh, ok := lq.layoutHolders[selfID]
	if !ok {
		return
	}
	// Offset convention matches requantizeLocalPolars: neighbor(fromID) center - self
	// center. fromID is the ONLY fresh update, so requantizePoleTraced re-derives selfID's
	// edge to fromID (theta, phi AND r, about selfID's rotating pole) at the cart<->polar
	// boundary, while every OTHER neighbor of selfID is carried forward as index x step.
	lq.requantizePoleTraced(lh, map[string]vec3{fromID: fromCenter.Sub(selfCenter)})

	// EVERY node that receives an abc change from a dragged peer logs its response so the
	// drag propagation is observable (probe-merge.sh --debug -> .probe/go-debug.jsonl) and
	// the in-editor overlay log can list all recipients — NOT gated to time nodes: any node
	// that gets the message (gate, time, pulse, ...) is a recipient and must be mentioned.
	// Behavior is still the plain stay-put re-quantize above; this is the observability
	// step, not a motion/propagation change. The logged abc is selfID's freshly re-quantized
	// edge to the peer.
	if md.tr != nil {
		var it, ip, ir int
		for _, lp := range lh.LocalPolarsSnapshot() {
			if lp.To == fromID {
				it, ip, ir = lp.QuantITheta, lp.QuantIPhi, lp.QuantIR
				break
			}
		}
		md.tr.Breadcrumb("abc-drag", selfID, fromID,
			fmt.Sprintf("peer=%s peerCenter=(%.3f,%.3f,%.3f) abc=(%d,%d,%d) delta=(%d,%d,%d)",
				fromID, fromCenter.X, fromCenter.Y, fromCenter.Z, it, ip, ir, deltaA, deltaB, deltaC))
	}
	// selfID's OWN recipient bit AND cumulative recipient count: this runs on selfID's
	// OWN nodeMover goroutine (neighborSetC dispatch, node_mover.go's
	// moveMsgKindNeighborSetC case), so it is safe to write nm's fields directly — no
	// shared map, no cross-goroutine channel. This is what writeStreamFrame (also this
	// goroutine) reads for its own dedicated stream frame; dragRequantCount is the sole
	// source of the "drag received ×N" count now (summed across node rows on the TS
	// side) — no central accumulator, so nothing can drop a tick.
	if nm, ok := md.mr.nodeMovers[selfID]; ok {
		nm.gotDragMsg = 1
		nm.dragRequantCount++
		nm.dragDeltaA, nm.dragDeltaB, nm.dragDeltaC = int32(deltaA), int32(deltaB), int32(deltaC)
		// selfID's own local-polars.json — this runs on selfID's own nodeMover goroutine
		// (see the comment block above), so nm persists it directly rather than reaching
		// through md.persist (.claude/rules/persistence-ownership.md "The owner writes, and owns the path").
		nm.persistLocalPolars(lh.LocalPolarsSnapshot(), lh.Pole())
		// Structured buffer counterpart of the "abc-drag" breadcrumb above, riding
		// THIS node's (selfID's) own dedicated stream — this runs on selfID's own
		// nodeMover goroutine, mirroring gotDragMsg/dragDelta* just above. it/ip/ir (selfID's re-quantized abc to
		// fromID) reuse Value/EdgeRow/Slot; deltaA/B/C (fromID's own change) reuse
		// X/Y/Z — none of those columns carry their ordinary meaning on a
		// Breadcrumb-kind row (see bufLayoutEvent's doc comment: columns are REUSED
		// per Kind).
		if md.tr != nil {
			var it, ip, ir int
			for _, lp := range lh.LocalPolarsSnapshot() {
				if lp.To == fromID {
					it, ip, ir = lp.QuantITheta, lp.QuantIPhi, lp.QuantIR
					break
				}
			}
			targetRow := int32(-1)
			if r, ok := md.NodeRowFor(fromID); ok {
				targetRow = r
			}
			nm.writeStreamFrame([]wire.RowEvent{{
				Kind: T.KindBreadcrumb, Label: T.BreadcrumbAbcDrag, Debug: 1,
				NodeRow: nm.nodeRow, PortRow: -1, TargetRow: targetRow, TargetPortRow: -1, EdgeRow: int32(ip), Slot: int32(ir),
				Value: int32(it), X: float64(deltaA), Y: float64(deltaB), Z: float64(deltaC),
			}})
		}
	}

	// Delta-forward (observability only — no re-quantize, no move): selfID is a direct
	// drag-recipient, so it forwards the SAME delta triple it just received from fromID
	// to every OTHER cascade-link neighbor (every cascade-link neighbor except fromID —
	// see nodeMover.forwardDelta's doc comment). Every forward recipient in turn does
	// its OWN hop (handle's moveMsgKindDeltaForward case) — independent, concurrent hops
	// that together spread the triple across whatever the per-kind relay rules make
	// reachable. The stored cascade-link set is NOT loop-free by construction (it now
	// includes both cycle-closing links); termination comes from those rules — see
	// nodeMover.forwardDelta. This still runs on EVERY move (not just the first this
	// drag), keeping the forwarded log in sync with the drag as it continues — there is
	// no once-per-drag guard to gate it.
	if nm, ok := md.mr.nodeMovers[selfID]; ok {
		nm.forwardDelta(md, fromID, int32(deltaA), int32(deltaB), int32(deltaC))
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
//     NEIGHBOUR's torus surface (nm.cascadeKinds gives the neighbour's kind; cascade
//     adjacency is validated equal to domain adjacency, so every direct neighbour has an
//     entry):
//     beadSource = neighborCenter - aimDir*nodeTorusOuterR(neighborKind)
func dragTouchingBeads(md *MoveDispatch, nm *nodeMover, prevPos vec3) []touchingBead {
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
		neighborKind := nm.cascadeKinds[neighborID]
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
func (lq *layoutQuantizer) commitNodeMoveLocal(md *MoveDispatch, nm *nodeMover, newPos vec3) {
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
		off = measureScalar(committedPolar, nm.quantOffset)
	}

	polars[nodeID] = committedPolar
	reach := reachRFromPolar(polars, edges)

	nm.applyCenter(committedPos, reach[nodeID])
	lq.broadcastToEdgesAndPartners(md, map[string]vec3{nodeID: committedPos}, nm.sendMove)

	if lq.quantizedLayout {
		nm.quantOffset = off
		nm.persistQuantOffset(off, committedPolar)
	}

	lq.requantizeLocalPolars(md, nm, committedPos)
}

// RootMove handles a node-drag under the flat absolute scene-polar layout: every node
// is positioned independently about the scene sphere center — there is no reference/
// parent concept, so dragging moves ONLY the dragged node (no cascade). The dragged
// node's COMMITTED world position (commitNodeMoveLocal) is the drag target SNAPPED to
// the scene lattice — no longer continuous: the scene lattice and the local-polar
// lattice beads are laid out from are now the SAME lattice (quantized_layout.go
// stepTheta/stepPhi/stepR == the local-polar defaults, docs/which-lattice-a-node-lives-on.md),
// so a node moves exactly one bead distance per commit, the same distance its own chain
// beads move — each neighbor's DISTANCE to it is quantized separately, each on that
// neighbor's own small grid (see requantizeLocalPolars), independently of this snap.
//
// RootMove is the decentralized drag entry, widened to EVERY node (the generalization
// that came with the quantizedOffsets data-race fix): dragging any node does not commit
// on the stdin reader's own goroutine — it routes a single moveMsgKindDrag to the
// dragged node's OWN inbox and returns. The dragged node's own moveMsgKindDrag handler
// (nodeMover.handle) does the rest, entirely on its own goroutine: commit its own new
// position (commitLocal — fan + persist + requantize, no cross-goroutine self-send).
// commitLocal's requantizeLocalPolars then sends every direct domain neighbor a single
// moveMsgKindNeighborSetC assignment (see that function's doc comment) — there is no
// equal-radii solve, no rule-node cascade, no gate-anchor broadcast; a drag never touches
// any node's position but its own. Returns false for an unknown node.
//
// NOTE: RootMove runs ONCE PER POINTER-MOVE EVENT, not once per drag (two bugs — commits
// 338f05da, 154a05bd — came from assuming otherwise; see
// memory/project_rootmove_is_per_pointer_move.md). The drag-log reset is NOT emitted
// here for that reason: the reset belongs at the real drag-start edge (the
// pending→dragging transition in gesture.go), not on every move tick RootMove sees.
func (lq *layoutQuantizer) RootMove(md *MoveDispatch, nodeID string, target vec3) bool {
	if _, ok := md.mr.nodeMovers[nodeID]; !ok {
		return false
	}
	// Route the drag itself to the dragged node's OWN inbox instead of committing on
	// the stdin reader's goroutine — every node's moveMsgKindDrag handler commits
	// (synchronous local apply, reported over reportCh) on its own goroutine. No
	// central commit call here.
	md.sendMove(nodeID, moveMsg{Kind: moveMsgKindDrag, NodeID: nodeID, Target: target})
	return true
}

// requantizeLocalPolars implements the cascade-link local-polar model on a drag: the
// dragged node X's new position gives each of its domain neighbors M a NEW distance to
// it. That distance is quantized to a whole tick on THAT neighbor's own small grid
// (layout_holder.go localStepTheta/localStepPhi/localStepR, or M's stored per-neighbor
// step constants) — and likewise X's own local polar TO M is requantized on X's own
// grid. The two ends' quantized values are independent and never reconciled or
// reconstructed from one another (MODEL.md "no blow-up, by construction" — this is the
// local-polar analogue: nothing rebuilds X's position from a local polar). Both ends'
// LayoutHolders are updated in memory and persisted.
//
// Decentralized (mirrors nodeMover.quantOffset): X requantizes+persists its OWN holder
// synchronously here, on its own (the mover's) goroutine — the single-writer case. Each
// domain neighbor M's own holder is written only by M's own goroutine: X sends M a
// single moveMsgKindNeighborSetC assignment (X's fresh newPos as FromCenter, X's newly
// requantized c to M as SnapC) instead of reaching into M's LayoutHolder directly, so a
// holder is mutated only by its own node's goroutine, exactly like quantOffset. M keeps
// its own stored bearing to X and repositions itself at the new distance along it (see
// neighborSetCReposition) — unconditional for every neighbor.
func (lq *layoutQuantizer) requantizeLocalPolars(md *MoveDispatch, nm *nodeMover, newPos vec3) {
	nodeID := nm.id
	lhX, okX := lq.layoutHolders[nodeID]
	if !okX {
		return
	}
	neighbors := map[string]bool{}
	for _, em := range md.mr.edgeMovers {
		if em.srcID == nodeID {
			neighbors[em.dstID] = true
		} else if em.dstID == nodeID {
			neighbors[em.srcID] = true
		}
	}
	if len(neighbors) == 0 {
		return
	}
	// X's local polars TO every reachable neighbor, resolved about X's rotating local
	// pole (rotating_pole.go) in ONE pass — the pole must see the WHOLE neighbor set, not
	// just one at a time, so a kick from one offset is checked against every other. cM is
	// read from nm's OWN partnerCenters map (kept current by each neighbor M's own
	// applyCenter push — see its doc comment); every m in neighbors is, by construction,
	// a key of nm.neighborIn, the same set partnerCenters is seeded/kept from.
	updatesX := map[string]vec3{}
	for m := range neighbors {
		cM, ok := nm.partnerCenters[m]
		if !ok {
			continue
		}
		updatesX[m] = cM.Sub(newPos)
	}
	if len(updatesX) == 0 {
		return
	}
	// oldByTo is X's DRAG-ANCHORED triple to each neighbor — the triple X had at the
	// START of the CURRENT drag, not at the previous move event. RootMove runs on every
	// ~8ms pointer-move, far finer than one quantize step (round(angle/step) lands on
	// the same integer for dozens of consecutive frames), so a per-move-event delta was
	// almost always (0,0,0) even mid-drag. The anchor is armed once, at the real
	// drag-start edge (gesture.go's gestPending->gestDragging transition sends
	// moveMsgKindDragStart to nodeID's own inbox, handled by armDragAnchor on this same
	// goroutine) so it accumulates across the whole drag and reads the drag's true total
	// on release. If no dragStart ever armed it (a programmatic RootMove with no
	// gesture, as several tests do), lazy-arm it right here from X's CURRENT
	// (pre-requantize) triple, so that first commit's delta is (0,0,0) and every later
	// commit in the same unarmed "drag" is relative to it — never a stale anchor from a
	// previous drag, since armDragAnchor always overwrites. Computed once, on X's own
	// goroutine — per CLAUDE.md's model (each goroutine reports what it itself picked
	// up). Pure observability: it does not gate or alter the requantize below in any way.
	if !nm.dragAnchorArmed {
		nm.armDragAnchor()
	}
	oldByTo := nm.dragAnchorByTo
	lq.requantizePoleTraced(lhX, updatesX)
	// X (nm) persists its OWN local-polars.json — this runs on X's own nodeMover
	// goroutine (requantizeLocalPolars is called from commitNodeMoveLocal, on nm's own
	// goroutine), so nm writes it directly (.claude/rules/persistence-ownership.md
	// "The model").
	nm.persistLocalPolars(lhX.LocalPolarsSnapshot(), lhX.Pole())

	// X tells EVERY direct domain neighbor M its NEW c (the quantized edge radius X just
	// requantized to M above) as a SINGLE ASSIGNMENT — moveMsgKindNeighborSetC. M keeps
	// its OWN stored bearing (QuantITheta/QuantIPhi) to X and repositions itself at the
	// new distance along that same stored direction, X held fixed; M does NOT
	// re-derive its bearing from a live offset and does NOT forward beyond this one
	// hop (neighborSetCReposition). Routed as a message on M's own directed inbound
	// channel (M's neighborIn[X] slot) instead of reaching into M's LayoutHolder from
	// X's (this) goroutine — each M's holder and center are written only by M's own
	// goroutine. Sent via X's OWN retry queue (nm.sendMove, the same
	// handle every other fan in this commit path uses — see commitNodeMoveLocal's
	// broadcastToEdgesAndPartners call above) instead of the direct-to-inbox sendMoveLossy
	// this used before: measured under the same mutually-adjacent concurrent-drag
	// flood TestMutuallyAdjacentDragFloodNoDeadlock drives, sendMoveLossy dropped
	// ~98% of NeighborSetC sends (9417/9600 in one run, TestNeighborSetCDropReachability)
	// — the "drop is safe, self-heals" justification was true in the sense that
	// nothing deadlocked, but the drop path was not a rare backstop, it was the
	// common case, silently discarding almost every NeighborSetC. Routing through
	// nodeID's own retry queue instead gets the SAME deadlock-safety property
	// sendMoveLossy was reaching for (decouples the send from this handler
	// goroutine, so two mutually-adjacent nodes committing concurrently can't block
	// each other) but via non-blocking send-and-retain-on-nm.pending, retried every
	// cycle of the sender's own clock-paced run loop, instead of a drop. Unconditional
	// for every neighbor — there is no rule/gate/anchor cascade left to defer to.
	enqueue := nm.sendMove
	lpByTo := map[string]wire.LocalPolar{}
	for _, lp := range lhX.LocalPolarsSnapshot() {
		lpByTo[lp.To] = lp
	}
	// The recipient set is THIS node's own stored cascadeEdges (nodes/<id>/
	// cascade-edges.json), NOT the domain-neighbor set updatesX was built from. There is
	// no behavior change: parseSpec's validateCascadeEdges now REQUIRES cascade adjacency
	// to equal domain adjacency, so the two sets are provably identical at load. The point
	// is removing the SECOND source of truth for "who is my neighbor" — the drag fan and
	// the delta-forward fan (forwardDelta) now read the same stored list, so they cannot
	// drift apart. They did drift, and it was a real bug: dragging node 8 reached node 5
	// over the domain edge 5-8 while node 5's cascade-edges.json had no entry for 8, so
	// forwardDelta read the sender's kind as "" and the Pulse gate-routing rule fell
	// through to a fan.
	//
	// X still re-quantizes its OWN triple to every domain neighbor above
	// (requantizePoleTraced takes updatesX unchanged); only the outbound assignment is
	// scoped here. A cascade neighbor X did not re-quantize toward on this commit has no
	// lpByTo entry and is skipped.
	for _, m := range nm.cascadeEdges {
		newLP, ok := lpByTo[m]
		if !ok {
			continue
		}
		// X's own triple-change for its edge to m: new - old, or (0,0,0) if X had no
		// prior stored triple to m (nothing to subtract from — see requantizeLocalPolars
		// doc). Computed once here, on X's own goroutine, and carried unmodified to
		// EVERY recipient — not recomputed per-recipient.
		var deltaA, deltaB, deltaC int
		if oldLP, hadOld := oldByTo[m]; hadOld {
			deltaA = newLP.QuantITheta - oldLP.QuantITheta
			deltaB = newLP.QuantIPhi - oldLP.QuantIPhi
			deltaC = newLP.QuantIR - oldLP.QuantIR
		}
		enqueue(m, moveMsg{Kind: moveMsgKindNeighborSetC, NodeID: m, SenderID: nodeID, FromCenter: newPos,
			DeltaA: deltaA, DeltaB: deltaB, DeltaC: deltaC})
	}
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
