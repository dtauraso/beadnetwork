package gesturefsm

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/camera"
	"github.com/dtauraso/wirefold/nodes/spatial"
	T "github.com/dtauraso/wirefold/src/Trace"
)

type ViewpointState struct {
	camera.Viewpoint

	Persist func(camera.Viewpoint)
}

func (v *ViewpointState) SetViewpoint(pivot spatial.Vec3, r float64, pos, up camera.Dir) {
	v.Pivot = pivot
	v.R = r
	v.Pos = pos
	v.Up = up
	v.LockedAxis = nil
}

func (v *ViewpointState) EmitViewpoint(tr *T.Trace) {

	if v.Persist != nil {
		v.Persist(v.Viewpoint)
	}
}

func (v *ViewpointState) OrbitViewpoint(from, to camera.Dir, tr *T.Trace) {
	v.Orbit(from, to)
	v.EmitViewpoint(tr)
}

func (v *ViewpointState) OrbitLockedViewpoint(from, to camera.Dir, tr *T.Trace) {
	v.OrbitLocked(from, to)
	v.EmitViewpoint(tr)
}

func (v *ViewpointState) ZoomViewpoint(factor float64, tr *T.Trace) {
	v.Zoom(factor)
	v.EmitViewpoint(tr)
}

func (v *ViewpointState) PanViewpoint(delta spatial.Vec3, tr *T.Trace) {
	v.Pan(delta)
	v.EmitViewpoint(tr)
}
