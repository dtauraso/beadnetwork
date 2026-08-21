package viewstate

import (
	"github.com/dtauraso/wirefold/src/Chrome/Tabs"
	"github.com/dtauraso/wirefold/src/valuefile"
)

func (ui *UIState) writeTabStripValues(lay Tabs.Layout) {
	w := ui.tabStripValues
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
		valuefile.LogPersistErr("tab_strip_values", "", err)
	}
}
