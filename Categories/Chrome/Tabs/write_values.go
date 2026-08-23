package Tabs

import (
	"fmt"
	"os"
)

// WriteValues writes the tab strip's own block. It takes the strip's writer and
// the strip's layout — not the view state, which would let it write anything.
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
