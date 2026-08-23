package PairNode

import "github.com/dtauraso/wirefold/Categories/Chrome/Pills/AngleDropdown"

func (a BuildArgs) LatticePointsSeed() int32 {
	if a.Deps.ClaimLatticeIn == nil {
		return AngleDropdown.DefaultLatticePoints
	}
	return a.Deps.LatticePoints
}

func (a BuildArgs) LatticeIn() <-chan int32 {
	if a.Deps.ClaimLatticeIn == nil {
		return make(chan int32)
	}
	return a.Deps.ClaimLatticeIn(a.Name)
}
