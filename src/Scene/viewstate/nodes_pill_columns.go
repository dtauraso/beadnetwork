package viewstate

import (
	"github.com/dtauraso/wirefold/src/Chrome/Pills/NodesDropdown"
	"github.com/dtauraso/wirefold/src/Chrome/Panels/Panel"
	B "github.com/dtauraso/wirefold/src/Buffer"
)

const refusedNotice = "edit refused — see the output channel"

func (ui *UIState) writeNodesPillColumns(lay NodesDropdown.Layout) {
	c := ui.singletonCols
	if c == nil {
		return
	}

	c.SetF32(B.ColStreamNodesPillPillX, lay.Pill.X)
	c.SetF32(B.ColStreamNodesPillPillY, lay.Pill.Y)
	c.SetF32(B.ColStreamNodesPillPillW, lay.Pill.W)
	c.SetF32(B.ColStreamNodesPillPillH, lay.Pill.H)
	c.SetU8(B.ColStreamNodesPillOpen, boolU8(lay.Open))
	c.SetF32(B.ColStreamNodesPillPopoverX, lay.Popover.X)
	c.SetF32(B.ColStreamNodesPillPopoverY, lay.Popover.Y)
	c.SetF32(B.ColStreamNodesPillPopoverW, lay.Popover.W)
	c.SetF32(B.ColStreamNodesPillPopoverH, lay.Popover.H)
	c.SetBytes(B.ColStreamNodesPillLabelText, []byte(NodesDropdown.Label))

	rows := newRunCols()
	for _, r := range lay.Rows {
		rows.Rect(B.ColStreamNodesPillRowX, B.ColStreamNodesPillRowY, B.ColStreamNodesPillRowW, B.ColStreamNodesPillRowH, r.Head)
		rows.U8(B.ColStreamNodesPillRowOpen, boolU8(r.Open))
		rows.Str(B.ColStreamNodesPillRowKindText, B.ColStreamNodesPillRowKindLen, r.Kind)
		rows.Str(B.ColStreamNodesPillRowFillText, B.ColStreamNodesPillRowFillLen, r.Fill)
		rows.Str(B.ColStreamNodesPillRowStrokeText, B.ColStreamNodesPillRowStrokeLen, r.Stroke)
		rows.Rect(B.ColStreamNodesPillSwatchX, B.ColStreamNodesPillSwatchY, B.ColStreamNodesPillSwatchW, B.ColStreamNodesPillSwatchH, r.Swatch)
		desc := ""
		if r.Open {
			desc = r.Desc
		}
		rows.Str(B.ColStreamNodesPillRowDescText, B.ColStreamNodesPillRowDescLen, desc)
		rows.Rect(B.ColStreamNodesPillDescX, B.ColStreamNodesPillDescY, B.ColStreamNodesPillDescW, B.ColStreamNodesPillDescH, r.DescRect)
	}
	rows.writeTo(c,
		B.ColStreamNodesPillRowX, B.ColStreamNodesPillRowY, B.ColStreamNodesPillRowW, B.ColStreamNodesPillRowH,
		B.ColStreamNodesPillRowOpen,
		B.ColStreamNodesPillRowKindText, B.ColStreamNodesPillRowKindLen,
		B.ColStreamNodesPillRowFillText, B.ColStreamNodesPillRowFillLen,
		B.ColStreamNodesPillRowStrokeText, B.ColStreamNodesPillRowStrokeLen,
		B.ColStreamNodesPillSwatchX, B.ColStreamNodesPillSwatchY, B.ColStreamNodesPillSwatchW, B.ColStreamNodesPillSwatchH,
		B.ColStreamNodesPillRowDescText, B.ColStreamNodesPillRowDescLen,
		B.ColStreamNodesPillDescX, B.ColStreamNodesPillDescY, B.ColStreamNodesPillDescW, B.ColStreamNodesPillDescH,
	)

	noticeW := Panel.TextWidth(refusedNotice, Panel.PillFontPx) + 16
	noticeH := Panel.LineHeight(Panel.PillFontPx) + 8
	noticeY := lay.Pill.Y + lay.Pill.H + Panel.PillGap
	if lay.Open {
		noticeY = lay.Popover.Y + lay.Popover.H + Panel.PillGap
	}
	c.SetU32(B.ColStreamNodesPillRefusedCount, ui.EditRefused)
	c.SetF32(B.ColStreamNodesPillRefusedX, lay.Pill.X+lay.Pill.W-noticeW)
	c.SetF32(B.ColStreamNodesPillRefusedY, noticeY)
	c.SetF32(B.ColStreamNodesPillRefusedW, noticeW)
	c.SetF32(B.ColStreamNodesPillRefusedH, noticeH)
	c.SetBytes(B.ColStreamNodesPillRefusedText, []byte(refusedNotice))
}
