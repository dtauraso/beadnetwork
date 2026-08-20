package NodeShape

import (
	"sync"

	"github.com/dtauraso/wirefold/src/Node/Wiring/framegeom"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/src/spatial"
)

const (
	RingSurfaceNu = nodegeom.ShadingParamNodeRingSurfaceNu
	RingSurfaceNv = nodegeom.ShadingParamNodeRingSurfaceNv
)

func CanonicalRingSurfacePoints() []spatial.Vec3 {
	return framegeom.CanonicalTorusSurfacePoints(nodegeom.ShadingParamNodeRingTubeRatio, RingSurfaceNu, RingSurfaceNv)
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
