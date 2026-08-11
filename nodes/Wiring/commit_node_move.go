// commit_node_move.go — the owner-goroutine single-node commit path.
//
// Split out of quantized_move.go (god-object decomposition, pure move — no logic
// changes): kept apart from held-state snapshots, touching-bead resolution, and the
// broadcast-after-commit fan-out.

package Wiring

import (
	"fmt"
	"strings"

	"github.com/dtauraso/wirefold/nodes/Wiring/beadcrud"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/Wiring/topoderive"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"

	T "github.com/dtauraso/wirefold/Trace"
)

// commitNodeMoveLocal is the OWNER-GOROUTINE single-node commit path
// (generalized to every node): used when the commit
// originates on nodeID's OWN mover goroutine (its own inbox handler for a
// movemsg.KindDrag). It applies nodeID's OWN new center SYNCHRONOUSLY via
// applyCenter — safe and correct here because applyCenter's doc contract is "called
// only from this nodeMover's own inbox-drain goroutine", which this is. Also fans
// centers to incident edges/partners, persists the per-node quantized-offset
// (nodeMover.quantOffset — never a shared map, so no other mover goroutine's commit
// can race this write even for a different node id), and requantizes nodeID's
// local-polar cascade-links against its (unmoved) neighbors.
func (lq *layoutQuantizer) commitNodeMoveLocal(mr *moverRegistry, ui *viewstate.UIState, nm *nodeactor.NodeGeometry, newPos vec3) {
	nodeID := nm.ID()
	edges := lq.heldEdges(mr)
	// reach[nodeID] only ever needs nodeID's own fresh polar plus its DIRECT
	// neighbors' polar (reachRFromPolar only accumulates reach for an edge's
	// SOURCE, from that edge's Target) — each direct neighbor's last-pushed
	// CARTESIAN center is read from THIS node's OWN partnerCenters map (nm.
	// PartnerCenters(), kept current by every neighbor's ApplyCenter push — see its
	// doc comment), resolved via nm.EdgeIDs() (this node's own incident edges, fixed
	// at construction; every edgeIDs neighbor is by construction a key of
	// nm's own neighborIn, the same set partnerCenters is seeded/kept from). scene polar
	// is a pure re-derive off the fixed, write-once ui.SceneSphere.Center (never
	// mutated after load), so this stays race-free with no cross-goroutine read at
	// all now (this runs on nm's own goroutine, reading nm's own map).
	polars := map[string]geom.Polar{}
	partnerCenters := nm.PartnerCenters()
	for _, edgeID := range nm.EdgeIDs() {
		em, ok := mr.edgeMovers[edgeID]
		if !ok {
			continue
		}
		neighborID := em.SrcID()
		if neighborID == nodeID {
			neighborID = em.DstID()
		}
		if c, ok := partnerCenters[neighborID]; ok {
			polars[neighborID] = geom.Cart2polar(c.Sub(ui.SceneSphere.Center))
		}
	}
	// Single cart2polar boundary conversion for this drag target — newPos is mouse-
	// derived cartesian (gesture.go ray/plane unproject); everything downstream
	// (reach, measureScalar, the persist schedule) reuses this one polar value rather
	// than re-deriving it from newPos.
	nodePolar := geom.Cart2polar(newPos.Sub(ui.SceneSphere.Center))

	// committedPos/committedPolar are what gets DRAWN (applyCenter), FANNED
	// (broadcastToEdgesAndPartners), PERSISTED (persistQuantOffset), and re-quantized
	// against by every neighbor (requantizeLocalPolars) for this commit — ONE position,
	// not the raw drag target for some of those and a quantized point for others
	// (docs/investigations/which-lattice-a-node-lives-on.md "Why the drag makes it worst": that split is
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
	if lq.quantizedLayout {
		prevPos := nm.WorldCenter()
		beads := dragTouchingBeads(mr, nm, prevPos)
		if len(beads) == 0 {
			committedPos = newPos
		} else {
			committedPos, _ = beadcrud.ResolveBeadCrudMove(beads, prevPos, newPos, lattice.BeadStepR)
		}
		committedPolar = geom.Cart2polar(committedPos.Sub(ui.SceneSphere.Center))

		// DIAGNOSTIC ONLY (task/log-node2-bead-crud): one breadcrumb per pointer-move
		// commit — node 2 (neighbours 1, 4, 5) can barely be dragged; long drags produce
		// no movement, and beads that should be ADDED to push it the right way are not
		// being added. This packs the whole event PLUS every touching bead's own CRUD
		// arithmetic (why it returned none/add/remove) so the actual numbers, not a
		// theory, explain it. Gated on nm.tr != nil exactly like the neighbor-center-recv/
		// neighbor-setc-recv breadcrumb sites above (node_mover.go) — cheap no-op with no
		// stream wired (headless tests, bare movers).
		if nm.Traced() {
			dragVector := newPos.Sub(prevPos)
			parts := make([]string, 0, len(beads))
			for _, b := range beads {
				diag := beadcrud.BeadCrudDiagnose(b.NeighborID, b.Source, b.Centre, b.AimDir, prevPos, newPos, dragVector, lattice.BeadStepR)
				verdictStr := "none"
				switch diag.Verdict {
				case beadcrud.BeadCrudAdd:
					verdictStr = "add"
				case beadcrud.BeadCrudRemove:
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
			nm.Breadcrumb("bead-crud", nodeID, "", value)
			nm.WriteStreamFrame([]wire.RowEvent{{
				Kind: T.KindBreadcrumb, Label: T.BreadcrumbBeadCrud, Debug: 1,
				NodeRow: nm.NodeRow(), PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
				Text: value,
			}})
		}
	}

	polars[nodeID] = committedPolar
	reach := topoderive.ReachRFromPolar(polars, edges)

	nm.ApplyCenter(committedPos, reach[nodeID])
	lq.broadcastToEdgesAndPartners(mr, map[string]vec3{nodeID: committedPos}, nm.SendMove())

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
	nm.CommitQuantOffset(committedPolar)
}
