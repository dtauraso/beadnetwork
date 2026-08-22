package edgegeom

import (
	"math"

	lattice "github.com/dtauraso/wirefold/src/Node/BeadAnimation/lattice"
	"github.com/dtauraso/wirefold/src/Node/nodegeom"
)

func EdgeStepCount(dist float64, srcKind, dstKind string) int {
	k := int(math.Round(dist / lattice.SlotR))
	n := k - nodegeom.NodeTorusSteps(srcKind) - nodegeom.NodeTorusSteps(dstKind)
	if n < 1 {
		return 1
	}
	return n
}
