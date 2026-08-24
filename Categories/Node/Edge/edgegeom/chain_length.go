package edgegeom

import (
	"math"

	lattice "github.com/dtauraso/beadnetwork/Categories/Node/BeadAnimation/lattice"
)

func EdgeStepCount(dist float64, srcSteps, dstSteps int) int {
	k := int(math.Round(dist / lattice.SlotR))
	n := k - srcSteps - dstSteps
	if n < 1 {
		return 1
	}
	return n
}
