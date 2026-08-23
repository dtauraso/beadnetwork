package SliderPanel

import (
	"fmt"
	"os"
)

func WriteValues(w *ValueWriter, lay Layout, speed float64) {
	if w == nil {
		return
	}
	selected := SelectedIndex(speed)

	w.Begin()
	w.Rect("boxX", "boxY", "boxW", "boxH", lay.Box)

	for i, r := range lay.Ticks {
		w.Rect("rectX", "rectY", "rectW", "rectH", r)

		var on uint8
		if i == selected {
			on = 1
		}
		w.U8("selected", on)

		s := Settings[i]
		w.Str("numText", "numLen", s.Num)
		w.Str("denText", "denLen", s.Den)
	}

	w.Rect("trackX", "trackY", "trackW", "trackH", lay.Track)

	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "slider_panel_values: %v\n", err)
	}
}
