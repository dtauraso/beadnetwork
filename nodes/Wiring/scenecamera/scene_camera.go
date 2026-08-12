package scenecamera

import (
	"math"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/camerapersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/spatial"
)

type vec3 = spatial.Vec3

func SeedInitialViewpoint(topologyPath string, setViewpoint func(pivot vec3, r float64, pos, up geom.Dir), emitViewpoint func(tr *T.Trace), tr *T.Trace) {
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

func LoadSceneViewpoint(topologyPath string) (pivot vec3, r float64, pos, up geom.Dir, ok bool) {
	var cp camerapersist.PolarCamera
	jsonpersist.ReadJSONBestEffort(scenepaths.CameraFilePath(topologyPath), &cp)

	if cp.Pivot == nil || cp.R == nil || cp.Pos == nil || cp.Up == nil {
		return vec3{}, 0, geom.Dir{}, geom.Dir{}, false
	}
	pivot = vec3{X: cp.Pivot[0], Y: cp.Pivot[1], Z: cp.Pivot[2]}
	r = *cp.R
	pos = geom.Dir{Theta: cp.Pos[0], Phi: cp.Pos[1]}
	up = geom.Dir{Theta: cp.Up[0], Phi: cp.Up[1]}
	return pivot, r, pos, up, true
}

const DefaultViewpointR = 500.0

func DefaultViewpoint() (pivot vec3, r float64, pos, up geom.Dir) {
	return vec3{X: 0, Y: 0, Z: 0},
		DefaultViewpointR,
		geom.Dir{Theta: math.Pi / 2, Phi: math.Pi / 2},
		geom.Dir{Theta: 0, Phi: 0}
}
