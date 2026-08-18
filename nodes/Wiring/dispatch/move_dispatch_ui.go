package dispatch

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
	"github.com/dtauraso/wirefold/tools/topology-vscode/AngleDropdown"
)

func initMoveDispatchUIDefaults(md *MoveDispatch) {
	md.UI.OV = viewstate.DefaultOverlayState()
	md.UI.PN = viewstate.DefaultPanelState()
	md.UI.Speed = 1
	md.UI.ClockDivisor = 1
	md.UI.LatticePoints = AngleDropdown.DefaultLatticePoints
}

func (md *MoveDispatch) bindUIClosures() {
	md.UI.NodeRowFor = md.RT.NodeRowFor
}
