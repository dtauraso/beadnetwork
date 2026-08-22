package Camera

import (
	"math"

	"github.com/dtauraso/wirefold/src/spatial"
)

func SeedInitialViewpoint(topologyPath string, setViewpoint func(pivot spatial.Vec3, r float64, pos, up Dir), emitViewpoint func()) {
	if setViewpoint == nil || emitViewpoint == nil {
		return
	}
	pivot, r, pos, up, ok := LoadSceneViewpoint(topologyPath)
	if !ok {
		pivot, r, pos, up = DefaultViewpoint()
	}
	setViewpoint(pivot, r, pos, up)
	emitViewpoint()
}

func LoadSceneViewpoint(topologyPath string) (pivot spatial.Vec3, r float64, pos, up Dir, ok bool) {
	v, ok := ReadSceneCamera(BlockPath(topologyPath))
	if !ok {
		return spatial.Vec3{}, 0, Dir{}, Dir{}, false
	}
	return v.Pivot, v.R, v.Pos, v.Up, true
}

const DefaultViewpointR = 500.0

func DefaultViewpoint() (pivot spatial.Vec3, r float64, pos, up Dir) {
	return spatial.Vec3{X: 0, Y: 0, Z: 0},
		DefaultViewpointR,
		Dir{Phi: math.Pi / 2, Theta: math.Pi / 2},
		Dir{Phi: 0, Theta: 0}
}
