package dispatch

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/distancegroups"
	"github.com/dtauraso/wirefold/nodes/Wiring/moverreg"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
)

func DistanceGroupLens(ui *viewstate.UIState, mr *moverreg.MoverRegistry) (timeLen, inputLen, gateLen float32) {
	return distancegroups.Lens(ui.HasDistanceGroups, mr.CenterOfNode)
}
