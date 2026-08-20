package nodeactor

import (
	"fmt"
	"math"

	"github.com/dtauraso/wirefold/src/Node/Wiring/framegeom"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodeactor/nodeframe"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodedrag"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/src/Node/rowevent"
	"github.com/dtauraso/wirefold/src/Chrome/TiltPanel"
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

	m.writeStreamFrame(nil)
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
		DefaultLatticePoints: TiltPanel.FullTurnPhiIdx,
	})
	composedIdx := nodegeom.ComposedIndexOf(m.geom)
	polePhi, poleTheta := fg.PolePhi, fg.PoleTheta
	ringMatrix := fg.RingMatrix
	points := fg.LatticePoints
	topTiltVectorLen := fg.TopTiltVectorLen
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
		Tick:             uint32(m.clocks.Tick()),
		NodeRow:          row,
		NodeID:           row + 1,
		IndexR:           int32(composedIdx.R),
		IndexPhi:         int32(composedIdx.Phi),
		IndexTheta:       int32(composedIdx.Theta),
		HasPos:           boolU8(m.geom.HasPos),
		Radius:           float32(nodegeom.NodeRadius(m.geom.Kind)),
		NavTubeR:         float32(math.Max(0.5, nodegeom.NodeRadius(m.geom.Kind)*0.08)),
		PoleAnchorX:      float32(fg.Center.X),
		PoleAnchorY:      float32(fg.Center.Y),
		PoleAnchorZ:      float32(fg.Center.Z),
		LabelAnchorX:     float32(fg.LabelAnchor.X),
		LabelAnchorY:     float32(fg.LabelAnchor.Y),
		LabelAnchorZ:     float32(fg.LabelAnchor.Z),
		PoleRingR:        float32(nodegeom.PoleRingR()),
		PolePhi:          float32(polePhi),
		PoleTheta:        float32(poleTheta),
		RingMatrix:       ringMatrix,
		TopTiltVectorLen: float32(topTiltVectorLen),
		TopTiltVectorIdx: fg.TopTiltVectorIdx,
		TiltArrows:       fg.TiltArrows,
		ChannelVectors:   m.channelVectors(),
		Selected:         selected,
		KindID:           kindID,
		Hovered:          hovered,
		LatchedSel:       latchedSel,
		LatticePoints:    uint8(points),
		RoundsToParallel: roundsToParallel,
		MsgsToParallel:   msgsToParallel,
		HasKindRule:      hasKindRuleU8(m.SelfKind()),
		KindRuleActive:   boolU8(m.KindRuleActive()),
		SelfRLocked:      selfRLocked,
		SelfPhiLocked:    selfPhiLocked,
		SelfThetaMax:     selfThetaMax,
		SelfActive:       boolU8(m.SelfRuleActive()),
		RuleGroupID:      ruleGroupID,
		RuleGroupSize:    ruleGroupSize,
		DragRLocked:      dragRLocked,
		DragPhiLocked:    dragPhiLocked,
		DragThetaMax:     dragThetaMax,
		DragActive:       dragActive,
		Label:            label,
		Events:           events,
	})
}
