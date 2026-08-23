package Node

import (
	"github.com/dtauraso/wirefold/Categories/Node/owners"
	"slices"

	"github.com/dtauraso/wirefold/Categories/Node/nodegeom"
	"github.com/dtauraso/wirefold/Categories/Polar/polar"
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
)

func (m *NodeGeometry) Topo() *owners.Topology { return &m.topo }

func (m *NodeGeometry) Msg() *owners.Messaging { return &m.msg }

func (m *NodeGeometry) Flags() *owners.Flags { return &m.flags }

func (m *NodeGeometry) Clocks() *owners.Clocks { return &m.clocks }

func (m *NodeGeometry) Beads() *owners.Beads { return &m.beads }

func (m *NodeGeometry) KindPosts() *owners.KindPosts { return &m.kindPosts }

func (m *NodeGeometry) Trace() *owners.Trace { return &m.trace }

func (m *NodeGeometry) Stream() *owners.Stream { return &m.stream }

func (m *NodeGeometry) ID() string { return m.id }

func (m *NodeGeometry) Kind() string { return m.geom.Kind }

func (m *NodeGeometry) SelfKind() string { return m.selfKind }

func (m *NodeGeometry) Deltas() *owners.Deltas { return &m.deltas }

func (m *NodeGeometry) Anim() *owners.NodeBeadAnimation { return m.anim }

func (m *NodeGeometry) OutEdges() *owners.OutEdges { return &m.outEdges }

func (m *NodeGeometry) Tilt() *owners.Tilt { return &m.tilt }

func (m *NodeGeometry) Channels() *owners.ChannelVectors { return &m.channels }

func (m *NodeGeometry) ScenePolar() polar.Polar { return nodegeom.ScenePolarOf(m.geom) }

func (m *NodeGeometry) ComposedIndex() polarindex.Index { return nodegeom.ComposedIndexOf(m.geom) }

func (m *NodeGeometry) Constants() polarindex.SceneConstants { return m.geom.SceneConstants }

func (m *NodeGeometry) SceneCenter() nodegeom.Vec3 { return m.geom.SceneCenter }

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

func (m *NodeGeometry) CommitIndex() {
	m.persistIndex(m.geom.DragIndex)
}
