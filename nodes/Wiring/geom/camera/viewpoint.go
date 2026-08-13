package camera

const ViewpointMinDist = 5.0

type Viewpoint struct {
	Pivot vec3
	R     float64
	Pos   Dir
	Up    Dir
}

func (v *Viewpoint) Rotate(rt Rot) {
	v.Pos = RotateDir(v.Pos, rt.Axis, rt.Angle)
	v.Up = RotateDir(v.Up, rt.Axis, rt.Angle)
}

func (v *Viewpoint) Orbit(from, to Dir) {
	v.Rotate(ArcBetween(from, to))
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
