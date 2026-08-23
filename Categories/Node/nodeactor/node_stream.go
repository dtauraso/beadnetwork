package nodeactor

import (
	"fmt"

	"github.com/dtauraso/wirefold/Categories/Node/nodeframe"
	"github.com/dtauraso/wirefold/Categories/Node/owners"
)

func (m *NodeGeometry) emitGeometry() {

	m.writeStreamFrame(nil)
}

func (m *NodeGeometry) postSelfEvents(events []owners.RowEvent) { m.trace.Post(events) }

func (m *NodeGeometry) drainSelfEvents() []owners.RowEvent { return m.trace.Drain() }

func (m *NodeGeometry) writeStreamFrame(events []owners.RowEvent) {
	if !m.stream.Ready() {
		return
	}
	m.trace.Append(events)

	row := m.stream.NodeRow()
	for _, e := range events {
		if e.NodeRow != row {
			panic(fmt.Sprintf(
				"NodeGeometry.writeStreamFrame: node %q (row %d) is carrying a %s event for row %d on its OWN dedicated stream — NodeRow is an ownership claim, not a reference; a foreign node belongs in TargetRow",
				m.id, row, e.Kind, e.NodeRow))
		}
	}

	m.stream.WriteFrame(nodeframe.BuildFrame(m.frameInputs(row)))
}

func (m *NodeGeometry) frameInputs(row int32) nodeframe.FrameInputs {
	coplanarEdges, upAxis := m.flags.Flags()
	topIdx, bottomIdx, normalIdx, receivedIdx, receivedSet, latticePoints := m.tilt.FrameGeometryFields()
	selected, hovered, latchedSel := m.ui.Flags()
	roundsToParallel, msgsToParallel := m.readout.RoundsToParallel()
	ruleGroupID, ruleGroupSize := m.RuleGroup()

	return nodeframe.FrameInputs{
		Geom: m.geom,

		Row:    row,
		KindID: m.stream.KindID(),
		Tick:   uint32(m.clocks.Tick()),
		ID:     m.id,

		UpAxis:        upAxis,
		CoplanarEdges: coplanarEdges,

		TopTiltVectorPhiIdx:  topIdx,
		BottomPhiIdx:         bottomIdx,
		NormalPhiIdx:         normalIdx,
		ReceivedVectorPhiIdx: receivedIdx,
		ReceivedVectorSet:    receivedSet,
		LatticePoints:        latticePoints,

		Selected:   selected,
		Hovered:    hovered,
		LatchedSel: latchedSel,

		RoundsToParallel: roundsToParallel,
		MsgsToParallel:   msgsToParallel,

		DragRule: m.topo.DragRule(),
		SelfRule: m.topo.SelfRule(),

		DragActive:     m.topo.DragRuleActive(),
		SelfActive:     m.topo.SelfRuleActive(),
		KindRuleActive: m.KindRuleActive(),
		HasKindRule:    m.HasKindRule(),

		RuleGroupID:   ruleGroupID,
		RuleGroupSize: ruleGroupSize,

		ChannelVectors: m.channelVectors(),
	}
}
