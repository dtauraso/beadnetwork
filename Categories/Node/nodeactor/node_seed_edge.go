package nodeactor

import (
	"github.com/dtauraso/wirefold/Categories/Node/nodeactor/nodefiles"
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
	"github.com/dtauraso/wirefold/Categories/Scene/loadspec"
)

func (m *NodeGeometry) SeedEdge(e loadspec.Edge, src bool, otherKind, sceneRoot string) {
	other := e.Target
	if !src {
		other = e.Source
	}

	m.RuleNode().SeedEdgeActive(other, nodefiles.LoadEdgeRuleActive(sceneRoot, e.Source, e.Target))
	m.AddNeighborKind(other, otherKind)
	if src {
		m.AddOutTarget(e.Target)
	}

	baseD, ok := e.BaseDeltaIndex()
	if !ok {
		return
	}
	dragD := e.DragDeltaIndex()
	if !src {
		baseD, dragD = polarindex.Neg(baseD), polarindex.Neg(dragD)
	}
	m.SetBaseDeltaTo(other, baseD)
	m.SetDragDeltaTo(other, dragD)
}
