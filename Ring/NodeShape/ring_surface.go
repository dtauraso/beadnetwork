package NodeShape

import (
	"sync"

	"github.com/dtauraso/wirefold/Node/framegeom"
	"github.com/dtauraso/wirefold/Node/nodegeom"
)

const (
	RingSurfaceNu = nodegeom.ShadingParamNodeRingSurfaceNu
	RingSurfaceNv = nodegeom.ShadingParamNodeRingSurfaceNv
)

func CanonicalRingSurfacePoints() []Vec3 {
	pts := framegeom.CanonicalTorusSurfacePoints(nodegeom.ShadingParamNodeRingTubeRatio, RingSurfaceNu, RingSurfaceNv)
	out := make([]Vec3, len(pts))
	for i, p := range pts {
		out[i] = Vec3(p)
	}
	return out
}

func toFrameGeomPoints(pts []Vec3) []framegeom.Vec3 {
	out := make([]framegeom.Vec3, len(pts))
	for i, p := range pts {
		out[i] = framegeom.Vec3(p)
	}
	return out
}

var (
	ringSurfaceFlatOnce sync.Once
	ringSurfaceFlat     []float32
)

func CanonicalRingSurfacePointsFlat() []float32 {
	ringSurfaceFlatOnce.Do(func() {
		ringSurfaceFlat = framegeom.FlattenPoints(toFrameGeomPoints(CanonicalRingSurfacePoints()))
	})
	return ringSurfaceFlat
}
