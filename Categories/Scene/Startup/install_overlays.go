package Startup

import (
	"github.com/dtauraso/beadnetwork/Categories/Scene/View"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/beadnetwork/Categories/Overlay"
)

func InstallOverlays(ui *View.UIState, topologyPath string) {
	ov, _ := Overlay.LoadSceneOverlays(topologyPath)
	ui.OV.SetGuideVisibility(ov)

	if err := Overlay.WriteSceneOverlays(topologyPath, ui.OV); err != nil {
		LogPersistErr("install_overlays", topologyPath, err)
	}

	ui.EmitViewFrame(nil)
}

func InstallPanels(ui *View.UIState, topologyPath string) {
	pn, _ := Panel.LoadScenePanels(topologyPath)
	ui.PN.SetPanelState(pn)
}
