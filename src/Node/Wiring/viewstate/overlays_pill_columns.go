package viewstate

import (
	"github.com/dtauraso/wirefold/src/Node/Wiring/overlayspanel"
	B "github.com/dtauraso/wirefold/src/schema/buffer-layout"
)

func (ui *UIState) writeOverlaysPillColumns(lay overlayspanel.Layout) {
	c := ui.singletonCols
	if c == nil {
		return
	}

	c.SetF32(B.ColStreamOverlaysPillPillX, lay.Pill.X)
	c.SetF32(B.ColStreamOverlaysPillPillY, lay.Pill.Y)
	c.SetF32(B.ColStreamOverlaysPillPillW, lay.Pill.W)
	c.SetF32(B.ColStreamOverlaysPillPillH, lay.Pill.H)
	c.SetF32(B.ColStreamOverlaysPillScrollY, lay.Scroll)
	c.SetU8(B.ColStreamOverlaysPillOpen, boolU8(lay.Open))
	c.SetU8(B.ColStreamOverlaysPillActive, boolU8(lay.Active))
	c.SetF32(B.ColStreamOverlaysPillPopoverX, lay.Popover.X)
	c.SetF32(B.ColStreamOverlaysPillPopoverY, lay.Popover.Y)
	c.SetF32(B.ColStreamOverlaysPillPopoverW, lay.Popover.W)
	c.SetF32(B.ColStreamOverlaysPillPopoverH, lay.Popover.H)
	c.SetBytes(B.ColStreamOverlaysPillLabelText, []byte(overlayspanel.Label))

	rows := newRunCols()
	for _, r := range lay.Rows {
		rows.U8(B.ColStreamOverlaysPillRowKind, uint8(r.Kind))
		rows.U8(B.ColStreamOverlaysPillRowDepth, uint8(r.Depth))
		rows.Rect(B.ColStreamOverlaysPillRowX, B.ColStreamOverlaysPillRowY, B.ColStreamOverlaysPillRowW, B.ColStreamOverlaysPillRowH, r.Rect)

		text := r.Label
		icon := r.Icon
		if r.Kind == overlayspanel.RowHeading {
			text = r.Heading
			icon = disclosureGlyph(r.Open)
		}
		rows.Str(B.ColStreamOverlaysPillRowTextData, B.ColStreamOverlaysPillRowTextLen, text)
		rows.Str(B.ColStreamOverlaysPillRowIconData, B.ColStreamOverlaysPillRowIconLen, icon)

		rows.U8(B.ColStreamOverlaysPillRowOn, boolU8(r.On))
		rows.U8(B.ColStreamOverlaysPillRowDisabled, boolU8(r.Disabled))
		rows.U32(B.ColStreamOverlaysPillRowCountOn, uint32(r.CountOn))
		rows.U32(B.ColStreamOverlaysPillRowCountAll, uint32(r.CountAll))
		rows.Rect(B.ColStreamOverlaysPillCountX, B.ColStreamOverlaysPillCountY, B.ColStreamOverlaysPillCountW, B.ColStreamOverlaysPillCountH, r.Count)
	}
	rows.writeTo(c,
		B.ColStreamOverlaysPillRowKind, B.ColStreamOverlaysPillRowDepth,
		B.ColStreamOverlaysPillRowX, B.ColStreamOverlaysPillRowY, B.ColStreamOverlaysPillRowW, B.ColStreamOverlaysPillRowH,
		B.ColStreamOverlaysPillRowTextData, B.ColStreamOverlaysPillRowTextLen,
		B.ColStreamOverlaysPillRowIconData, B.ColStreamOverlaysPillRowIconLen,
		B.ColStreamOverlaysPillRowOn, B.ColStreamOverlaysPillRowDisabled,
		B.ColStreamOverlaysPillRowCountOn, B.ColStreamOverlaysPillRowCountAll,
		B.ColStreamOverlaysPillCountX, B.ColStreamOverlaysPillCountY, B.ColStreamOverlaysPillCountW, B.ColStreamOverlaysPillCountH,
	)
}

func (ui *UIState) writeFitChipColumns(r overlayspanel.Rect) {
	c := ui.singletonCols
	if c == nil {
		return
	}
	c.SetF32(B.ColStreamFitChipX, r.X)
	c.SetF32(B.ColStreamFitChipY, r.Y)
	c.SetF32(B.ColStreamFitChipW, r.W)
	c.SetF32(B.ColStreamFitChipH, r.H)
	c.SetBytes(B.ColStreamFitChipLabelText, []byte(FitLabel))
}

func disclosureGlyph(open bool) string {
	if open {
		return "▼"
	}
	return "▶"
}
