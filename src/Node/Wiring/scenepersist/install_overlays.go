package scenepersist

import (
	"github.com/dtauraso/wirefold/src/Node/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/src/Node/Wiring/viewstate"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/Panel"
	"github.com/dtauraso/wirefold/src/Overlay"
)

func InstallOverlays(ui *viewstate.UIState, topologyPath string) {
	ov, _ := Overlay.LoadSceneOverlays(scenepaths.OverlaysDirPath(topologyPath))
	ui.OV.SetGuideVisibility(ov)

	ui.EmitViewFrame(nil)
}

func InstallPanels(ui *viewstate.UIState, topologyPath string) {
	pn, _ := Panel.LoadScenePanels(scenepaths.PanelsDirPath(topologyPath))
	ui.PN.SetPanelState(pn)
}
