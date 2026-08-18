package nodeactor

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/framegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/nodeframe"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodedrag"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
	"github.com/dtauraso/wirefold/nodes/rowevent"

	T "github.com/dtauraso/wirefold/tools/topology-vscode/Trace"
)

func boolU8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

func hasKindRuleU8(kind string) uint8 {
	if nodedrag.HasKindRule(kind) {
		return 1
	}
	return 0
}

func (m *NodeGeometry) emitGeometry() {

	m.writeStreamFrame([]rowevent.RowEvent{{
		Kind: T.KindNodeGeometry, NodeRow: m.stream.NodeRow(),
		PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1,
	}})
}

func (m *NodeGeometry) postSelfEvents(events []rowevent.RowEvent) {
	m.stream.PostSelfEvents(events)
}

func (m *NodeGeometry) drainSelfEvents() []rowevent.RowEvent {
	return m.stream.DrainSelfEvents()
}

func (m *NodeGeometry) writeStreamFrame(events []rowevent.RowEvent) {
	if !m.stream.Ready() {
		return
	}

	row := m.stream.NodeRow()
	for _, e := range events {
		if e.NodeRow != row {
			panic(fmt.Sprintf(
				"NodeGeometry.writeStreamFrame: node %q (row %d) is carrying a %s event for row %d on its OWN dedicated stream — NodeRow is an ownership claim, not a reference; a foreign node belongs in TargetRow",
				m.id, row, e.Kind, e.NodeRow))
		}
	}

	coplanarEdges, upAxis := m.flags.Flags()
	topIdx, bottomIdx, normalIdx, receivedIdx, receivedSet, latticePoints := m.tilt.FrameGeometryFields()
	fg := framegeom.DeriveFrameGeometry(framegeom.FrameGeometryInputs{
		Geom:                 m.geom,
		UpAxis:               upAxis,
		CoplanarEdges:        coplanarEdges,
		TopTiltVectorPhiIdx:  topIdx,
		BottomPhiIdx:         bottomIdx,
		NormalPhiIdx:         normalIdx,
		ReceivedVectorPhiIdx: receivedIdx,
		ReceivedVectorSet:    receivedSet,
		LatticePoints:        latticePoints,
		DefaultLatticePoints: tiltvector.FullTurnPhiIdx,
	})
	center := fg.Center
	polePhi, poleTheta := fg.PolePhi, fg.PoleTheta
	ringAxisPhi, ringAxisTheta := fg.RingAxisPhi, fg.RingAxisTheta
	ringMatrix := fg.RingMatrix
	points := fg.LatticePoints
	topTiltVectorLen := fg.TopTiltVectorLen
	receivedVectorLen := fg.ReceivedVectorLen
	label := m.geom.Label
	if label == "" {
		label = m.id
	}
	selected, hovered, latchedSel := m.ui.Flags()
	kindID := m.stream.KindID()
	roundsToParallel, msgsToParallel := m.readout.RoundsToParallel()

	var dragRLocked, dragPhiLocked uint8
	dragThetaMax := float32(-1)
	if rule := m.DragRule(); rule != nil {
		if rule.R != nil {
			dragRLocked = 1
		}
		if rule.Phi != nil {
			dragPhiLocked = 1
		}
		if rule.MaxTheta != nil {
			dragThetaMax = float32(*rule.MaxTheta)
		}
	}

	var selfRLocked, selfPhiLocked uint8
	selfThetaMax := float32(-1)
	if rule := m.SelfRule(); rule != nil {
		if rule.R != nil {
			selfRLocked = 1
		}
		if rule.Phi != nil {
			selfPhiLocked = 1
		}
		if rule.MaxTheta != nil {
			selfThetaMax = float32(*rule.MaxTheta)
		}
	}
	dragActive := uint8(0)
	if m.DragRuleActive() {
		dragActive = 1
	}
	ruleGroupID, ruleGroupSize := m.RuleGroup()

	m.stream.WriteFrame(nodeframe.NodeFrameInput{
		Tick:                uint32(m.clocks.Tick()),
		NodeRow:             row,
		NodeID:              row + 1,
		CX:                  float32(center.X),
		CY:                  float32(center.Y),
		CZ:                  float32(center.Z),
		Radius:              float32(nodegeom.NodeRadius(m.geom.Kind)),
		VRX:                 loadspec.VerticalRingNormalX,
		VRY:                 loadspec.VerticalRingNormalY,
		VRZ:                 loadspec.VerticalRingNormalZ,
		FRX:                 loadspec.FlatRingNormalX,
		FRY:                 loadspec.FlatRingNormalY,
		FRZ:                 loadspec.FlatRingNormalZ,
		PoleRingR:           float32(nodegeom.PoleRingR()),
		PolePhi:             float32(polePhi),
		PoleTheta:           float32(poleTheta),
		RingAxisPhi:         float32(ringAxisPhi),
		RingAxisTheta:       float32(ringAxisTheta),
		RingMatrix:          ringMatrix,
		TopTiltVectorLen:    float32(topTiltVectorLen),
		TopTiltVectorIdx:    fg.TopTiltVectorIdx,
		TopTiltVectorPhi:    float32(fg.TopTiltVectorPhi),
		BottomTiltVectorPhi: float32(fg.BottomTiltVectorPhi),
		CoplanarNormalPhi:   float32(fg.CoplanarNormalPhi),
		ReceivedVectorLen:   float32(receivedVectorLen),
		ReceivedVectorPhi:   float32(fg.ReceivedVectorPhi),
		Selected:            selected,
		KindID:              kindID,
		Hovered:             hovered,
		LatchedSel:          latchedSel,
		LatticePoints:       uint8(points),
		RoundsToParallel:    roundsToParallel,
		MsgsToParallel:      msgsToParallel,
		HasKindRule:         hasKindRuleU8(m.SelfKind()),
		KindRuleActive:      boolU8(m.KindRuleActive()),
		SelfRLocked:         selfRLocked,
		SelfPhiLocked:       selfPhiLocked,
		SelfThetaMax:        selfThetaMax,
		SelfActive:          boolU8(m.SelfRuleActive()),
		RuleGroupID:         ruleGroupID,
		RuleGroupSize:       ruleGroupSize,
		DragRLocked:         dragRLocked,
		DragPhiLocked:       dragPhiLocked,
		DragThetaMax:        dragThetaMax,
		DragActive:          dragActive,
		Label:               label,
		Events:              events,
	})
}
