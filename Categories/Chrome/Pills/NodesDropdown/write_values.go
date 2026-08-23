package NodesDropdown

import (
	"fmt"
	"os"

	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/Panel"
)

const refusedNotice = "edit refused — see the output channel"

func WriteValues(w *ValueWriter, lay Layout, refusedCount uint32) {
	if w == nil {
		return
	}
	w.Begin()

	w.Rect("pillX", "pillY", "pillW", "pillH", lay.Pill)
	w.Bool("open", lay.Open)
	w.Rect("popoverX", "popoverY", "popoverW", "popoverH", lay.Popover)
	w.Text("labelText", Label)

	for _, r := range lay.Rows {
		w.Rect("rowX", "rowY", "rowW", "rowH", r.Head)
		w.Bool("rowOpen", r.Open)
		w.Str("rowKindText", "rowKindLen", r.Kind)
		w.Str("rowFillText", "rowFillLen", r.Fill)
		w.Str("rowStrokeText", "rowStrokeLen", r.Stroke)
		w.Rect("swatchX", "swatchY", "swatchW", "swatchH", r.Swatch)
		desc := ""
		if r.Open {
			desc = r.Desc
		}
		w.Str("rowDescText", "rowDescLen", desc)
		w.F32("descX", r.DescRect.X)
		w.F32("descY", r.DescRect.Y)
		w.F32("descW", r.DescRect.W)
	}

	noticeW := Panel.TextWidth(refusedNotice, Panel.PillFontPx) + 16
	noticeH := Panel.LineHeight(Panel.PillFontPx) + 8
	noticeY := lay.Pill.Y + lay.Pill.H + Panel.PillGap
	if lay.Open {
		noticeY = lay.Popover.Y + lay.Popover.H + Panel.PillGap
	}
	w.U32("refusedCount", refusedCount)
	w.F32("refusedX", lay.Pill.X+lay.Pill.W-noticeW)
	w.F32("refusedY", noticeY)
	w.F32("refusedW", noticeW)
	w.F32("refusedH", noticeH)
	w.Text("refusedText", refusedNotice)

	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "nodes_pill_values: %v\n", err)
	}
}
