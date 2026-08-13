package scenecamera

import (
	"math"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/camerapersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/camera"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/spatial"
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
	var cp camerapersist.PolarCamera
	jsonpersist.ReadJSONBestEffort(scenepaths.CameraFilePath(topologyPath), &cp)

	if cp.Pivot == nil || cp.R == nil || cp.Pos == nil || cp.Up == nil {
		return vec3{}, 0, camera.Dir{}, camera.Dir{}, false
	}
	pivot = vec3{X: cp.Pivot[0], Y: cp.Pivot[1], Z: cp.Pivot[2]}
	r = *cp.R
	pos = camera.Dir{Phi: cp.Pos[0], Theta: cp.Pos[1]}
	up = camera.Dir{Phi: cp.Up[0], Theta: cp.Up[1]}
	return pivot, r, pos, up, true
}

const DefaultViewpointR = 500.0

func DefaultViewpoint() (pivot vec3, r float64, pos, up camera.Dir) {
	return vec3{X: 0, Y: 0, Z: 0},
		DefaultViewpointR,
		camera.Dir{Phi: math.Pi / 2, Theta: math.Pi / 2},
		camera.Dir{Phi: 0, Theta: 0}
}
