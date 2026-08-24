package input

import lattice "github.com/dtauraso/beadnetwork/Categories/Node/BeadAnimation/lattice"

func popEnd(working, backup *[]int, init []int) int {
	v := (*working)[len(*working)-1]
	*working = (*working)[:len(*working)-1]
	if len(*working) == 0 {

		*working = *backup
		*backup = append([]int(nil), init...)
	}
	return v
}

func cadenceTicks(steps int) int64 {
	c := int64(float64(steps) * lattice.PulsesPerSlot)
	if c < 1 {
		return 1
	}
	return c
}
