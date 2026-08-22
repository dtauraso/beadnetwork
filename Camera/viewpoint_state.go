package Camera

import ()

type ViewpointState struct {
	Viewpoint

	Persist func(Viewpoint)
}

func (v *ViewpointState) SetViewpoint(pivot Vec3, r float64, pos, up Dir) {
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

func (v *ViewpointState) OrbitViewpoint(from, to Dir) {
	v.Orbit(from, to)
	v.EmitViewpoint()
}

func (v *ViewpointState) OrbitLockedViewpoint(from, to Dir) {
	v.OrbitLocked(from, to)
	v.EmitViewpoint()
}

func (v *ViewpointState) ZoomViewpoint(factor float64) {
	v.Zoom(factor)
	v.EmitViewpoint()
}

func (v *ViewpointState) PanViewpoint(delta Vec3) {
	v.Pan(delta)
	v.EmitViewpoint()
}
