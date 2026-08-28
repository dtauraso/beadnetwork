package Camera

const ViewpointMinDist = 5.0

type Viewpoint struct {
	Pivot      Vec3
	R          float64
	Pos        Dir
	Up         Dir
}

func (v *Viewpoint) Rotate(rt Rot) {
	v.Pos = RotateDir(v.Pos, rt.Axis, rt.Angle)
	v.Up = RotateDir(v.Up, rt.Axis, rt.Angle)
}

func (v *Viewpoint) Orbit(base Viewpoint, from, to Dir, scale float64) {
	rt := ArcBetween(from, to)
	rt.Angle *= scale
	*v = base
	v.Rotate(rt)
}

func (v *Viewpoint) OrbitLocked(base Viewpoint, from, to Dir, centre Vec3) {
	*v = base

	eye := EyeOf(base)
	spoke := eye.Sub(centre)
	if spoke.Length() < 1e-9 {
		v.Rotate(Rot{Axis: base.Pos, Angle: AngleAboutAxis(from, to, base.Pos)})
		return
	}
	axis := vecToDir(spoke)
	angle := AngleAboutAxis(from, to, axis)

	k := dirToVec(axis)
	pivot := centre.Add(rotateVecAbout(base.Pivot.Sub(centre), k, angle))
	spokeToPivot := eye.Sub(pivot)

	v.Pivot = pivot
	v.R = spokeToPivot.Length()
	v.Pos = WorldDirToAngles(spokeToPivot)
	v.Up = RotateDir(base.Up, axis, angle)
}

func (v *Viewpoint) Zoom(factor float64) {
	nr := v.R * factor
	if nr < ViewpointMinDist {
		nr = ViewpointMinDist
	}
	v.R = nr
}

func (v *Viewpoint) Pan(delta Vec3) {
	v.Pivot = v.Pivot.Add(delta)
}
