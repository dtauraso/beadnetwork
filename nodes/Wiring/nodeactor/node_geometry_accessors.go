package nodeactor

import (
	"github.com/dtauraso/wirefold/src/PolarRulesPanel"
	"slices"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/polarindex"
	"github.com/dtauraso/wirefold/nodes/rowevent"
)

func (m *NodeGeometry) ID() string { return m.id }

func (m *NodeGeometry) Traced() bool { return m.tr != nil }

func (m *NodeGeometry) Breadcrumb(label, node, port, value string) {
	if m.tr != nil {
		m.tr.Breadcrumb(label, node, port, value)
	}
}

func (m *NodeGeometry) Kind() string { return m.geom.Kind }

func (m *NodeGeometry) SelfKind() string { return m.selfKind }

func (m *NodeGeometry) DragRule() *PolarRulesPanel.DragRule { return m.topo.DragRule() }

func (m *NodeGeometry) DragRuleActive() bool { return m.topo.DragRuleActive() }

func (m *NodeGeometry) SelfRule() *PolarRulesPanel.DragRule { return m.topo.SelfRule() }

func (m *NodeGeometry) SelfRuleActive() bool { return m.topo.SelfRuleActive() }

func (m *NodeGeometry) Tick() int64 { return m.clocks.Tick() }

func (m *NodeGeometry) Label() string { return m.geom.Label }

func (m *NodeGeometry) WorldCenter() vec3 { return nodegeom.NodeWorldPos(m.geom) }

func (m *NodeGeometry) NodeRow() int32 { return m.stream.NodeRow() }

func (m *NodeGeometry) EdgeIDs() []string { return m.topo.EdgeIDs() }

func (m *NodeGeometry) NeighborKinds() map[string]string { return m.topo.NeighborKinds() }

func (m *NodeGeometry) OutTargets() []string { return m.outTargets }

func (m *NodeGeometry) SetBaseDeltaTo(otherID string, off polarindex.Offset) {
	m.deltas.SetBaseDeltaTo(otherID, off)
}

func (m *NodeGeometry) SetDragDeltaTo(otherID string, off polarindex.Offset) {
	m.deltas.SetDragDeltaTo(otherID, off)
}

func (m *NodeGeometry) DeltaTo(otherID string) (polarindex.Offset, bool) {
	return m.deltas.DeltaTo(otherID)
}

func (m *NodeGeometry) DeltaFrom(otherID string) (polarindex.Offset, bool) {
	return m.deltas.DeltaFrom(otherID)
}

func (m *NodeGeometry) ShiftDeltasBy(delta polarindex.Offset) { m.deltas.ShiftSelfBy(delta) }

func (m *NodeGeometry) ScenePolar() polar.Polar { return nodegeom.ScenePolarOf(m.geom) }

func (m *NodeGeometry) ComposedIndex() polarindex.Index { return nodegeom.ComposedIndexOf(m.geom) }

func (m *NodeGeometry) Constants() polarindex.SceneConstants { return m.geom.SceneConstants }

func (m *NodeGeometry) SceneCenter() vec3 { return m.geom.SceneCenter }

func (m *NodeGeometry) IsOutTarget(neighborID string) bool {
	return slices.Contains(m.outTargets, neighborID)
}

func (m *NodeGeometry) KindRuleActive() bool {
	if m.RuleNode() == nil {
		return true
	}
	return m.RuleNode().KindActive()
}

func (m *NodeGeometry) EdgeRuleActive(otherID string) bool {
	if m.RuleNode() == nil {
		return true
	}
	return m.RuleNode().EdgeActive(otherID)
}

func (m *NodeGeometry) SendMove() func(id string, msg movemsg.Msg) { return m.msg.SendMove() }

func (m *NodeGeometry) NeighborIDs() []string { return m.msg.NeighborIDs() }

func (m *NodeGeometry) CommitIndex() {
	m.persistIndex(m.geom.DragIndex)
}

func (m *NodeGeometry) WriteStreamFrame(events []rowevent.RowEvent) {
	m.writeStreamFrame(events)
}
