package Wiring

import (
	"math"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// bead_cell_solve.go — the bead-CELL solver: a node lives in N lattices, one per
// neighbour it has an edge with. For neighbour j with centre C_j, the admissible node
// centres are the concentric spheres
//
//	dist(node, C_j) == K_j * wire.BeadStepR      K_j a positive integer
//
// and the node's centre must satisfy this for EVERY neighbour simultaneously — it sits at
// an INTERSECTION of all N sphere families. Moving from K_j to K_j±1 adds or removes
// exactly one bead on that edge.
//
// This file is pure geometry: input is a neighbour set (each neighbour's centre and its
// CURRENT admissible integer K) plus a mouse-derived target; output is every admissible
// candidate NODE CENTRE reachable by moving each neighbour's K by -1, 0 or +1
// independently (a "neighbouring cell", not a global search over all integers). The caller
// (quantized_move.go) picks whichever candidate lands nearest the target, or leaves the
// node where it was if no candidate exists.
//
// N==1: one sphere — every point on it is admissible; direction is free, radius quantised.
// N==2: intersection of two spheres is a circle.
// N>=3: three spheres pin the point (0, 1 or 2 solutions); a 4th+ neighbour is a
// constraint a candidate must satisfy within a small tolerance, not a new equation to
// solve — discard candidates that don't satisfy it.

// beadCellNeighbor is one neighbour's centre and the CURRENT integer K (its live distance
// from the node, in whole wire.BeadStepR multiples) the enumeration steps outward/inward
// from.
type beadCellNeighbor struct {
	Center vec3
	K      int
}

// beadCellTol is the slack a 4th+ constraint is allowed before a candidate is discarded,
// and the slack sphere-sphere/circle-sphere solves use to treat a near-miss (float error
// off an exact tangency) as a hit. Expressed as a fraction of one bead step so it scales
// with the lattice rather than being an arbitrary world-unit constant.
const beadCellTol = 1e-6 * wire.BeadStepR

// solveBeadCells enumerates the admissible node-centre candidates reachable from the
// node's CURRENT configuration: for every neighbour, K_j-1, K_j or K_j+1 (K clamped to
// stay >= 1 — a node's distance to a neighbour is never zero or negative), the product of
// those per-neighbour choices, each solved for its sphere-intersection point(s). The
// all-unchanged combination (every delta zero) always reproduces the node's OWN current
// admissible position as one of its candidates (see commitNodeMoveLocal's call site,
// which is what guarantees this enumeration is never empty for a node already on its
// lattice), so a normal drag always has somewhere to land; a candidate list can still come
// back empty for a genuinely degenerate neighbour set (e.g. two neighbours farther apart
// than the sum of every reachable radius pair), and that must be observable rather than
// silently keeping the node still with no signal — see commitNodeMoveLocal's tr.Breadcrumb
// on the empty case.
func solveBeadCells(neighbors []beadCellNeighbor, target vec3) []vec3 {
	n := len(neighbors)
	if n == 0 {
		// No constraints at all: the node is free — every point is "admissible", so the
		// nearest admissible point to the target is the target itself.
		return []vec3{target}
	}
	var out []vec3
	ks := make([]int, n)
	var rec func(i int)
	rec = func(i int) {
		if i == n {
			out = append(out, intersectSpheres(neighbors, ks, target)...)
			return
		}
		for _, d := range [3]int{-1, 0, 1} {
			k := neighbors[i].K + d
			if k < 1 {
				continue
			}
			ks[i] = k
			rec(i + 1)
		}
	}
	rec(0)
	return out
}

// intersectSpheres solves for the candidate point(s) satisfying every neighbour's sphere
// at the given per-neighbour K (radius K*BeadStepR), for ONE specific K combination.
func intersectSpheres(neighbors []beadCellNeighbor, ks []int, target vec3) []vec3 {
	n := len(neighbors)
	radius := func(i int) float64 { return float64(ks[i]) * wire.BeadStepR }

	switch {
	case n == 1:
		return []vec3{nearestOnSphere(neighbors[0].Center, radius(0), target)}

	case n == 2:
		center, u, v, h, ok := sphereSphereCircle(neighbors[0].Center, radius(0), neighbors[1].Center, radius(1))
		if !ok {
			return nil
		}
		if h <= beadCellTol {
			// Degenerate to (essentially) a single point: the two spheres are tangent.
			return []vec3{center}
		}
		return []vec3{nearestOnCircle(center, u, v, h, target)}

	default: // n >= 3
		center, u, v, h, ok := sphereSphereCircle(neighbors[0].Center, radius(0), neighbors[1].Center, radius(1))
		if !ok {
			return nil
		}
		var pts []vec3
		if h <= beadCellTol {
			pts = []vec3{center}
		} else {
			pts = intersectCircleSphere(center, u, v, h, neighbors[2].Center, radius(2))
		}
		var out []vec3
		for _, p := range pts {
			ok := true
			for i := 3; i < n; i++ {
				if math.Abs(p.Sub(neighbors[i].Center).Length()-radius(i)) > beadCellTol {
					ok = false
					break
				}
			}
			// Also re-check the first two constraints, which the circle solve satisfies
			// only up to float error, and the third, which intersectCircleSphere derived
			// rather than checked directly for the h<=tol branch.
			if ok && math.Abs(p.Sub(neighbors[2].Center).Length()-radius(2)) > beadCellTol {
				ok = false
			}
			if ok {
				out = append(out, p)
			}
		}
		return out
	}
}

// nearestOnSphere returns the point on the sphere (center, r) nearest to target. Direction
// is free when target coincides with center; an arbitrary fixed axis is used in that
// degenerate case since no direction is preferred.
func nearestOnSphere(center vec3, r float64, target vec3) vec3 {
	dir := target.Sub(center)
	if dir.Length() < beadCellTol {
		dir = vec3{X: 0, Y: 1, Z: 0}
	}
	return center.Add(dir.Normalize().Scale(r))
}

// nearestOnCircle returns the point on the circle (center, radius h, in the plane spanned
// by orthonormal u,v) nearest to target.
func nearestOnCircle(center vec3, u, v vec3, h float64, target vec3) vec3 {
	rel := target.Sub(center)
	x, y := u.Dot(rel), v.Dot(rel)
	norm := math.Hypot(x, y)
	if norm < beadCellTol {
		// target projects to the circle's own center: direction is free, pick u.
		return center.Add(u.Scale(h))
	}
	return center.Add(u.Scale(h * x / norm)).Add(v.Scale(h * y / norm))
}

// sphereSphereCircle returns the circle of intersection of two spheres (c1,r1) and
// (c2,r2): its center, an orthonormal basis (u,v) of its plane, and its radius h. ok is
// false when the spheres do not intersect (too far apart, one inside the other with no
// touch, or concentric).
func sphereSphereCircle(c1 vec3, r1 float64, c2 vec3, r2 float64) (center, u, v vec3, h float64, ok bool) {
	d3 := c2.Sub(c1)
	dist := d3.Length()
	if dist < beadCellTol {
		return vec3{}, vec3{}, vec3{}, 0, false
	}
	if dist > r1+r2+beadCellTol || dist < math.Abs(r1-r2)-beadCellTol {
		return vec3{}, vec3{}, vec3{}, 0, false
	}
	a := (dist*dist + r1*r1 - r2*r2) / (2 * dist)
	h2 := r1*r1 - a*a
	if h2 < 0 {
		h2 = 0
	}
	h = math.Sqrt(h2)
	normal := d3.Scale(1 / dist)
	center = c1.Add(normal.Scale(a))
	u, v = orthonormalBasis(normal)
	return center, u, v, h, true
}

// intersectCircleSphere returns the point(s) where the circle (center, radius h, in the
// plane spanned by orthonormal u,v) meets the sphere (c3, r3). Parametrise the circle as
// P(t) = center + h·cos(t)·u + h·sin(t)·v and solve |P(t)-c3| = r3, which reduces to
//
//	a·cos(t) + b·sin(t) = rhs      a = u·(center-c3), b = v·(center-c3)
//
// via a·cos+b·sin having amplitude sqrt(a²+b²); at most two roots in [0,2π).
func intersectCircleSphere(center, u, v vec3, h float64, c3 vec3, r3 float64) []vec3 {
	rel := center.Sub(c3)
	// |P(t)-c3|^2 = |rel|^2 + h^2 + 2h[cos(t)(u.rel) + sin(t)(v.rel)] = r3^2, so
	// 2h(u.rel)*cos(t) + 2h(v.rel)*sin(t) = r3^2 - |rel|^2 - h^2 — the 2h factor folds
	// into a/b below so `amp` and `rhs` are directly comparable without a stray factor
	// of 2 anywhere.
	a := 2 * h * u.Dot(rel)
	b := 2 * h * v.Dot(rel)
	amp := math.Hypot(a, b)
	rhs := r3*r3 - rel.Dot(rel) - h*h
	if amp < beadCellTol {
		// center-c3 is along the plane's normal: every point on the circle is
		// equidistant from c3. Either every point solves it (rhs ~ 0, but that is a
		// whole circle of solutions, not a discrete cell — treat as no admissible
		// discrete point) or none do.
		return nil
	}
	ratio := rhs / amp
	if ratio > 1+beadCellTol || ratio < -1-beadCellTol {
		return nil
	}
	ratio = clamp(ratio, -1, 1)
	phase := math.Atan2(b, a)
	delta := math.Acos(ratio)
	pointAt := func(t float64) vec3 {
		return center.Add(u.Scale(h * math.Cos(t))).Add(v.Scale(h * math.Sin(t)))
	}
	if delta < beadCellTol {
		return []vec3{pointAt(phase)}
	}
	return []vec3{pointAt(phase + delta), pointAt(phase - delta)}
}

// orthonormalBasis returns two unit vectors u, v orthogonal to each other and to the given
// unit vector n, spanning n's perpendicular plane.
func orthonormalBasis(n vec3) (u, v vec3) {
	arb := vec3{X: 1, Y: 0, Z: 0}
	if math.Abs(n.X) > 0.9 {
		arb = vec3{X: 0, Y: 1, Z: 0}
	}
	u = n.Cross(arb).Normalize()
	v = n.Cross(u)
	return u, v
}

// snapToBeadCell is the drag placement rule: given the node's own current world center
// (prev), its live neighbour constraints, and the mouse-derived target, move to whichever
// admissible bead-cell candidate (solveBeadCells) lands nearest the target. Returns prev
// unchanged when no candidate exists — the node holds its position rather than landing off
// every lattice.
func snapToBeadCell(prev vec3, neighbors []beadCellNeighbor, target vec3) vec3 {
	cands := solveBeadCells(neighbors, target)
	if len(cands) == 0 {
		return prev
	}
	best := cands[0]
	bestD := best.Sub(target).Length()
	for _, c := range cands[1:] {
		if d := c.Sub(target).Length(); d < bestD {
			best, bestD = c, d
		}
	}
	return best
}
