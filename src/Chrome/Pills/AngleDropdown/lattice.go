package AngleDropdown

import (
	"encoding/json"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/TiltPanel"
	"github.com/dtauraso/wirefold/src/Node/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/src/Node/Wiring/scenepaths"
)

const DefaultLatticePoints int32 = TiltPanel.FullTurnPhiIdx

func WriteSceneLattice(latticePath string, points int32) error {
	obj := map[string]json.RawMessage{
		"points": json.RawMessage(FormatLatticeJSON(points)),
	}
	return jsonpersist.WriteJSONAtomic(latticePath, obj)
}

func FormatLatticeJSON(points int32) []byte {
	b, err := json.Marshal(points)
	if err != nil {
		b = []byte("24")
	}
	return b
}

type sceneLatticeFile struct {
	Points *int32 `json:"points"`
}

func LoadSceneLattice(latticePath string) (int32, bool) {
	var lf sceneLatticeFile
	jsonpersist.ReadJSONBestEffort(latticePath, &lf)
	if lf.Points == nil {
		return DefaultLatticePoints, false
	}
	return *lf.Points, true
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
