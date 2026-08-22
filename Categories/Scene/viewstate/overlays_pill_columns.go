package viewstate

import (
	"github.com/dtauraso/wirefold/Categories/Chrome/Pills"
)

func (ui *UIState) writeOverlaysPillValues(lay Pills.Layout) {
	w := ui.overlaysPillValues
	if w == nil {
		return
	}
	w.Begin()

	w.F32("scrollY", lay.Scroll)
	w.Rect("pillX", "pillY", "pillW", "pillH", lay.Pill)
	w.Bool("open", lay.Open)
	w.Bool("active", lay.Active)
	w.Rect("popoverX", "popoverY", "popoverW", "popoverH", lay.Popover)
	w.Text("labelText", Pills.Label)

	for _, r := range lay.Rows {
		w.U8("rowKind", uint8(r.Kind))
		w.U8("rowDepth", uint8(r.Depth))
		w.Rect("rowX", "rowY", "rowW", "rowH", r.Rect)

		text := r.Label
		icon := r.Icon
		if r.Kind == Pills.RowHeading {
			text = r.Heading
			icon = disclosureGlyph(r.Open)
		}
		w.Str("rowTextData", "rowTextLen", text)
		w.Str("rowIconData", "rowIconLen", icon)

		w.Bool("rowOn", r.On)
		w.Bool("rowDisabled", r.Disabled)
		w.U32("rowCountOn", uint32(r.CountOn))
		w.U32("rowCountAll", uint32(r.CountAll))
		w.Rect("countX", "countY", "countW", "countH", r.Count)
	}

	if err := w.Flush(); err != nil {
		LogPersistErr("overlays_pill_values", "", err)
	}
}

func (ui *UIState) writeFitChipValues(r Pills.Rect) {
	w := ui.fitChipValues
	if w == nil {
		return
	}
	w.Begin()
	w.Rect("x", "y", "w", "h", r)
	w.Text("labelText", FitLabel)
	if err := w.Flush(); err != nil {
		LogPersistErr("fit_chip_values", "", err)
	}
}

func disclosureGlyph(open bool) string {
	if open {
		return "▼"
	}
	return "▶"
}
