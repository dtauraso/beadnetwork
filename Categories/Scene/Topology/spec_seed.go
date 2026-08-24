package Topology

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/PolarRulesPanel"

	NodeBuf "github.com/dtauraso/beadnetwork/Categories/Node"
	"github.com/dtauraso/beadnetwork/Categories/Vector/polar"
	"github.com/dtauraso/beadnetwork/Categories/Vector/polarindex"
)

func SeedNode(m *NodeBuf.NodeGeometry, n Node, sceneRoot string) {
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

func SeedEdge(m *NodeBuf.NodeGeometry, e Edge, src bool, otherKind, sceneRoot string) {
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

func SeedGeometry(s TopoSpec, sceneCenter Vec3) (
	map[string]NodeBuf.NodeGeom, map[string]polarindex.Index, map[string]polarindex.Offset,
) {
	geoms := make(map[string]NodeBuf.NodeGeom, len(s.Nodes))
	base := make(map[string]polarindex.Index, len(s.Nodes))
	drag := make(map[string]polarindex.Offset, len(s.Nodes))
	for _, n := range s.Nodes {
		geoms[n.ID], base[n.ID], drag[n.ID] = SeedNodeGeometry(n, sceneCenter, s.Constants)
	}
	return geoms, base, drag
}

func SeedNodeGeometry(n Node, sceneCenter Vec3, sc polarindex.SceneConstants) (NodeBuf.NodeGeom, polarindex.Index, polarindex.Offset) {
	g := toNodeGeom(n, sceneCenter, sc)

	base := declaredIndex(n, sc)
	if base == nil && g.HasPos {
		p := polar.Cart2polar(polar.Vec3(NodeBuf.NodeWorldPos(g).Sub(NodeBuf.Vec3(sceneCenter))))
		m := polarindex.MeasureIndex(p, sc)
		base = &m
	}
	if base == nil {
		base = &polarindex.Index{}
	}

	NodeBuf.SetNodeWorld(&g, *base)
	return g, *base, dragIndex(n)
}

func toNodeGeom(n Node, sceneCenter Vec3, sc polarindex.SceneConstants) NodeBuf.NodeGeom {
	g := NodeBuf.NodeGeom{NodeIdentity: NodeBuf.NodeIdentity{Kind: n.Type, Label: n.Label(), SceneCenter: NodeBuf.Vec3(sceneCenter), SceneConstants: sc}}
	if n.HasPoint() {
		g.BaseIndex = n.Index()
		g.HasPos = true
	}
	if n.DragIndexPhi != nil && n.DragIndexTheta != nil && n.DragIndexR != nil {
		g.DragIndex = polarindex.Offset{Phi: *n.DragIndexPhi, Theta: *n.DragIndexTheta, R: *n.DragIndexR}
	}
	return g
}

func declaredIndex(n Node, sc polarindex.SceneConstants) *polarindex.Index {
	if n.IndexPhi == nil || n.IndexTheta == nil || n.IndexR == nil {
		return nil
	}
	i := polarindex.Canonical(polarindex.Index{
		Phi:   *n.IndexPhi,
		Theta: *n.IndexTheta,
		R:     *n.IndexR,
	}, sc)
	return &i
}

func dragIndex(n Node) polarindex.Offset {
	var o polarindex.Offset
	if n.DragIndexPhi != nil {
		o.Phi = *n.DragIndexPhi
	}
	if n.DragIndexTheta != nil {
		o.Theta = *n.DragIndexTheta
	}
	if n.DragIndexR != nil {
		o.R = *n.DragIndexR
	}
	return o
}
