package scenebuild

import (
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/PolarRulesPanel"

	NodeBuf "github.com/dtauraso/wirefold/Categories/Node"
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
	"github.com/dtauraso/wirefold/Categories/Scene/loadspec"
)

func SeedNode(m *NodeBuf.NodeGeometry, n loadspec.Node, sceneRoot string) {
	m.SetSelfKind(n.Type)

	rn := m.RuleNode()
	rn.SetPersist(NodeBuf.Persist{
		Rule:       func(r *PolarRulesPanel.DragRule) error { return NodeBuf.WriteDragRule(sceneRoot, n.ID, r) },
		SelfRule:   func(r *PolarRulesPanel.DragRule) error { return NodeBuf.WriteSelfDragRule(sceneRoot, n.ID, r) },
		SelfActive: func(a bool) error { return NodeBuf.WriteSelfRuleActive(sceneRoot, n.ID, a) },
		DragActive: func(a bool) error { return NodeBuf.WriteDragActive(sceneRoot, n.ID, a) },
		KindActive: func(a bool) error { return NodeBuf.WriteKindRuleActive(sceneRoot, n.ID, a) },
		EdgeActive: func(target string, a bool) error {
			return NodeBuf.WriteEdgeRuleActive(sceneRoot, n.ID, target, a)
		},
	})

	active := NodeBuf.LoadDragActive(sceneRoot, n.ID)
	m.Topo().SetDragRule(n.Drag)
	m.Topo().SetDragActive(active)
	rn.SeedRule(n.Drag, active)
	rn.SeedKindActive(NodeBuf.LoadKindRuleActive(sceneRoot, n.ID))

	selfActive := NodeBuf.LoadSelfRuleActive(sceneRoot, n.ID)
	m.Topo().SetSelfRule(n.SelfDrag)
	m.Topo().SetSelfRuleActive(selfActive)
	rn.SeedSelfRule(n.SelfDrag, selfActive)

	if n.TopTiltVectorPhiIdx != nil {
		m.Tilt().SetTopTiltVectorPhiIdx(*n.TopTiltVectorPhiIdx)
	}
}

func SeedEdge(m *NodeBuf.NodeGeometry, e loadspec.Edge, src bool, otherKind, sceneRoot string) {
	other := e.Target
	if !src {
		other = e.Source
	}

	m.RuleNode().SeedEdgeActive(other, NodeBuf.LoadEdgeRuleActive(sceneRoot, e.Source, e.Target))
	m.Topo().AddNeighborKind(other, otherKind)
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
	m.Deltas().SetBaseDeltaTo(other, baseD)
	m.Deltas().SetDragDeltaTo(other, dragD)
}
