package gesturefsm

import (
	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/spatial"
)

type ViewpointState struct {
	geom.Viewpoint

	Persist func(geom.Viewpoint)
}

func (v *ViewpointState) SetViewpoint(pivot spatial.Vec3, r float64, pos, up geom.Dir) {
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

func (v *ViewpointState) OrbitViewpoint(from, to geom.Dir, tr *T.Trace) {
	v.Orbit(from, to)
	v.EmitViewpoint(tr)
}

func (v *ViewpointState) OrbitLockedViewpoint(from, to geom.Dir, tr *T.Trace) {
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
