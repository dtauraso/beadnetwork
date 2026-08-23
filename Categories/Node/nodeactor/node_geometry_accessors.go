package nodeactor

import (
	"github.com/dtauraso/wirefold/Categories/Node/nodeactor/owners"
	"slices"

	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/PolarRulesPanel"

	"github.com/dtauraso/wirefold/Categories/Node/nodegeom"
	"github.com/dtauraso/wirefold/Categories/Polar/polar"
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
)

func (m *NodeGeometry) Topo() *owners.Topology { return &m.topo }

func (m *NodeGeometry) ID() string { return m.id }

func (m *NodeGeometry) Kind() string { return m.geom.Kind }

func (m *NodeGeometry) SelfKind() string { return m.selfKind }

func (m *NodeGeometry) DragRule() *PolarRulesPanel.DragRule { return m.topo.DragRule() }

func (m *NodeGeometry) DragRuleActive() bool { return m.topo.DragRuleActive() }

func (m *NodeGeometry) SelfRule() *PolarRulesPanel.DragRule { return m.topo.SelfRule() }

func (m *NodeGeometry) SelfRuleActive() bool { return m.topo.SelfRuleActive() }

func (m *NodeGeometry) Tick() int64 { return m.clocks.Tick() }

func (m *NodeGeometry) Label() string { return m.geom.Label }

func (m *NodeGeometry) NodeRow() int32 { return m.stream.NodeRow() }

func (m *NodeGeometry) EdgeIDs() []string { return m.topo.EdgeIDs() }

func (m *NodeGeometry) NeighborKinds() map[string]string { return m.topo.NeighborKinds() }

func (m *NodeGeometry) Deltas() *owners.Deltas { return &m.deltas }

func (m *NodeGeometry) Anim() *NodeBeadAnimation { return m.anim }

func (m *NodeGeometry) OutEdges() *owners.OutEdges { return &m.outEdges }

func (m *NodeGeometry) Tilt() *owners.Tilt { return &m.tilt }

func (m *NodeGeometry) Channels() *owners.ChannelVectors { return &m.channels }

func (m *NodeGeometry) ScenePolar() polar.Polar { return nodegeom.ScenePolarOf(m.geom) }

func (m *NodeGeometry) ComposedIndex() polarindex.Index { return nodegeom.ComposedIndexOf(m.geom) }

func (m *NodeGeometry) Constants() polarindex.SceneConstants { return m.geom.SceneConstants }

func (m *NodeGeometry) SceneCenter() Vec3 { return Vec3(m.geom.SceneCenter) }

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

func (m *NodeGeometry) SendMove() func(id string, msg owners.Msg) { return m.msg.SendMove() }

func (m *NodeGeometry) CommitIndex() {
	m.persistIndex(m.geom.DragIndex)
}

func (m *NodeGeometry) WriteStreamFrame(events []owners.RowEvent) {
	m.writeStreamFrame(events)
}
