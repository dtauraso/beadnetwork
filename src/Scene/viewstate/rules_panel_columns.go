package viewstate

import (
	"github.com/dtauraso/wirefold/src/Chrome/Panels/PolarRulesPanel"
	"github.com/dtauraso/wirefold/src/valuefile"
)

func (ui *UIState) writeRulesPanelValues(lay PolarRulesPanel.Layout) {
	w := ui.rulesValues
	if w == nil {
		return
	}
	w.Begin()

	w.F32("clipY", lay.RowsClip.Y)
	w.F32("clipH", lay.RowsClip.H)
	w.F32("scrollY", lay.Scroll)
	w.Rect("boxX", "boxY", "boxW", "boxH", lay.Box)
	w.Bool("open", lay.Open)

	w.F32("toggleX", lay.Toggle.X)
	w.F32("toggleY", lay.Toggle.Y)
	w.F32("toggleH", lay.Toggle.H)
	toggle := PolarRulesPanel.LabelClosed
	if lay.Open {
		toggle = PolarRulesPanel.LabelOpen
	}
	w.Text("toggleText", toggle)

	for _, r := range lay.Rows {
		w.U8("rowKind", uint8(r.Kind))
		w.Rect("rowX", "rowY", "rowW", "rowH", r.Rect)
		w.Str("rowTextData", "rowTextLen", r.Text)
		w.Str("rowGlyphData", "rowGlyphLen", r.Glyph)
		w.Bool("rowFree", r.Free)
		w.I32("rowNodeRow", r.NodeRow)
		w.I32("rowEdgeRow", r.EdgeRow)
		w.U8("rowCheck", uint8(r.Check))
		w.Rect("rowCheckX", "rowCheckY", "rowCheckW", "rowCheckH", r.CheckRect)
		w.U8("rowValue", uint8(r.Value))
		w.Point("rowValueX", "rowValueY", r.ValueRect)
		w.Rect("rowSharedX", "rowSharedY", "rowSharedW", "rowSharedH", r.SharedRect)
		w.Bool("rowEditing", r.Editing)
	}

	w.Text("draftText", lay.Draft)
	w.Rect("draftX", "draftY", "draftW", "draftH", lay.DraftRect)

	w.Bool("menuOpen", lay.MenuOpen)
	w.I32("menuAnchorRow", lay.MenuAnchorRow)
	w.Rect("menuX", "menuY", "menuW", "menuH", lay.MenuBox)

	for _, m := range lay.MenuRows {
		w.Rect("menuRowX", "menuRowY", "menuRowW", "menuRowH", m.Rect)
		w.Point("menuCheckX", "menuCheckY", m.CheckRect)
		w.Str("menuLabelData", "menuLabelLen", m.Label)
		w.I32("menuNodeRow", m.NodeRow)
	}

	if err := w.Flush(); err != nil {
		valuefile.LogPersistErr("rules_panel_values", "", err)
	}
}
