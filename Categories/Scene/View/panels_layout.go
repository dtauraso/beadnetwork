package View

import (
	Chrome "github.com/dtauraso/beadnetwork/Categories/Chrome"
)

func (ui *UIState) PanelLayout() Chrome.Layout {
	return Chrome.LayoutOf(Chrome.Of{
		ViewW: ui.ViewW, ViewH: ui.ViewH,
		SceneEditable: ui.SceneEditable,
		SceneKinds:    ui.SceneKinds,
		SpeedSlider:   ui.SceneSpeedSlider,
		AnglesPill:    ui.SceneAnglesPill,
		PlusButton:    ui.ScenePlusButton,
		OverlaysDock:  ui.SceneOverlaysDock,
		HiddenFlags:   ui.SceneUntoggledOverlays(),
		LatticePoints: ui.LatticePoints,
		Speed:         ui.Speed,
		Overlays:      &ui.OV,
		Panels:        &ui.PN,
		Sheet:         &ui.Sheet,
		Builder:       ui.Builder,
		Angle:         ui.Angle,
		Nodes:         ui.Nodes,
		Tabs:          ui.TabStrip,
		PillsBar:      ui.OverlaysPill,
	})
}
