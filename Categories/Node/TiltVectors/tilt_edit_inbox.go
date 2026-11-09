package TiltVectors

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/AngleDropdown"
)

type TiltEditInboxes struct {
	lattice map[string]chan int32
}

func (ib *TiltEditInboxes) ClaimLatticeIn(id string, ch chan int32) {
	if ib.lattice == nil {
		ib.lattice = map[string]chan int32{}
	}
	ib.lattice[id] = ch
}

func (ib *TiltEditInboxes) BroadcastLatticePoints(points int32) {
	for _, ch := range ib.lattice {
		AngleDropdown.SendLatticePointsNonBlocking(ch, points)
	}
}
