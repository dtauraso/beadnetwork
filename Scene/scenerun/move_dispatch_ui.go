package scenerun

import (
	"github.com/dtauraso/wirefold/Chrome/Panels/Panel"
	"github.com/dtauraso/wirefold/Chrome/Pills/AngleDropdown"
	"github.com/dtauraso/wirefold/Overlay"
)

func initMoveDispatchUIDefaults(md *MoveDispatch) {
	md.UI.OV = Overlay.DefaultOverlayState()
	md.UI.PN = Panel.DefaultPanelState()
	md.UI.Speed = 1
	md.UI.ClockDivisor = 1
	md.UI.LatticePoints = AngleDropdown.DefaultLatticePoints
}

func (md *MoveDispatch) bindUIClosures() {
	md.UI.NodeRowFor = md.RT.NodeRowFor
}
