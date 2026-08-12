package geom

// camera_angles.go — the RENDERER-EDGE Cartesian camera math, ported FORMULA-FOR-FORMULA
// from the TS source so Go's gesture state machine (gesture.go) can turn RAW pointer/wheel
// pixels into the polar viewpoint ops (orbit / zoom / pan) that already live in
// spherical.go + viewpoint.go. This is the SAME quarantined Cartesian the TS side isolates
// in polar.ts (cameraFrame / toWorld / planeSlide) and viewpoint-bridge.ts
// (anglesToWorldOffset / worldDirToAngles); porting it here — instead of reinventing it —
// keeps the hard-won great-circle orbit form and avoids reintroducing the arcball /
// wrong-axis bug class. Every function names the TS source it mirrors.
//
// The great-circle ORBIT itself is NOT re-derived here: this file (split into
// camera_angles.go / camera_basis.go / camera_screen.go / camera_focus.go /
// camera_homefit.go / camera_project.go, same geom package, one concern each) only
// produces the two world directions (prevDir, currDir) the orbit carries; Viewpoint.Orbit
// (spherical.go ArcBetween/RotateDir) does the rotation. Radius / pan translation are
// likewise handed to Viewpoint.Zoom / Viewpoint.Pan.

import "math"

// ---------------------------------------------------------------------------
// vec3 dot / cross moved to wire.Vec3.Dot/Cross (nodes/wire/geometry.go) — vec3
// here is an alias of wire.Vec3, so geom calls the exported methods directly.
// ---------------------------------------------------------------------------
// angle ↔ world direction (mirrors viewpoint-bridge.ts)
// ---------------------------------------------------------------------------

// AnglesToWorldOffset mirrors viewpoint-bridge.ts anglesToWorldOffset:
//
//	x = r*sin(theta)*cos(phi), y = r*cos(theta), z = r*sin(theta)*sin(phi)
func AnglesToWorldOffset(r, theta, phi float64) vec3 {
	sinT := math.Sin(theta)
	return vec3{
		X: r * sinT * math.Cos(phi),
		Y: r * math.Cos(theta),
		Z: r * sinT * math.Sin(phi),
	}
}

// WorldDirToAngles mirrors polar.ts worldDirToFrameAngles with Y_POLE_FRAME (pole=+y,
// refX=+x, refY=+z), i.e. viewpoint-bridge.ts worldDirToAngles:
//
//	theta = acos(clamp(d.y, -1, 1)); phi = atan2(d.z, d.x)   (d = normalize(v))
func WorldDirToAngles(v vec3) Dir {
	d := v.Normalize()
	return Dir{
		Theta: math.Acos(Clamp(d.Y, -1, 1)),
		Phi:   math.Atan2(d.Z, d.X),
	}
}
