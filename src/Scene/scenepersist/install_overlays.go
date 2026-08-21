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

	// The files ARE what the renderer reads, so they hold Go's state from the
	// start rather than only after the first toggle. On a fresh tree the
	// defaults would otherwise live in Go alone, and the renderer would read
	// nothing and draw a scene with every overlay off.
	if err := Overlay.WriteSceneOverlays(dir, ui.OV); err != nil {
		valuefile.LogPersistErr("install_overlays", dir, err)
	}

	ui.EmitViewFrame(nil)
}

func InstallPanels(ui *viewstate.UIState, topologyPath string) {
	pn, _ := Panel.LoadScenePanels(scenepaths.PanelsDirPath(topologyPath))
	ui.PN.SetPanelState(pn)
}
