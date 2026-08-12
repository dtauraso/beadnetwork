package layoutquant

import (
	"fmt"
	"strings"

	"github.com/dtauraso/wirefold/nodes/Wiring/beadcrud"
	"github.com/dtauraso/wirefold/nodes/Wiring/edgemover"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/Wiring/topoderive"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"

	T "github.com/dtauraso/wirefold/Trace"
)

func (lq *LayoutQuantizer) CommitNodeMoveLocal(nodeGeoms map[string]*nodeactor.NodeGeometry, edgeMovers map[string]*edgemover.EdgeMover, ui *viewstate.UIState, nm *nodeactor.NodeGeometry, newPos wire.Vec3) {
	nodeID := nm.ID()
	edges := HeldEdges(edgeMovers)
	polars := neighborPolars(nm, edgeMovers, ui)

	nodePolar := geom.Cart2polar(newPos.Sub(ui.SceneSphere.Center))

	committedPos, committedPolar := lq.resolveCommittedPosition(edgeMovers, ui, nm, newPos, nodePolar)

	polars[nodeID] = committedPolar
	reach := topoderive.ReachRFromPolar(polars, edges)

	nm.ApplyCenter(committedPos, reach[nodeID])
	BroadcastToEdgesAndPartners(edgeMovers, nodeGeoms, map[string]wire.Vec3{nodeID: committedPos}, nm.SendMove())

	nm.CommitQuantOffset(committedPolar)
}

func neighborPolars(nm *nodeactor.NodeGeometry, edgeMovers map[string]*edgemover.EdgeMover, ui *viewstate.UIState) map[string]geom.Polar {
	nodeID := nm.ID()
	polars := map[string]geom.Polar{}
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
			polars[neighborID] = geom.Cart2polar(c.Sub(ui.SceneSphere.Center))
		}
	}
	return polars
}

func (lq *LayoutQuantizer) resolveCommittedPosition(edgeMovers map[string]*edgemover.EdgeMover, ui *viewstate.UIState, nm *nodeactor.NodeGeometry, newPos wire.Vec3, nodePolar geom.Polar) (committedPos wire.Vec3, committedPolar geom.Polar) {
	committedPos = newPos
	committedPolar = nodePolar
	if !lq.QuantizedLayout {
		return committedPos, committedPolar
	}
	prevPos := nm.WorldCenter()
	beads := DragTouchingBeads(edgeMovers, nm, prevPos)
	if len(beads) == 0 {
		committedPos = newPos
	} else {
		committedPos, _ = beadcrud.ResolveBeadCrudMove(beads, prevPos, newPos, lattice.BeadStepR)
	}
	committedPolar = geom.Cart2polar(committedPos.Sub(ui.SceneSphere.Center))

	emitBeadCrudDiagnostic(nm, nm.ID(), prevPos, newPos, committedPos, beads)
	return committedPos, committedPolar
}

func emitBeadCrudDiagnostic(nm *nodeactor.NodeGeometry, nodeID string, prevPos, newPos, committedPos wire.Vec3, beads []beadcrud.TouchingBead) {
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
	nm.WriteStreamFrame([]wire.RowEvent{{
		Kind: T.KindBreadcrumb, Label: T.BreadcrumbBeadCrud, Debug: 1,
		NodeRow: nm.NodeRow(), PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Text: value,
	}})
}
