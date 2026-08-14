package dispatch

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
)

func initMoveDispatchUIDefaults(md *MoveDispatch) {
	md.UI.OV = viewstate.DefaultOverlayState()
	md.UI.PN = viewstate.DefaultPanelState()
	md.UI.Speed = 1
	md.UI.ClockDivisor = 1
	md.UI.LatticePoints = scenepersist.DefaultLatticePoints
}

func (md *MoveDispatch) bindUIClosures() {
	md.UI.NodeRowFor = md.RT.NodeRowFor
	mrForLens, uiForLens := &md.MR, &md.UI
	md.UI.DistanceGroupLensFn = func() (float32, float32, float32) {
		return DistanceGroupLens(uiForLens, mrForLens)
	}
}
