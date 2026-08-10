package geom

import (
	"math"
	"math/rand"
	"testing"
)

// eye is a test-only oracle for the camera world position: pivot + r along pos.
// (Production Go never does this polar→Cartesian step; the renderer does, at its edge.)
func (v Viewpoint) eye() vec3 {
	return v.Pivot.Add(Polar2cart(Polar{R: v.R, Theta: v.Pos.Theta, Phi: v.Pos.Phi}))
}

func randViewpoint(rng *rand.Rand) Viewpoint {
	return Viewpoint{
		Pivot: vec3{X: rng.Float64() * 100, Y: rng.Float64() * 100, Z: rng.Float64() * 100},
		R:     10 + rng.Float64()*200,
		Pos:   randDir(rng),
		Up:    randDir(rng),
	}
}

func TestViewpointRotateIsRigid(t *testing.T) {
	rng := rand.New(rand.NewSource(11))
	const tol = 1e-7
	for i := 0; i < 3000; i++ {
		v := randViewpoint(rng)
		sep0 := AngularDistance(v.Pos, v.Up) // frame "shape"
		r0, pivot0 := v.R, v.Pivot
		v.Rotate(Rot{Axis: randDir(rng), Angle: (rng.Float64()*2 - 1) * math.Pi})
		// A rotation is rigid: pos↔up separation, radius, and pivot are all unchanged.
		if d := math.Abs(AngularDistance(v.Pos, v.Up) - sep0); d > tol {
			t.Fatalf("rotate changed pos↔up separation by %v", d)
		}
		if v.R != r0 {
			t.Fatalf("rotate changed radius: %v != %v", v.R, r0)
		}
		if v.Pivot != pivot0 {
			t.Fatalf("rotate moved pivot: %v != %v", v.Pivot, pivot0)
		}
	}
}

func TestViewpointOrbitGrabFollows(t *testing.T) {
	rng := rand.New(rand.NewSource(12))
	const tol = 1e-7
	for i := 0; i < 3000; i++ {
		v := randViewpoint(rng)
		from := v.Pos // grab the current position direction
		to := randDir(rng)
		v.Orbit(from, to)
		// The grabbed direction lands on the target.
		if d := AngularDistance(v.Pos, to); d > tol {
			t.Fatalf("orbit: grabbed dir landed at %v, want %v (Δ=%v)", v.Pos, to, d)
		}
	}
}

func TestViewpointZoomClamps(t *testing.T) {
	v := Viewpoint{R: 100}
	v.Zoom(0.5)
	if math.Abs(v.R-50) > 1e-12 {
		t.Fatalf("zoom 0.5 → %v want 50", v.R)
	}
	for i := 0; i < 100; i++ {
		v.Zoom(0.5) // drive far below the floor
	}
	if v.R != ViewpointMinDist {
		t.Fatalf("zoom floor: r=%v want %v", v.R, ViewpointMinDist)
	}
	v.Zoom(3)
	if math.Abs(v.R-3*ViewpointMinDist) > 1e-12 {
		t.Fatalf("zoom out from floor → %v want %v", v.R, 3*ViewpointMinDist)
	}
}

func TestViewpointPanMovesPivotAndEye(t *testing.T) {
	v := Viewpoint{Pivot: vec3{X: 1, Y: 2, Z: 3}, R: 50, Pos: Dir{Theta: 1, Phi: 0.5}, Up: Dir{Theta: 0, Phi: 0}}
	eye0 := v.eye()
	delta := vec3{X: 10, Y: -5, Z: 2}
	v.Pan(delta)
	if v.Pivot != (vec3{X: 11, Y: -3, Z: 5}) {
		t.Fatalf("pan pivot = %v want {11 -3 5}", v.Pivot)
	}
	// Camera rides with the pivot: eye shifts by the same delta.
	if got := v.eye().Sub(eye0); got.Sub(delta).Length() > 1e-9 {
		t.Fatalf("pan eye shift = %v want %v", got, delta)
	}
}
