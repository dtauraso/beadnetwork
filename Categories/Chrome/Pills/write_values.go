package Pills

import (
	"fmt"
	"os"
)

// WriteValues writes the overlays pill's own block, from its own writer and
// its own layout.
func WriteValues(w *ValueWriter, lay Layout) {
	if w == nil {
		return
	}
	w.Begin()

	w.F32("scrollY", lay.Scroll)
	w.Rect("pillX", "pillY", "pillW", "pillH", lay.Pill)
	w.Bool("open", lay.Open)
	w.Bool("active", lay.Active)
	w.Rect("popoverX", "popoverY", "popoverW", "popoverH", lay.Popover)
	w.Text("labelText", Label)

	for _, r := range lay.Rows {
		w.U8("rowKind", uint8(r.Kind))
		w.U8("rowDepth", uint8(r.Depth))
		w.Rect("rowX", "rowY", "rowW", "rowH", r.Rect)

		text := r.Label
		icon := r.Icon
		if r.Kind == RowHeading {
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
		fmt.Fprintf(os.Stderr, "overlays_pill_values: %v\n", err)
	}
}

func disclosureGlyph(open bool) string {
	if open {
		return "▼"
	}
	return "▶"
}
