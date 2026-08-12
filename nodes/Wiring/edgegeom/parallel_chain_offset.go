package edgegeom

import (
	"math"
	"strconv"

	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"
)

func ParallelChainOffset(selfID, targetID string, selfCenter, targetCenter, sceneCenter vec3) (vec3, bool) {
	lowCenter, highCenter := selfCenter, targetCenter
	if !NodeIDLess(selfID, targetID) {
		lowCenter, highCenter = targetCenter, selfCenter
	}
	delta := highCenter.Sub(lowCenter)
	if delta.Length() < 1e-9 {
		return vec3{}, false
	}
	dir := delta.Normalize()

	poleAxis := sceneCenter.Sub(lowCenter)
	if poleAxis.Length() < 1e-9 {

		return vec3{}, false
	}
	pole := poleAxis.Normalize()
	perp := pole.Cross(dir)
	if perp.Length() < 1e-6 {

		alt := vec3{X: 0, Y: 1, Z: 0}
		if math.Abs(pole.Dot(alt)) > 0.9 {
			alt = vec3{X: 1, Y: 0, Z: 0}
		}
		perp = pole.Cross(alt)
	}
	if perp.Length() < 1e-9 {
		return vec3{}, false
	}
	sign := 1.0
	if !NodeIDLess(selfID, targetID) {
		sign = -1.0
	}
	return perp.Normalize().Scale(sign * lattice.BeadStepR), true
}

func NodeIDLess(a, b string) bool {
	ai, aerr := strconv.Atoi(a)
	bi, berr := strconv.Atoi(b)
	if aerr == nil && berr == nil {
		return ai < bi
	}
	return a < b
}
