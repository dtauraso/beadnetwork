package scenebuild

import (
	NodeBuf "github.com/dtauraso/wirefold/Categories/Node"
	"github.com/dtauraso/wirefold/Categories/Polar/polar"
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
	"github.com/dtauraso/wirefold/Categories/Scene/loadspec"
)

func SeedGeometry(s loadspec.TopoSpec, sceneCenter loadspec.Vec3) (
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

func SeedNodeGeometry(n loadspec.Node, sceneCenter loadspec.Vec3, sc polarindex.SceneConstants) (NodeBuf.NodeGeom, polarindex.Index, polarindex.Offset) {
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

func toNodeGeom(n loadspec.Node, sceneCenter loadspec.Vec3, sc polarindex.SceneConstants) NodeBuf.NodeGeom {
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

func declaredIndex(n loadspec.Node, sc polarindex.SceneConstants) *polarindex.Index {
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

func dragIndex(n loadspec.Node) polarindex.Offset {
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
