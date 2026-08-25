package Node

import (
	"context"

	"github.com/dtauraso/beadnetwork/Categories/Node/TopVector"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

func (m *NodeGeometry) ArmTopVector(sceneRoot, targetID string) {
	off, ok := m.deltas.DeltaTo(targetID)
	if !ok {
		panic("Node.ArmTopVector: node " + m.id + " was asked to hold a top vector to " + targetID +
			", but it stores no delta to that node. The top vector IS the delta — without one there is nothing to compose its own index with, and it would draw an arrow from this node to itself")
	}
	m.topVector.SetDelta(off)

	values := TopVector.NewValueWriter(sceneRoot, int(m.stream.NodeRow()))
	m.topVector.SetRunner(TopVector.NewRunner(m.id, TopVector.Owner{
		SelfIndex: m.ComposedIndex,
		Constants: m.Constants,
		WorldPosAt: func(idx polarindex.Index) TopVector.Vec3 {
			return TopVector.Vec3(WorldPosAt(m.geom.SceneCenter, idx, m.geom.SceneConstants))
		},
	}, &m.topVector, values))
}

func (m *NodeGeometry) HasTopVector() bool { return m.topVector.Armed() }

func (m *NodeGeometry) RunTopVector(ctx context.Context) {
	m.topVector.Runner().Run(ctx)
}
