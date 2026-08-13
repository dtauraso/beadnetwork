package camera

const ViewpointMinDist = 5.0

type Viewpoint struct {
	Pivot vec3
	R     float64
	Pos   Dir
	Up    Dir

	// LockedAxis is what makes a handhold drag a ROLL rather than an orbit: it
	// is pinned on the gesture's first move and held until the gesture ends, so
	// every later move turns about that same axis. Nil between gestures.
	LockedAxis *Dir
}

func (v *Viewpoint) Rotate(rt Rot) {
	v.Pos = RotateDir(v.Pos, rt.Axis, rt.Angle)
	v.Up = RotateDir(v.Up, rt.Axis, rt.Angle)
}

func (v *Viewpoint) Orbit(from, to Dir) {
	v.Rotate(ArcBetween(from, to))
}

// OrbitLocked rolls about one axis for the whole gesture. The first move names
// the axis; every move after it contributes only its turn about that axis.
func (v *Viewpoint) OrbitLocked(from, to Dir) {
	if v.LockedAxis == nil {
		ax := ArcBetween(from, to).Axis
		v.LockedAxis = &ax
	}
	v.Rotate(Rot{Axis: *v.LockedAxis, Angle: AngleAboutAxis(from, to, *v.LockedAxis)})
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
