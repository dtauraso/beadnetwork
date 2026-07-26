package Wiring

import (
	"math"
	"testing"
)

// roundTrip cart→polar→cart must return the original vector within epsilon.
func TestPolarCartRoundTrip(t *testing.T) {
	cases := []vec3{
		{X: 1, Y: 0, Z: 0},   // +x (equator, φ=0)
		{X: 0, Y: 1, Z: 0},   // +y (north pole)
		{X: 0, Y: -1, Z: 0},  // -y (south pole)
		{X: 0, Y: 0, Z: 1},   // +z (equator, φ=π/2)
		{X: 0, Y: 0, Z: -1},  // -z
		{X: -1, Y: 0, Z: 0},  // -x
		{X: 3, Y: 4, Z: 12},  // arbitrary, length 13
		{X: -5, Y: 2, Z: -7}, // arbitrary negative octant
		{X: 0, Y: 0, Z: 0},   // origin
	}
	const eps = 1e-9
	for _, v := range cases {
		got := polar2cart(cart2polar(v))
		if math.Abs(got.X-v.X) > eps || math.Abs(got.Y-v.Y) > eps || math.Abs(got.Z-v.Z) > eps {
			t.Errorf("round-trip %v -> %v", v, got)
		}
	}
}

// Known polar values convert to the expected Cartesian.
func TestPolar2CartKnown(t *testing.T) {
	const eps = 1e-9
	// θ=π/2 (equator), φ=0 -> +x at radius r
	got := polar2cart(polar{R: 2, Theta: math.Pi / 2, Phi: 0})
	if math.Abs(got.X-2) > eps || math.Abs(got.Y) > eps || math.Abs(got.Z) > eps {
		t.Errorf("equator φ=0: got %v want (2,0,0)", got)
	}
	// θ=0 -> +y pole
	got = polar2cart(polar{R: 5, Theta: 0, Phi: 1.234})
	if math.Abs(got.X) > eps || math.Abs(got.Y-5) > eps || math.Abs(got.Z) > eps {
		t.Errorf("north pole: got %v want (0,5,0)", got)
	}
	// θ=π/2, φ=π/2 -> +z
	got = polar2cart(polar{R: 3, Theta: math.Pi / 2, Phi: math.Pi / 2})
	if math.Abs(got.X) > eps || math.Abs(got.Y) > eps || math.Abs(got.Z-3) > eps {
		t.Errorf("equator φ=π/2: got %v want (0,0,3)", got)
	}
}

// polar2cart symmetry: flipping the sign of φ (azimuth about +y) flips only z;
// x and y are unchanged. (A pure coordinate property; no longer used by a lock.)
func TestPolarMirrorPhiFlipsOnlyZ(t *testing.T) {
	const eps = 1e-9
	p := polar{R: 4, Theta: 1.1, Phi: 0.7}
	a := polar2cart(p)
	b := polar2cart(polar{R: p.R, Theta: p.Theta, Phi: -p.Phi})
	if math.Abs(a.X-b.X) > eps || math.Abs(a.Y-b.Y) > eps {
		t.Errorf("mirror_φ changed x or y: %v vs %v", a, b)
	}
	if math.Abs(a.Z+b.Z) > eps {
		t.Errorf("mirror_φ did not negate z: %v vs %v", a, b)
	}
}
