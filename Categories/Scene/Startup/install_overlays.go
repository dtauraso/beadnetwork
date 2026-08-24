package Startup

import (
	"github.com/dtauraso/beadnetwork/Categories/Scene/View"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
	Flags "github.com/dtauraso/beadnetwork/Categories/Scene/View/Flags"
)

func InstallOverlays(ui *View.UIState, topologyPath string) {
	ov, _ := Flags.LoadSceneOverlays(topologyPath)
	ui.OV.SetGuideVisibility(ov)

	if err := Flags.WriteSceneOverlays(topologyPath, ui.OV); err != nil {
		LogPersistErr("install_overlays", topologyPath, err)
	}

	ui.EmitViewFrame(nil)
}

func InstallPanels(ui *View.UIState, topologyPath string) {
	pn, _ := Panel.LoadScenePanels(topologyPath)
	ui.PN.SetPanelState(pn)
}
