package dispatch

import (
	"github.com/dtauraso/wirefold/src/Chrome/AngleDropdown"
	"github.com/dtauraso/wirefold/src/Chrome/OverlaysDropdown"
)

func initMoveDispatchUIDefaults(md *MoveDispatch) {
	md.UI.OV = OverlaysDropdown.DefaultOverlayState()
	md.UI.PN = OverlaysDropdown.DefaultPanelState()
	md.UI.Speed = 1
	md.UI.ClockDivisor = 1
	md.UI.LatticePoints = AngleDropdown.DefaultLatticePoints
}

func (md *MoveDispatch) bindUIClosures() {
	md.UI.NodeRowFor = md.RT.NodeRowFor
}
