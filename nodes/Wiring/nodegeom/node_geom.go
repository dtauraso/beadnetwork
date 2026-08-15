package nodegeom

import (
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"
)

type NodeIdentity struct {
	Kind  string
	Label string

	SceneCenter vec3
}

type NodeGeom struct {
	NodeIdentity

	BasePolar polar.Polar
	DragPolar polar.Polar
	HasPos    bool
}

func ScenePolarOf(g NodeGeom) polar.Polar {
	return polar.Compose(g.BasePolar, g.DragPolar)
}

func KindWidthHeight(kind string) (float64, float64) {
	if d, ok := KindDims[kind]; ok {
		return d.Width, d.Height
	}
	return 110, 60
}

func BareNodeRadius(kind string) float64 {
	w, h := KindWidthHeight(kind)
	return min(w, h) / float64(CurveParamNodeRadiusDivisor)
}

func NodeRadius(kind string) float64 {
	return NodeTorusOuterR(kind) / (1 + ShadingParamNodeRingTubeRatio)
}

func NodeWorldPos(g NodeGeom) vec3 {
	if !g.HasPos {
		return vec3{}
	}
	return g.SceneCenter.Add(polar.Polar2cart(ScenePolarOf(g)))
}

func SetNodeWorld(g *NodeGeom, world vec3) {
	scene := polar.Cart2polar(world.Sub(g.SceneCenter))
	g.DragPolar = polar.Between(g.BasePolar, scene)
	g.HasPos = true
}

func NodeTorusSteps(kind string) int {
	unsnapped := BareNodeRadius(kind) * (1 + ShadingParamNodeRingTubeRatio)
	return int(math.Round(unsnapped / lattice.BeadStepR))
}

func NodeTorusOuterR(kind string) float64 {
	return float64(NodeTorusSteps(kind)) * lattice.BeadStepR
}
