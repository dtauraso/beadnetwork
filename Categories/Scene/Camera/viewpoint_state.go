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
}

func (v *ViewpointState) EmitViewpoint() {

	if v.Persist != nil {
		v.Persist(v.Viewpoint)
	}
}

func (v *ViewpointState) OrbitViewpoint(base Viewpoint, from, to Dir, scale float64) {
	v.Orbit(base, from, to, scale)
	v.EmitViewpoint()
}

func (v *ViewpointState) OrbitLockedViewpoint(base Viewpoint, from, to Dir) {
	v.OrbitLocked(base, from, to)
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
