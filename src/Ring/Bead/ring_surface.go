package bead

import (
	"sync"

	"github.com/dtauraso/wirefold/src/Node/framegeom"
	"github.com/dtauraso/wirefold/src/Node/nodegeom"
	"github.com/dtauraso/wirefold/src/spatial"
)

const (
	RingSurfaceNu = nodegeom.ShadingParamBeadRingSurfaceNu
	RingSurfaceNv = nodegeom.ShadingParamBeadRingSurfaceNv
)

func CanonicalRingSurfacePoints() []spatial.Vec3 {
	return framegeom.CanonicalTorusSurfacePoints(nodegeom.ShadingParamBeadRingTubeRatio, RingSurfaceNu, RingSurfaceNv)
}

var (
	ringSurfaceFlatOnce sync.Once
	ringSurfaceFlat     []float32
)

func CanonicalRingSurfacePointsFlat() []float32 {
	ringSurfaceFlatOnce.Do(func() {
		ringSurfaceFlat = framegeom.FlattenPoints(CanonicalRingSurfacePoints())
	})
	return ringSurfaceFlat
}
