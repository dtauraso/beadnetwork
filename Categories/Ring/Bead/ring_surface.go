package bead

import (
	"sync"

	Ring "github.com/dtauraso/beadnetwork/Categories/Ring"
)

const (
	RingSurfaceNu = ShadingParamBeadRingSurfaceNu
	RingSurfaceNv = ShadingParamBeadRingSurfaceNv
)

func CanonicalRingSurfacePoints() []Vec3 {
	pts := Ring.CanonicalTorusSurfacePoints(ShadingParamBeadRingTubeRatio, RingSurfaceNu, RingSurfaceNv)
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
