package AngleDropdown

import (
	"github.com/dtauraso/wirefold/Chrome/Panels/TiltPanel"
	"github.com/dtauraso/wirefold/Scene/scenepaths"
)

const DefaultLatticePoints int32 = TiltPanel.FullTurnPhiIdx

func WriteSceneLattice(latticePath string, points int32) error {
	return WriteAtomic(latticePath, points)
}

func LoadSceneLattice(latticePath string) (int32, bool) {
	var points int32
	if !ReadIfExists(latticePath, &points) {
		return DefaultLatticePoints, false
	}
	return points, true
}

func LatticePointsFor(topologyPath string) int32 {
	points, _ := LoadSceneLattice(scenepaths.LatticeFilePath(topologyPath))
	return points
}

func SendLatticePointsNonBlocking(ch chan int32, points int32) {
	select {
	case ch <- points:
		return
	default:
	}
	select {
	case <-ch:
	default:
	}
	select {
	case ch <- points:
	default:
	}
}
