package geom

// viewpoint.go — the polar camera state and its navigation ops, expressed in angle-only
// spherical terms (spherical.go). The renderer (three.js) turns this into a Cartesian
// camera at draw time; nothing here forms a rotation vector or quaternion. The only
// Cartesian is the `Pivot`, which is a plain anchor POINT (like a node center), never
// rotation math — it is just added to the camera position at the renderer edge.
//
// Orientation is carried as two directions: `Pos` (pivot → camera, which also fixes the
// look direction since the camera looks at the pivot) and `Up` (the screen-up hint, which
// carries ROLL). A rotation spins BOTH rigidly, so tilt accumulates and is preserved.

const ViewpointMinDist = 5.0

type Viewpoint struct {
	Pivot      vec3    // world orbit center (anchor point; not rotation math)
	R          float64 // distance from pivot to camera
	Pos        Dir     // direction pivot → camera
	Up         Dir     // up-hint direction (carries roll)
	LockedAxis *Dir    // locked rotation axis for handhold-constrained orbit; nil between gestures
}

// Rotate spins the whole camera frame (position direction AND up) by the same rotation,
// so roll is preserved as a rigid turn of the frame rather than recomputed.
func (v *Viewpoint) Rotate(rt Rot) {
	v.Pos = RotateDir(v.Pos, rt.Axis, rt.Angle)
	v.Up = RotateDir(v.Up, rt.Axis, rt.Angle)
}

// Orbit applies the shortest-arc rotation carrying `from` to `to`, so a grabbed direction
// follows the cursor (the motion-driven great-circle gesture).
func (v *Viewpoint) Orbit(from, to Dir) {
	v.Rotate(ArcBetween(from, to))
}

// OrbitLocked performs a handhold-constrained rotation: the first call locks the rotation
// axis from the from→to arc; subsequent calls keep the same axis and track only the angle.
// The lock is cleared by SetViewpoint (gesture end). Mirrors the TS handhold path.
func (v *Viewpoint) OrbitLocked(from, to Dir) {
	if v.LockedAxis == nil {
		ax := ArcBetween(from, to).Axis
		v.LockedAxis = &ax
	}
	angle := AngleAboutAxis(from, to, *v.LockedAxis)
	v.Rotate(Rot{Axis: *v.LockedAxis, Angle: angle})
}

// Zoom scales the orbit radius about the pivot, floored so the camera never reaches it.
func (v *Viewpoint) Zoom(factor float64) {
	nr := v.R * factor
	if nr < ViewpointMinDist {
		nr = ViewpointMinDist
	}
	v.R = nr
}

// Pan slides the orbit pivot by a world delta; the camera rides along (position is pivot
// plus the radial offset). The delta is a world vector computed at the renderer edge.
func (v *Viewpoint) Pan(delta vec3) {
	v.Pivot = v.Pivot.Add(delta)
}
