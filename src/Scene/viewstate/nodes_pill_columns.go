package viewstate

import (
	"github.com/dtauraso/wirefold/src/Chrome/Panels/Panel"
	"github.com/dtauraso/wirefold/src/Chrome/Pills/NodesDropdown"
	"github.com/dtauraso/wirefold/src/valuefile"
)

const refusedNotice = "edit refused — see the output channel"

func (ui *UIState) writeNodesPillValues(lay NodesDropdown.Layout) {
	w := ui.nodesPillValues
	if w == nil {
		return
	}
	w.Begin()

	w.Rect("pillX", "pillY", "pillW", "pillH", lay.Pill)
	w.Bool("open", lay.Open)
	w.Rect("popoverX", "popoverY", "popoverW", "popoverH", lay.Popover)
	w.Text("labelText", NodesDropdown.Label)

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
		w.Rect("descX", "descY", "descW", "descH", r.DescRect)
	}

	noticeW := Panel.TextWidth(refusedNotice, Panel.PillFontPx) + 16
	noticeH := Panel.LineHeight(Panel.PillFontPx) + 8
	noticeY := lay.Pill.Y + lay.Pill.H + Panel.PillGap
	if lay.Open {
		noticeY = lay.Popover.Y + lay.Popover.H + Panel.PillGap
	}
	w.U32("refusedCount", ui.EditRefused)
	w.F32("refusedX", lay.Pill.X+lay.Pill.W-noticeW)
	w.F32("refusedY", noticeY)
	w.F32("refusedW", noticeW)
	w.F32("refusedH", noticeH)
	w.Text("refusedText", refusedNotice)

	if err := w.Flush(); err != nil {
		valuefile.LogPersistErr("nodes_pill_values", "", err)
	}
}
