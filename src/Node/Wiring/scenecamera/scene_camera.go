package scenecamera

import (
	"math"

	"github.com/dtauraso/wirefold/src/Node/Wiring/camerapersist"
	"github.com/dtauraso/wirefold/src/Node/Wiring/geom/camera"
	"github.com/dtauraso/wirefold/src/Node/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/src/Node/spatial"
	T "github.com/dtauraso/wirefold/src/Trace"
)

type vec3 = spatial.Vec3

func SeedInitialViewpoint(topologyPath string, setViewpoint func(pivot vec3, r float64, pos, up camera.Dir), emitViewpoint func(tr *T.Trace), tr *T.Trace) {
	if setViewpoint == nil || emitViewpoint == nil {
		return
	}
	pivot, r, pos, up, ok := LoadSceneViewpoint(topologyPath)
	if !ok {
		pivot, r, pos, up = DefaultViewpoint()
	}
	setViewpoint(pivot, r, pos, up)
	emitViewpoint(tr)
}

func LoadSceneViewpoint(topologyPath string) (pivot vec3, r float64, pos, up camera.Dir, ok bool) {
	v, ok := camerapersist.ReadSceneCamera(scenepaths.CameraDirPath(topologyPath))
	if !ok {
		return vec3{}, 0, camera.Dir{}, camera.Dir{}, false
	}
	return v.Pivot, v.R, v.Pos, v.Up, true
}

const DefaultViewpointR = 500.0

func DefaultViewpoint() (pivot vec3, r float64, pos, up camera.Dir) {
	return vec3{X: 0, Y: 0, Z: 0},
		DefaultViewpointR,
		camera.Dir{Phi: math.Pi / 2, Theta: math.Pi / 2},
		camera.Dir{Phi: 0, Theta: 0}
}
