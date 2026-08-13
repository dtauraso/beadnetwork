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

// OrbitLocked ROLLS: the scene spins in place, like turning a steering wheel,
// and the viewpoint does not travel.
//
// The axis is the VIEW DIRECTION. Rotating Pos about Pos leaves Pos exactly
// where it is and turns only Up, which is what a roll is.
//
// It used to lock onto ArcBetween(from, to).Axis — the axis perpendicular to
// the gesture's first drag step. Rotating about a perpendicular axis MOVES the
// viewpoint along a great circle, carrying Up around with it, so the scene
// swung sideways and twisted instead of spinning: an orbit constrained to one
// axis, not a roll.
//
// The axis is still pinned for the gesture. It is invariant under its own
// rotation, so this is belt and braces rather than load-bearing.
func (v *Viewpoint) OrbitLocked(from, to Dir) {
	if v.LockedAxis == nil {
		ax := v.Pos
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
