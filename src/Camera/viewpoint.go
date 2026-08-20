package Camera

const ViewpointMinDist = 5.0

type Viewpoint struct {
	Pivot      vec3
	R          float64
	Pos        Dir
	Up         Dir
	LockedAxis *Dir
}

func (v *Viewpoint) Rotate(rt Rot) {
	v.Pos = RotateDir(v.Pos, rt.Axis, rt.Angle)
	v.Up = RotateDir(v.Up, rt.Axis, rt.Angle)
}

func (v *Viewpoint) Orbit(from, to Dir) {
	v.Rotate(ArcBetween(from, to))
}

func (v *Viewpoint) OrbitLocked(from, to Dir) {
	if v.LockedAxis == nil {
		ax := ArcBetween(from, to).Axis
		v.LockedAxis = &ax
	}
	angle := AngleAboutAxis(from, to, *v.LockedAxis)
	v.Rotate(Rot{Axis: *v.LockedAxis, Angle: angle})
}

func (v *Viewpoint) Zoom(factor float64) {
	nr := v.R * factor
	if nr < ViewpointMinDist {
		nr = ViewpointMinDist
	}
	v.R = nr
}

func (v *Viewpoint) Pan(delta vec3) {
	v.Pivot = v.Pivot.Add(delta)
}
