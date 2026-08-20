package scenepersist

import (
	"github.com/dtauraso/wirefold/src/Node/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/src/Node/Wiring/viewstate"

	"github.com/dtauraso/wirefold/src/Chrome/OverlaysDropdown"
	T "github.com/dtauraso/wirefold/src/Trace"
)

func InstallOverlays(ui *viewstate.UIState, topologyPath string, tr *T.Trace) {
	ov, _ := OverlaysDropdown.LoadSceneOverlays(scenepaths.OverlaysDirPath(topologyPath))
	ui.OV.SetGuideVisibility(ov)

	ui.EmitViewFrame(nil)
}

func InstallPanels(ui *viewstate.UIState, topologyPath string) {
	pn, _ := OverlaysDropdown.LoadScenePanels(scenepaths.PanelsDirPath(topologyPath))
	ui.PN.SetPanelState(pn)
}
