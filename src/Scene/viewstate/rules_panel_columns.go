package viewstate

import (
	"github.com/dtauraso/wirefold/src/Chrome/Panels/PolarRulesPanel"
	B "github.com/dtauraso/wirefold/src/Buffer"
)

func (ui *UIState) writeRulesPanelColumns(lay PolarRulesPanel.Layout) {
	c := ui.singletonCols
	if c == nil {
		return
	}

	c.SetF32(B.ColStreamRulesPanelClipY, lay.RowsClip.Y)
	c.SetF32(B.ColStreamRulesPanelClipH, lay.RowsClip.H)
	c.SetF32(B.ColStreamRulesPanelScrollY, lay.Scroll)
	c.SetF32(B.ColStreamRulesPanelBoxX, lay.Box.X)
	c.SetF32(B.ColStreamRulesPanelBoxY, lay.Box.Y)
	c.SetF32(B.ColStreamRulesPanelBoxW, lay.Box.W)
	c.SetF32(B.ColStreamRulesPanelBoxH, lay.Box.H)
	c.SetU8(B.ColStreamRulesPanelOpen, boolU8(lay.Open))

	c.SetF32(B.ColStreamRulesPanelToggleX, lay.Toggle.X)
	c.SetF32(B.ColStreamRulesPanelToggleY, lay.Toggle.Y)
	c.SetF32(B.ColStreamRulesPanelToggleH, lay.Toggle.H)
	toggle := PolarRulesPanel.LabelClosed
	if lay.Open {
		toggle = PolarRulesPanel.LabelOpen
	}
	c.SetBytes(B.ColStreamRulesPanelToggleText, []byte(toggle))

	rows := newRunCols()
	for _, r := range lay.Rows {
		rows.U8(B.ColStreamRulesPanelRowKind, uint8(r.Kind))
		rows.Rect(B.ColStreamRulesPanelRowX, B.ColStreamRulesPanelRowY, B.ColStreamRulesPanelRowW, B.ColStreamRulesPanelRowH, r.Rect)
		rows.Str(B.ColStreamRulesPanelRowTextData, B.ColStreamRulesPanelRowTextLen, r.Text)
		rows.Str(B.ColStreamRulesPanelRowGlyphData, B.ColStreamRulesPanelRowGlyphLen, r.Glyph)
		rows.U8(B.ColStreamRulesPanelRowFree, boolU8(r.Free))
		rows.I32(B.ColStreamRulesPanelRowNodeRow, r.NodeRow)
		rows.I32(B.ColStreamRulesPanelRowEdgeRow, r.EdgeRow)
		rows.U8(B.ColStreamRulesPanelRowCheck, uint8(r.Check))
		rows.Rect(B.ColStreamRulesPanelRowCheckX, B.ColStreamRulesPanelRowCheckY, B.ColStreamRulesPanelRowCheckW, B.ColStreamRulesPanelRowCheckH, r.CheckRect)
		rows.U8(B.ColStreamRulesPanelRowValue, uint8(r.Value))
		rows.Point(B.ColStreamRulesPanelRowValueX, B.ColStreamRulesPanelRowValueY, r.ValueRect)
		rows.Rect(B.ColStreamRulesPanelRowSharedX, B.ColStreamRulesPanelRowSharedY, B.ColStreamRulesPanelRowSharedW, B.ColStreamRulesPanelRowSharedH, r.SharedRect)
		rows.U8(B.ColStreamRulesPanelRowEditing, boolU8(r.Editing))
	}
	rows.writeTo(c,
		B.ColStreamRulesPanelRowKind,
		B.ColStreamRulesPanelRowX, B.ColStreamRulesPanelRowY, B.ColStreamRulesPanelRowW, B.ColStreamRulesPanelRowH,
		B.ColStreamRulesPanelRowTextData, B.ColStreamRulesPanelRowTextLen,
		B.ColStreamRulesPanelRowGlyphData, B.ColStreamRulesPanelRowGlyphLen,
		B.ColStreamRulesPanelRowFree,
		B.ColStreamRulesPanelRowNodeRow, B.ColStreamRulesPanelRowEdgeRow,
		B.ColStreamRulesPanelRowCheck,
		B.ColStreamRulesPanelRowCheckX, B.ColStreamRulesPanelRowCheckY, B.ColStreamRulesPanelRowCheckW, B.ColStreamRulesPanelRowCheckH,
		B.ColStreamRulesPanelRowValue,
		B.ColStreamRulesPanelRowValueX, B.ColStreamRulesPanelRowValueY,
		B.ColStreamRulesPanelRowSharedX, B.ColStreamRulesPanelRowSharedY, B.ColStreamRulesPanelRowSharedW, B.ColStreamRulesPanelRowSharedH,
		B.ColStreamRulesPanelRowEditing,
	)

	c.SetBytes(B.ColStreamRulesPanelDraftText, []byte(lay.Draft))
	c.SetF32(B.ColStreamRulesPanelDraftX, lay.DraftRect.X)
	c.SetF32(B.ColStreamRulesPanelDraftY, lay.DraftRect.Y)
	c.SetF32(B.ColStreamRulesPanelDraftW, lay.DraftRect.W)
	c.SetF32(B.ColStreamRulesPanelDraftH, lay.DraftRect.H)

	c.SetU8(B.ColStreamRulesPanelMenuOpen, boolU8(lay.MenuOpen))
	c.SetI32(B.ColStreamRulesPanelMenuAnchorRow, lay.MenuAnchorRow)
	c.SetF32(B.ColStreamRulesPanelMenuX, lay.MenuBox.X)
	c.SetF32(B.ColStreamRulesPanelMenuY, lay.MenuBox.Y)
	c.SetF32(B.ColStreamRulesPanelMenuW, lay.MenuBox.W)
	c.SetF32(B.ColStreamRulesPanelMenuH, lay.MenuBox.H)

	menu := newRunCols()
	for _, m := range lay.MenuRows {
		menu.Rect(B.ColStreamRulesPanelMenuRowX, B.ColStreamRulesPanelMenuRowY, B.ColStreamRulesPanelMenuRowW, B.ColStreamRulesPanelMenuRowH, m.Rect)
		menu.Point(B.ColStreamRulesPanelMenuCheckX, B.ColStreamRulesPanelMenuCheckY, m.CheckRect)
		menu.Str(B.ColStreamRulesPanelMenuLabelData, B.ColStreamRulesPanelMenuLabelLen, m.Label)
		menu.I32(B.ColStreamRulesPanelMenuNodeRow, m.NodeRow)
	}
	menu.writeTo(c,
		B.ColStreamRulesPanelMenuRowX, B.ColStreamRulesPanelMenuRowY, B.ColStreamRulesPanelMenuRowW, B.ColStreamRulesPanelMenuRowH,
		B.ColStreamRulesPanelMenuCheckX, B.ColStreamRulesPanelMenuCheckY,
		B.ColStreamRulesPanelMenuLabelData, B.ColStreamRulesPanelMenuLabelLen,
		B.ColStreamRulesPanelMenuNodeRow,
	)
}
