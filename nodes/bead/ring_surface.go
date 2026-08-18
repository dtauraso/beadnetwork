package bead

import (
	"sync"

	"github.com/dtauraso/wirefold/nodes/Wiring/framegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/spatial"
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
