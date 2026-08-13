package layoutquant

import (
	"fmt"
	"strings"

	"github.com/dtauraso/wirefold/nodes/Wiring/beadcrud"
	"github.com/dtauraso/wirefold/nodes/Wiring/edgemover"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/Wiring/topoderive"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
	"github.com/dtauraso/wirefold/nodes/rowevent"
	"github.com/dtauraso/wirefold/nodes/spatial"
	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"

	T "github.com/dtauraso/wirefold/Trace"
)

func (lq *LayoutQuantizer) CommitNodeMoveLocal(nodeGeoms map[string]*nodeactor.NodeGeometry, edgeMovers map[string]*edgemover.EdgeMover, ui *viewstate.UIState, nm *nodeactor.NodeGeometry, newPos spatial.Vec3) {
	nodeID := nm.ID()
	edges := HeldEdges(edgeMovers)
	polars := neighborPolars(nm, edgeMovers, ui)

	nodePolar := polar.Cart2polar(newPos.Sub(ui.SceneSphere.Center))

	committedPos, committedPolar := lq.resolveCommittedPosition(edgeMovers, ui, nm, newPos, nodePolar)

	polars[nodeID] = committedPolar
	reach := topoderive.ReachRFromPolar(polars, edges)

	nm.ApplyCenter(committedPos, reach[nodeID])
	BroadcastToEdgesAndPartners(edgeMovers, nodeGeoms, map[string]spatial.Vec3{nodeID: committedPos}, nm.SendMove())

	nm.CommitQuantOffset(committedPolar)
}

func neighborPolars(nm *nodeactor.NodeGeometry, edgeMovers map[string]*edgemover.EdgeMover, ui *viewstate.UIState) map[string]polar.Polar {
	nodeID := nm.ID()
	polars := map[string]polar.Polar{}
	partnerCenters := nm.PartnerCenters()
	for _, edgeID := range nm.EdgeIDs() {
		em, ok := edgeMovers[edgeID]
		if !ok {
			continue
		}
		neighborID := em.SrcID()
		if neighborID == nodeID {
			neighborID = em.DstID()
		}
		if c, ok := partnerCenters[neighborID]; ok {
			polars[neighborID] = polar.Cart2polar(c.Sub(ui.SceneSphere.Center))
		}
	}
	return polars
}

func (lq *LayoutQuantizer) resolveCommittedPosition(edgeMovers map[string]*edgemover.EdgeMover, ui *viewstate.UIState, nm *nodeactor.NodeGeometry, newPos spatial.Vec3, nodePolar polar.Polar) (committedPos spatial.Vec3, committedPolar polar.Polar) {
	committedPos = HoldAgainstInNeighbors(nm, newPos)
	committedPolar = nodePolar
	if committedPos != newPos {
		committedPolar = polar.Cart2polar(committedPos.Sub(ui.SceneSphere.Center))
	}
	if !lq.QuantizedLayout {
		return committedPos, committedPolar
	}
	newPos = committedPos
	prevPos := nm.WorldCenter()
	beads := DragTouchingBeads(edgeMovers, nm, prevPos)
	if len(beads) == 0 {
		committedPos = newPos
	} else {
		committedPos, _ = beadcrud.ResolveBeadCrudMove(beads, prevPos, newPos, lattice.BeadStepR)
	}
	// Bead snapping moves the destination again, so the hold is the LAST
	// thing applied — otherwise the snap is free to put the node back off
	// the angles, which is the frame-late correction this exists to avoid.
	committedPos = HoldAgainstInNeighbors(nm, committedPos)
	committedPolar = polar.Cart2polar(committedPos.Sub(ui.SceneSphere.Center))

	emitBeadCrudDiagnostic(nm, nm.ID(), prevPos, newPos, committedPos, beads)
	return committedPos, committedPolar
}

func emitBeadCrudDiagnostic(nm *nodeactor.NodeGeometry, nodeID string, prevPos, newPos, committedPos spatial.Vec3, beads []beadcrud.TouchingBead) {
	if !nm.Traced() {
		return
	}
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
	nm.WriteStreamFrame([]rowevent.RowEvent{{
		Kind: T.KindBreadcrumb, Label: T.BreadcrumbBeadCrud, Debug: 1,
		NodeRow: nm.NodeRow(), PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Text: value,
	}})
}
