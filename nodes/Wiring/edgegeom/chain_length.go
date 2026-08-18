package edgegeom

import (
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	lattice "github.com/dtauraso/wirefold/nodes/bead/lattice"
)

func EdgeStepCount(dist float64, srcKind, dstKind string) int {
	k := int(math.Round(dist / lattice.BeadStepR))
	n := k - nodegeom.NodeTorusSteps(srcKind) - nodegeom.NodeTorusSteps(dstKind)
	if n < 1 {
		return 1
	}
	return n
}
