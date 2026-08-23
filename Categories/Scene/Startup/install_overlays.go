package Startup

import (
	"github.com/dtauraso/wirefold/Categories/Scene/viewstate"

	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/wirefold/Categories/Overlay"
)

func InstallOverlays(ui *viewstate.UIState, topologyPath string) {
	ov, _ := Overlay.LoadSceneOverlays(topologyPath)
	ui.OV.SetGuideVisibility(ov)

	if err := Overlay.WriteSceneOverlays(topologyPath, ui.OV); err != nil {
		LogPersistErr("install_overlays", topologyPath, err)
	}

	ui.EmitViewFrame(nil)
}

func InstallPanels(ui *viewstate.UIState, topologyPath string) {
	pn, _ := Panel.LoadScenePanels(topologyPath)
	ui.PN.SetPanelState(pn)
}
