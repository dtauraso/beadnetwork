package gesturefsm

import (
	"github.com/dtauraso/wirefold/src/Node/Wiring/geom/camera"
	"github.com/dtauraso/wirefold/src/Node/spatial"
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

func (v *ViewpointState) EmitViewpoint() {

	if v.Persist != nil {
		v.Persist(v.Viewpoint)
	}
}

func (v *ViewpointState) OrbitViewpoint(from, to camera.Dir) {
	v.Orbit(from, to)
	v.EmitViewpoint()
}

func (v *ViewpointState) OrbitLockedViewpoint(from, to camera.Dir) {
	v.OrbitLocked(from, to)
	v.EmitViewpoint()
}

func (v *ViewpointState) ZoomViewpoint(factor float64) {
	v.Zoom(factor)
	v.EmitViewpoint()
}

func (v *ViewpointState) PanViewpoint(delta spatial.Vec3) {
	v.Pan(delta)
	v.EmitViewpoint()
}
