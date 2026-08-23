package Tabs

import (
	"fmt"
	"os"
)

func WriteValues(w *ValueWriter, lay Layout) {
	if w == nil {
		return
	}
	w.Begin()

	w.Rect("stripX", "stripY", "stripW", "stripH", lay.Strip)

	for _, t := range lay.Tabs {
		w.Rect("tabX", "tabY", "tabW", "tabH", t.Rect)
		w.Str("tabNameText", "tabNameLen", t.Name)
		w.Bool("tabSelected", t.Selected)
	}

	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "tab_strip_values: %v\n", err)
	}
}
