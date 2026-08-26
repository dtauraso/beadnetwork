package NodePhi

import "github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/AngleDropdown"

func (a BuildArgs) LatticePointsSeed() int32 {
	if a.Deps == nil {
		return AngleDropdown.DefaultLatticePoints
	}
	return a.Deps.LatticePointsSeed()
}

func (a BuildArgs) LatticeIn() <-chan int32 {
	if a.Deps == nil {
		return make(chan int32)
	}
	return a.Deps.LatticeChan(a.Name)
}
