package geom

// camera_screen.go — screen ↔ sphere conversions (mirrors polar.ts screenToPolar / toWorld
// / planeSlide) and the polar-frame pan displacement, split from camera_angles.go by
// concern (see that file's header for the shared quarantine rationale).

import "math"

// ---------------------------------------------------------------------------
// screen ↔ sphere (mirrors polar.ts screenToPolar / toWorld)
// ---------------------------------------------------------------------------

// PolarDir is a direction on the sphere in a CamBasis frame (polar.ts Polar).
type PolarDir struct {
	Theta float64
	Phi   float64
}

// ScreenToPolar mirrors polar.ts screenToPolar:
//
//	phi = hypot(dx,dy)/scale; theta = atan2(-dy, dx)
func ScreenToPolar(dxFromCenter, dyFromCenter, scale float64) PolarDir {
	return PolarDir{
		Phi:   math.Hypot(dxFromCenter, dyFromCenter) / scale,
		Theta: math.Atan2(-dyFromCenter, dxFromCenter),
	}
}

// ToWorldDir mirrors polar.ts toWorld with center=C and radius=1, then .Sub(C): it returns
// the UNIT world direction (pole*cos(phi) + equatorDir*sin(phi)) — the .Sub(C).Normalize()
// in the TS orbit path just recovers this direction, since radius is 1.
//
//	equatorDir = refX*cos(theta) + refY*sin(theta)
//	dir        = pole*cos(phi) + equatorDir*sin(phi)
func ToWorldDir(b CamBasis, q PolarDir) vec3 {
	s := math.Sin(q.Phi)
	equator := b.RefX.Scale(math.Cos(q.Theta)).Add(b.RefY.Scale(math.Sin(q.Theta)))
	return b.Pole.Scale(math.Cos(q.Phi)).Add(equator.Scale(s))
}

// PlaneSlide mirrors polar.ts planeSlide: a polar in-screen-plane slide (r, angle) → a world
// translation along the camera's right/up basis, scaled by worldPerPixel.
func PlaneSlide(b CamBasis, r, angle, worldPerPixel float64) vec3 {
	return b.RefX.Scale(r * math.Cos(angle) * worldPerPixel).
		Add(b.RefY.Scale(r * math.Sin(angle) * worldPerPixel))
}

// DeltaToPolar mirrors polar.ts deltaToPolar: (dx,dy) → (r, angle).
func DeltaToPolar(dx, dy float64) (r, angle float64) {
	return math.Hypot(dx, dy), math.Atan2(dy, dx)
}

// PanDisplacementPolar builds the lateral pan displacement in the POLAR frame
// (polar-frame-rewrite.md): the mouse drag gives r (distance) and a screen bearing; the
// displacement DIRECTION is a direction 90° off the camera view axis (i.e. in the screen
// plane) at that bearing, derived from the camera's own (θ,φ) and up via the spherical
// toolkit — no cartesian basis vectors. r is the magnitude. The single cartesian is the
// Polar2cart at the end, composing the finished displacement for the scene-center move (the
// pointer input boundary). This is locked to the known-correct PlaneSlide by a unit test.
//
//	psiUp    = bearing of the up-hint about the view axis (AzimuthFrom)
//	psiRight = psiUp − π/2 (screen right is a quarter-turn before up, right-handed about pos)
//	dir      = FromAxisFrame(pos, π/2, psiRight + bearing)   // on the view-axis equator
func PanDisplacementPolar(pos, up Dir, dx, dy, worldPerPixel float64) vec3 {
	r, bearing := DeltaToPolar(dx, -dy)
	_, psiUp := AzimuthFrom(pos, up)
	d := FromAxisFrame(pos, math.Pi/2, psiUp-math.Pi/2+bearing)
	return Polar2cart(Polar{R: r * worldPerPixel, Theta: d.Theta, Phi: d.Phi})
}
