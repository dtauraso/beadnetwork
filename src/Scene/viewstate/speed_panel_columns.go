package viewstate

import (
	"github.com/dtauraso/wirefold/src/Chrome/Panels/SliderPanel"
)

func (ui *UIState) writeSpeedPanelColumns(lay SliderPanel.Layout) {
	w := ui.sliderPanelValues
	if w == nil {
		return
	}
	selected := SliderPanel.SelectedIndex(ui.Speed)

	w.Begin()
	w.Rect("boxX", "boxY", "boxW", "boxH", lay.Box)

	for i, r := range lay.Ticks {
		w.Rect("rectX", "rectY", "rectW", "rectH", r)

		var on uint8
		if i == selected {
			on = 1
		}
		w.U8("selected", on)

		s := SliderPanel.Settings[i]
		w.Str("numText", "numLen", s.Num)
		w.Str("denText", "denLen", s.Den)
	}

	w.Rect("trackX", "trackY", "trackW", "trackH", lay.Track)

	if err := w.Flush(); err != nil {
		LogPersistErr("slider_panel_values", "", err)
	}
}
