package View

import (
	Chrome "github.com/dtauraso/beadnetwork/Categories/Chrome"
)

func (ui *UIState) PanelLayout() Chrome.Layout {
	return Chrome.LayoutOf(Chrome.Of{
		ViewW: ui.ViewW, ViewH: ui.ViewH,
		SceneEditable: ui.SceneEditable,
		SceneKinds:    ui.SceneKinds,
		LatticePoints: ui.LatticePoints,
		Overlays:      &ui.OV,
		Panels:        &ui.PN,
		Tilt:          ui.Tilt,
		Angle:         ui.Angle,
		Nodes:         ui.Nodes,
		Tabs:          ui.TabStrip,
		Rules:         ui.Rules,
		PillsBar:      ui.OverlaysPill,
	})
}
