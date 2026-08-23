package loadspec

import (
	NodeBuf "github.com/dtauraso/wirefold/Categories/Node"
	"github.com/dtauraso/wirefold/Categories/Polar/polar"
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
)

func (s TopoSpec) SeedGeometry(sceneCenter Vec3) (
	map[string]NodeBuf.NodeGeom, map[string]polarindex.Index, map[string]polarindex.Offset,
) {
	geoms := make(map[string]NodeBuf.NodeGeom, len(s.Nodes))
	base := make(map[string]polarindex.Index, len(s.Nodes))
	drag := make(map[string]polarindex.Offset, len(s.Nodes))
	for _, n := range s.Nodes {
		geoms[n.ID], base[n.ID], drag[n.ID] = n.SeedGeometry(sceneCenter, s.Constants)
	}
	return geoms, base, drag
}

func (n Node) SeedGeometry(sceneCenter Vec3, sc polarindex.SceneConstants) (NodeBuf.NodeGeom, polarindex.Index, polarindex.Offset) {
	g := n.ToNodeGeom(sceneCenter, sc)

	base := n.declaredIndex(sc)
	if base == nil && g.HasPos {
		p := polar.Cart2polar(polar.Vec3(NodeBuf.NodeWorldPos(g).Sub(NodeBuf.Vec3(sceneCenter))))
		m := polarindex.MeasureIndex(p, sc)
		base = &m
	}
	if base == nil {
		base = &polarindex.Index{}
	}

	NodeBuf.SetNodeWorld(&g, *base)
	return g, *base, n.dragIndex()
}

func (n Node) declaredIndex(sc polarindex.SceneConstants) *polarindex.Index {
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

func (n Node) dragIndex() polarindex.Offset {
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
