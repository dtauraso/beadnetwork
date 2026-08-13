package polar

import "math"

type Polar struct {
	R     float64
	Theta float64
	Phi   float64
}

func Polar2cart(p Polar) vec3 {
	st := math.Sin(p.Theta)
	return vec3{
		X: p.R * st * math.Cos(p.Phi),
		Y: p.R * math.Cos(p.Theta),
		Z: p.R * st * math.Sin(p.Phi),
	}
}

// InwardPole is the direction opposite p — the one pointing back at the centre.
//
// It negates the VECTOR and converts, rather than computing pi-theta and
// phi+pi. Those were angle arithmetic, and phi+pi is exactly the sum that could
// land outside atan2's range and so needed wrapping. Coming back through
// Cart2polar, the answer is in range because that is the only range atan2 has.
// The zero-radius guard goes too: negating the zero vector is the zero vector,
// which Cart2polar already answers with {0,0,0}.
func InwardPole(p Polar) (theta, phi float64) {
	back := Cart2polar(Polar2cart(p).Scale(-1))
	return back.Theta, back.Phi
}

// thetaOf is the polar angle down from world +y, for ANY cartesian vector —
// unit or not, zero included.
//
// hypot(x,z) is r·sinθ and y is r·cosθ, so atan2 of the two is θ with r
// cancelled. The first argument is never negative, which is what pins the
// result to [0,π] — the same range acos(y/r) gave, reached without acos's
// domain: no clamp, no NaN when rounding pushes y/r past ±1, and no separate
// zero-radius branch, since atan2(0,0) is 0.
//
// It is also better conditioned exactly where this layout works. dθ/d(cosθ) =
// -1/sinθ blows up at the poles, so the acos form amplified error precisely
// for vectors along the pole axis; atan2 reads both legs instead of one ratio.
// It is UNEXPORTED on purpose. Cart2polar is the one way in, so no caller can
// compute half a conversion by hand — which is how the same two lines came to
// be written in five places.
func thetaOf(v vec3) float64 {
	return math.Atan2(math.Hypot(v.X, v.Z), v.Y)
}

func Cart2polar(v vec3) Polar {
	r := v.Length()
	if r == 0 {
		return Polar{}
	}
	return Polar{R: r, Theta: thetaOf(v), Phi: math.Atan2(v.Z, v.X)}
}

func PolarDist(a, b Polar) float64 {
	cosG := math.Cos(a.Theta)*math.Cos(b.Theta) +
		math.Sin(a.Theta)*math.Sin(b.Theta)*math.Cos(a.Phi-b.Phi)
	d2 := a.R*a.R + b.R*b.R - 2*a.R*b.R*cosG
	if d2 <= 0 {
		return 0
	}
	return math.Sqrt(d2)
}
