package polar

import "math"

type Polar struct {
	R     float64
	Phi   float64
	Theta float64
}

func Polar2cart(p Polar) vec3 {
	st := math.Sin(p.Phi)
	return vec3{
		X: p.R * st * math.Cos(p.Theta),
		Y: p.R * math.Cos(p.Phi),
		Z: p.R * st * math.Sin(p.Theta),
	}
}

// InwardPole is the direction opposite p — the one pointing back at the centre.
//
// It negates the VECTOR and converts, rather than computing pi-phi and
// theta+pi. Those were angle arithmetic, and theta+pi is exactly the sum that could
// land outside atan2's range and so needed wrapping. Coming back through
// Cart2polar, the answer is in range because that is the only range atan2 has.
// The zero-radius guard goes too: negating the zero vector is the zero vector,
// which Cart2polar already answers with {0,0,0}.
func InwardPole(p Polar) (phi, theta float64) {
	back := Cart2polar(Polar2cart(p).Scale(-1))
	return back.Phi, back.Theta
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

// The zero vector needs no guard: nothing divides by r any more, and atan2(0,0)
// is 0, so the origin answers {0,0,0} on its own. The guard was load-bearing
// only while phi was acos(y/r).
func Cart2polar(v vec3) Polar {
	return Polar{R: v.Length(), Phi: thetaOf(v), Theta: math.Atan2(v.Z, v.X)}
}

// PolarDist is the distance between two points: the length of the vector from
// one to the other.
//
// It was the spherical law of cosines feeding the law of cosines — trigonometry
// to find a length, with an angle subtraction (a.Theta - b.Theta) inside it. That
// form computes a.R^2 + b.R^2 - 2*a.R*b.R*cos, a difference of two large nearly
// equal numbers for points close together, so rounding could drive the squared
// distance below zero and it needed a guard before taking the square root.
//
// Subtracting the vectors first leaves nothing large to cancel, so the result
// cannot come out negative and there is nothing to guard.
func PolarDist(a, b Polar) float64 {
	return Polar2cart(a).Sub(Polar2cart(b)).Length()
}
