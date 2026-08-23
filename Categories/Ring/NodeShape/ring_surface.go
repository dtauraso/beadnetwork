package NodeShape

import (
	"sync"

	"github.com/dtauraso/wirefold/Categories/Node"
	Ring "github.com/dtauraso/wirefold/Categories/Ring"
)

const (
	RingSurfaceNu = ShadingParamNodeRingSurfaceNu
	RingSurfaceNv = ShadingParamNodeRingSurfaceNv
)

func CanonicalRingSurfacePoints() []Vec3 {
	pts := Ring.CanonicalTorusSurfacePoints(Node.ShadingParamNodeRingTubeRatio, RingSurfaceNu, RingSurfaceNv)
	out := make([]Vec3, len(pts))
	for i, p := range pts {
		out[i] = Vec3(p)
	}
	return out
}

func toFrameGeomPoints(pts []Vec3) []Ring.Vec3 {
	out := make([]Ring.Vec3, len(pts))
	for i, p := range pts {
		out[i] = Ring.Vec3(p)
	}
	return out
}

var (
	ringSurfaceFlatOnce sync.Once
	ringSurfaceFlat     []float32
)

func CanonicalRingSurfacePointsFlat() []float32 {
	ringSurfaceFlatOnce.Do(func() {
		ringSurfaceFlat = Ring.FlattenPoints(toFrameGeomPoints(CanonicalRingSurfacePoints()))
	})
	return ringSurfaceFlat
}
