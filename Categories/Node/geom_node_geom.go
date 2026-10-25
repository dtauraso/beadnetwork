package Node

import (
	"math"

	lattice "github.com/dtauraso/beadnetwork/Categories/Vectors/lattice"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

type NodeIdentity struct {
	Kind  string
	Label string

	SceneCenter    Vec3
	SceneConstants polarindex.SceneConstants
}

type NodeGeom struct {
	NodeIdentity

	BaseIndex polarindex.Index
	DragIndex polarindex.Offset
	Rest      polarindex.Place
	HasPos    bool
}

func ComposedIndexOf(g NodeGeom) polarindex.Index {
	return polarindex.Compose(g.BaseIndex, g.DragIndex, g.SceneConstants)
}

func ScenePolarOf(g NodeGeom) polar.Polar {
	return polarindex.ToPolarAt(ComposedIndexOf(g), g.Rest, g.SceneConstants)
}

func WorldPosOf(g NodeGeom, idx polarindex.Index) Vec3 {
	return WorldPosAtRest(g.SceneCenter, idx, g.Rest, g.SceneConstants)
}

func WorldPosAtRest(sceneCenter Vec3, idx polarindex.Index, rest polarindex.Place, sc polarindex.SceneConstants) Vec3 {
	return sceneCenter.Add(Vec3(polar.Polar2cart(polarindex.ToPolarAt(idx, rest, sc))))
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

func NodeWorldPos(g NodeGeom) Vec3 {
	if !g.HasPos {
		return Vec3{}
	}
	return g.SceneCenter.Add(Vec3(polar.Polar2cart(ScenePolarOf(g))))
}

func IndexAtTheta(sceneCenter, world Vec3, theta float64, sc polarindex.SceneConstants) polarindex.Index {
	return polarindex.MeasureIndex(polar.Cart2polarAtTheta(polar.Vec3(world.Sub(sceneCenter)), theta), sc)
}

func WorldPosAt(sceneCenter Vec3, idx polarindex.Index, sc polarindex.SceneConstants) Vec3 {
	return sceneCenter.Add(Vec3(polar.Polar2cart(polarindex.ToPolar(idx, sc))))
}

func SetNodeWorld(g *NodeGeom, composed polarindex.Index) {
	g.DragIndex = polarindex.Delta(composed, g.BaseIndex)
	g.HasPos = true
}

func NodeTorusSteps(kind string) int {
	unsnapped := BareNodeRadius(kind) * (1 + ShadingParamNodeRingTubeRatio)
	return int(math.Round(unsnapped / lattice.SlotR))
}

func NodeTorusOuterR(kind string) float64 {
	return float64(NodeTorusSteps(kind)) * lattice.SlotR
}

const PoleRingSteps = 6

func PoleRingR() float64 {
	return PoleRingSteps * lattice.BeadStepR
}
