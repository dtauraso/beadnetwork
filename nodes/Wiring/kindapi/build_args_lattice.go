package kindapi

import "github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"

func (a BuildArgs) LatticePointsSeed() int32 {
	if a.deps.ClaimLatticeIn == nil {
		return scenepersist.DefaultLatticePoints
	}
	return a.deps.LatticePoints
}

func (a BuildArgs) LatticeIn() <-chan int32 {
	if a.deps.ClaimLatticeIn == nil {
		return make(chan int32)
	}
	return a.deps.ClaimLatticeIn(a.name)
}
