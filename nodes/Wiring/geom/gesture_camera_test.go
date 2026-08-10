package geom

import (
	"math"
	"testing"
)

// gesture_camera_test.go — CROSS-CHECK the Go camera-math port (gesture_camera.go) against
// the TS formulas it mirrors. Each test either (a) hand-transcribes the TS arithmetic to an
// independent expected value, or (b) checks a hardcoded oracle case computed by hand from the
// TS source. This guards against reintroducing the arcball / wrong-axis bug class by pinning
// the ported pixel→world-direction pipeline to the TS values.

func vecClose(a, b vec3, tol float64) bool {
	return math.Abs(a.X-b.X) < tol && math.Abs(a.Y-b.Y) < tol && math.Abs(a.Z-b.Z) < tol
}

// AnglesToWorldOffset must equal production Polar2cart (the same formula the renderer uses),
// and the TS viewpoint-bridge.ts anglesToWorldOffset.
func TestGestureAnglesToWorldOffsetMatchesPolar2Cart(t *testing.T) {
	cases := []struct{ r, th, ph float64 }{
		{1, math.Pi / 2, math.Pi / 2}, {5, 0, 1.234}, {3, math.Pi / 2, 0}, {2, 1.1, -0.7},
	}
	for _, c := range cases {
		got := AnglesToWorldOffset(c.r, c.th, c.ph)
		want := Polar2cart(Polar{R: c.r, Theta: c.th, Phi: c.ph})
		if !vecClose(got, want, 1e-12) {
			t.Fatalf("AnglesToWorldOffset(%v)=%v want %v", c, got, want)
		}
	}
}

// WorldDirToAngles: hand oracle. dir +Z → theta=acos(0)=π/2, phi=atan2(1,0)=π/2.
func TestGestureWorldDirToAngles(t *testing.T) {
	d := WorldDirToAngles(vec3{X: 0, Y: 0, Z: 1})
	if math.Abs(d.Theta-math.Pi/2) > 1e-12 || math.Abs(d.Phi-math.Pi/2) > 1e-12 {
		t.Fatalf("+Z → %v want {π/2, π/2}", d)
	}
	// Round-trip: AnglesToWorldOffset(1, θ, φ) then back must recover (θ, φ).
	in := Dir{Theta: 1.0, Phi: -0.5}
	back := WorldDirToAngles(AnglesToWorldOffset(1, in.Theta, in.Phi))
	if math.Abs(back.Theta-in.Theta) > 1e-9 || math.Abs(back.Phi-in.Phi) > 1e-9 {
		t.Fatalf("round-trip: %v → %v", in, back)
	}
}

// BasisFromViewpoint hardcoded oracle: camera at +Z (pos=+Z), up=+Y →
// three.js lookAt basis refX=+X, refY=+Y, pole=+Z.
func TestGestureBasisOracle(t *testing.T) {
	pos := Dir{Theta: math.Pi / 2, Phi: math.Pi / 2} // +Z
	up := Dir{Theta: 0, Phi: 0}                      // +Y
	b := BasisFromViewpoint(pos, up)
	if !vecClose(b.Pole, vec3{X: 0, Y: 0, Z: 1}, 1e-12) {
		t.Fatalf("pole=%v want +Z", b.Pole)
	}
	if !vecClose(b.RefX, vec3{X: 1, Y: 0, Z: 0}, 1e-12) {
		t.Fatalf("refX=%v want +X", b.RefX)
	}
	if !vecClose(b.RefY, vec3{X: 0, Y: 1, Z: 0}, 1e-12) {
		t.Fatalf("refY=%v want +Y", b.RefY)
	}
}

// BasisFromViewpoint must be orthonormal (right-handed) for random viewpoints — the
// property the TS cameraFrame relies on (quaternion basis).
func TestGestureBasisOrthonormal(t *testing.T) {
	poses := []Dir{{Theta: 0.3, Phi: 0.7}, {Theta: 1.9, Phi: -2.1}, {Theta: math.Pi / 2, Phi: 0}, {Theta: 2.4, Phi: 1.1}}
	ups := []Dir{{Theta: 1.1, Phi: 0.2}, {Theta: 0.4, Phi: 2.9}, {Theta: 0, Phi: 0}, {Theta: 1.5, Phi: -1.2}}
	for i := range poses {
		b := BasisFromViewpoint(poses[i], ups[i])
		for _, v := range []vec3{b.RefX, b.RefY, b.Pole} {
			if math.Abs(v.Length()-1) > 1e-9 {
				t.Fatalf("basis vector not unit: %v (|v|=%v)", v, v.Length())
			}
		}
		if math.Abs(b.RefX.Dot(b.Pole)) > 1e-9 || math.Abs(b.RefX.Dot(b.RefY)) > 1e-9 || math.Abs(b.RefY.Dot(b.Pole)) > 1e-9 {
			t.Fatalf("basis not orthogonal: %+v", b)
		}
		// refY == pole × refX (right-handed)
		if !vecClose(b.RefY, b.Pole.Cross(b.RefX), 1e-9) {
			t.Fatalf("not right-handed: refY=%v pole×refX=%v", b.RefY, b.Pole.Cross(b.RefX))
		}
	}
}

// ScreenToPolar + ToWorldDir hand oracles in the canonical (+X,+Y,+Z) frame.
func TestGestureScreenToWorldOracle(t *testing.T) {
	b := CamBasis{RefX: vec3{X: 1, Y: 0, Z: 0}, RefY: vec3{X: 0, Y: 1, Z: 0}, Pole: vec3{X: 0, Y: 0, Z: 1}}
	// cursor exactly at center → pole direction (+Z, toward camera).
	center := ToWorldDir(b, ScreenToPolar(0, 0, 100))
	if !vecClose(center, vec3{X: 0, Y: 0, Z: 1}, 1e-12) {
		t.Fatalf("center cursor → %v want +Z", center)
	}
	// cursor one scale-unit to the RIGHT: dx=scale, dy=0 → phi=1, theta=atan2(0,1)=0 →
	// equator=+X, dir=(sin1, 0, cos1).
	right := ToWorldDir(b, ScreenToPolar(100, 0, 100))
	want := vec3{X: math.Sin(1), Y: 0, Z: math.Cos(1)}
	if !vecClose(right, want, 1e-12) {
		t.Fatalf("right cursor → %v want %v", right, want)
	}
	// cursor one scale-unit UP: screen dy is negative up in client coords; the handler
	// passes (y - cy). A point ABOVE center has dyFromCenter<0. theta=atan2(-(-1),0)=π/2 →
	// equator=+Y, dir=(0, sin1, cos1).
	up := ToWorldDir(b, ScreenToPolar(0, -100, 100))
	wantUp := vec3{X: 0, Y: math.Sin(1), Z: math.Cos(1)}
	if !vecClose(up, wantUp, 1e-12) {
		t.Fatalf("up cursor → %v want %v", up, wantUp)
	}
}

// PlaneSlide hand oracle: refX=+X, refY=+Y; (r=2, angle=0, wpp=3) → (6,0,0).
func TestGesturePlaneSlideOracle(t *testing.T) {
	b := CamBasis{RefX: vec3{X: 1, Y: 0, Z: 0}, RefY: vec3{X: 0, Y: 1, Z: 0}, Pole: vec3{X: 0, Y: 0, Z: 1}}
	got := PlaneSlide(b, 2, 0, 3)
	if !vecClose(got, vec3{X: 6, Y: 0, Z: 0}, 1e-12) {
		t.Fatalf("PlaneSlide=%v want (6,0,0)", got)
	}
	got = PlaneSlide(b, 2, math.Pi/2, 3)
	if !vecClose(got, vec3{X: 0, Y: 6, Z: 0}, 1e-9) {
		t.Fatalf("PlaneSlide(π/2)=%v want (0,6,0)", got)
	}
}

func TestGestureDeltaToPolar(t *testing.T) {
	r, a := DeltaToPolar(3, 4)
	if math.Abs(r-5) > 1e-12 || math.Abs(a-math.Atan2(4, 3)) > 1e-12 {
		t.Fatalf("DeltaToPolar(3,4)=(%v,%v)", r, a)
	}
}

// ContentSphereOf hand oracle: centers (0,0,0),(10,0,0) → center (5,0,0), radius 5*1.1=5.5.
func TestGestureContentSphereOracle(t *testing.T) {
	c, r := ContentSphereOf(map[string]vec3{"a": {X: 0, Y: 0, Z: 0}, "b": {X: 10, Y: 0, Z: 0}})
	if !vecClose(c, vec3{X: 5, Y: 0, Z: 0}, 1e-12) || math.Abs(r-5.5) > 1e-12 {
		t.Fatalf("ContentSphere=%v r=%v want (5,0,0) 5.5", c, r)
	}
	c, r = ContentSphereOf(nil)
	if !vecClose(c, vec3{}, 1e-12) || r != 100 {
		t.Fatalf("empty ContentSphere=%v r=%v want origin 100", c, r)
	}
}

// RegionFocus hand oracle: empty centers → eye + forward*FOCUS_MIN. Camera at +Z, r=100:
// eye=(0,0,100), forward=-pole=(0,0,-1) → focus=(0,0,90).
func TestGestureRegionFocusEmpty(t *testing.T) {
	v := Viewpoint{Pivot: vec3{X: 0, Y: 0, Z: 0}, R: 100, Pos: Dir{Theta: math.Pi / 2, Phi: math.Pi / 2}, Up: Dir{Theta: 0, Phi: 0}}
	f := RegionFocus(v, nil)
	if !vecClose(f, vec3{X: 0, Y: 0, Z: 90}, 1e-9) {
		t.Fatalf("RegionFocus(empty)=%v want (0,0,90)", f)
	}
}

// ProjectNDC hand oracle: point at origin, camera at +Z looking at origin → NDC (0,0), inFront.
func TestGestureProjectNDCOracle(t *testing.T) {
	v := Viewpoint{Pivot: vec3{X: 0, Y: 0, Z: 0}, R: 100, Pos: Dir{Theta: math.Pi / 2, Phi: math.Pi / 2}, Up: Dir{Theta: 0, Phi: 0}}
	b := BasisFromViewpoint(v.Pos, v.Up)
	nx, ny, inFront := ProjectNDC(vec3{X: 0, Y: 0, Z: 0}, EyeOf(v), b, 50, 800.0/600.0)
	if !inFront || math.Abs(nx) > 1e-9 || math.Abs(ny) > 1e-9 {
		t.Fatalf("ProjectNDC(origin)=(%v,%v,inFront=%v) want (0,0,true)", nx, ny, inFront)
	}
	// A point behind the camera (further +Z than the eye) is not in front.
	_, _, inFront2 := ProjectNDC(vec3{X: 0, Y: 0, Z: 200}, EyeOf(v), b, 50, 800.0/600.0)
	if inFront2 {
		t.Fatalf("point behind camera reported inFront")
	}
}
