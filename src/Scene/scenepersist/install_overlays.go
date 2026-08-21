package scenepersist

import (
	"github.com/dtauraso/wirefold/src/Scene/scenepaths"
	"github.com/dtauraso/wirefold/src/Scene/viewstate"
	"github.com/dtauraso/wirefold/src/valuefile"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/Panel"
	"github.com/dtauraso/wirefold/src/Overlay"
)

func InstallOverlays(ui *viewstate.UIState, topologyPath string) {
	dir := scenepaths.OverlaysDirPath(topologyPath)
	ov, _ := Overlay.LoadSceneOverlays(dir)
	ui.OV.SetGuideVisibility(ov)

	if err := Overlay.WriteSceneOverlays(dir, ui.OV); err != nil {
		valuefile.LogPersistErr("install_overlays", dir, err)
	}

	ui.EmitViewFrame(nil)
}

func InstallPanels(ui *viewstate.UIState, topologyPath string) {
	pn, _ := Panel.LoadScenePanels(scenepaths.PanelsDirPath(topologyPath))
	ui.PN.SetPanelState(pn)
}
