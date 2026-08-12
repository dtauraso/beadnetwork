package nodegeom

import (
	"math"

	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"
)

func EdgeStepCount(dist float64, srcKind, dstKind string) int {
	k := int(math.Round(dist / lattice.BeadStepR))
	n := k - NodeTorusSteps(srcKind) - NodeTorusSteps(dstKind)
	if n < 1 {
		return 1
	}
	return n
}
