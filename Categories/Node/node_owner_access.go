package Node

import (
	beadanimation "github.com/dtauraso/wirefold/Categories/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/Categories/Node/ChannelVectors"
	"github.com/dtauraso/wirefold/Categories/Node/TiltVectors"
	"slices"

	"github.com/dtauraso/wirefold/Categories/Polar/polar"
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
)

func (m *NodeGeometry) Topo() *Topology { return &m.topo }

func (m *NodeGeometry) Msg() *Messaging { return &m.msg }

func (m *NodeGeometry) Flags() *Flags { return &m.flags }

func (m *NodeGeometry) Clocks() *Clocks { return &m.clocks }

func (m *NodeGeometry) Beads() *beadanimation.Beads { return &m.beads }

func (m *NodeGeometry) KindPosts() *KindPosts { return &m.kindPosts }

func (m *NodeGeometry) Trace() *Trace { return &m.trace }

func (m *NodeGeometry) Stream() *Stream { return &m.stream }

func (m *NodeGeometry) ID() string { return m.id }

func (m *NodeGeometry) Kind() string { return m.geom.Kind }

func (m *NodeGeometry) SelfKind() string { return m.selfKind }

func (m *NodeGeometry) Deltas() *Deltas { return &m.deltas }

func (m *NodeGeometry) Anim() *NodeBeadAnimation { return m.anim }

func (m *NodeGeometry) OutEdges() *OutEdges { return &m.outEdges }

func (m *NodeGeometry) Tilt() *TiltVectors.Tilt { return &m.tilt }

func (m *NodeGeometry) Channels() *ChannelVectors.PeerCenters { return &m.channels }

func (m *NodeGeometry) ScenePolar() polar.Polar { return ScenePolarOf(m.geom) }

func (m *NodeGeometry) ComposedIndex() polarindex.Index { return ComposedIndexOf(m.geom) }

func (m *NodeGeometry) Constants() polarindex.SceneConstants { return m.geom.SceneConstants }

func (m *NodeGeometry) SceneCenter() Vec3 { return m.geom.SceneCenter }

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
