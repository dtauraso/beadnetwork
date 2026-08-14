package polar

import "math"

type Polar struct {
	R     float64
	Phi   float64
	Theta float64
}

// A triple is a VECTOR — a distance out from wherever it starts, an angle down
// from that start's own +y pole, and a turn around it. The three letters of the
// model are three such vectors closing a triangle:
//
//	A = scene centre -> node        the node's own point
//	D = node -> neighbour           the edge's vector
//	B = scene centre -> neighbour   the neighbour's point
//
//	A + D = B
//
// D starts AT THE NODE. That is what makes its phi the angle from that node's
// own pole to its neighbour — the quantity the out-angle constraint names, read
// straight off the number the constraint is applied to.
//
// Between and Compose are how those vectors are taken apart and put together.
// They are NOT component arithmetic: r+r, phi+phi, theta+theta was tried and is
// not addition of anything. Measured against where the vectors land it was off
// by 447 scene units on a sum, and it made the constraint hold a number that
// was not the angle in the triangle — pinned at exactly 90 degrees while the
// picture sat at 99.79.

// Compose is the vector p followed by the vector q: the three numbers added.
// Between is the vector from one point to another: the three numbers
// subtracted. Neg turns a vector around: the three numbers negated.
//
// Only whole turns of theta are removed afterwards, and that moves nothing.
// NOTHING ELSE IS FOLDED — r may be negative and phi may pass the pole, because
// each of those folds pays for itself by rewriting the other two components,
// which puts the triple somewhere the arithmetic never said. One did exactly
// that and walked a node outward 118 -> 373 -> 1891 across reloads.
//
// These three used to resolve onto the pole axis and the plane across it and
// read the answer back off those legs — sin and cos in, atan2 out, a cartesian
// round trip behind a polar signature. The conversion happens ONCE now, where a
// world point enters the system (Cart2polar) and where one leaves it
// (Polar2cart), and nowhere in between.
func Compose(p, q Polar) Polar {
	return Polar{R: p.R + q.R, Phi: p.Phi + q.Phi, Theta: wrapTurn(p.Theta + q.Theta)}
}

func Between(from, to Polar) Polar {
	return Polar{R: to.R - from.R, Phi: to.Phi - from.Phi, Theta: wrapTurn(to.Theta - from.Theta)}
}

// SnapDeltaTheta puts a MOVE's theta on the nearest integer multiple of pi —
// negative, zero, or positive. It is applied to a delta triple the moment that
// triple is formed, before any rule reads it, so every rule downstream sees a
// theta that is already one of the allowed turns.
//
// Between has already removed whole turns, so the delta arrives in (-pi, pi]
// and the reachable answers are -pi, 0 and pi (and -pi is pi, a half turn
// either way around the same pole). A drag therefore either leaves theta alone
// or turns the vector half a turn; there is no partial turn to accumulate.
//
// Only theta is touched. R and Phi pass through, because this is the same shape
// as the other rules on a delta: one component decided on its own.
func SnapDeltaTheta(d Polar) Polar {
	return Polar{R: d.R, Phi: d.Phi, Theta: wrapTurn(math.Round(d.Theta/math.Pi) * math.Pi)}
}

func (p Polar) Neg() Polar {
	return Polar{R: -p.R, Phi: -p.Phi, Theta: wrapTurn(-p.Theta)}
}

// wrapTurn removes whole turns, which do not move a vector, leaving an angle in
// (-pi, pi] — the range atan2 answers in, so a composed and a negated triple
// are comparable.
func wrapTurn(a float64) float64 {
	const twoPi = 2 * math.Pi
	for a > math.Pi {
		a -= twoPi
	}
	for a <= -math.Pi {
		a += twoPi
	}
	return a
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
