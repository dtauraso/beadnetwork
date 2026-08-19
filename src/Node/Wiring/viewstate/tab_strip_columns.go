package viewstate

import (
	"github.com/dtauraso/wirefold/src/Node/Wiring/tabstrip"
	B "github.com/dtauraso/wirefold/src/Buffer"
)

func (ui *UIState) writeTabStripColumns(lay tabstrip.Layout) {
	c := ui.singletonCols
	if c == nil {
		return
	}

	c.SetF32(B.ColStreamTabStripStripX, lay.Strip.X)
	c.SetF32(B.ColStreamTabStripStripY, lay.Strip.Y)
	c.SetF32(B.ColStreamTabStripStripW, lay.Strip.W)
	c.SetF32(B.ColStreamTabStripStripH, lay.Strip.H)

	tabs := newRunCols()
	for _, t := range lay.Tabs {
		tabs.Rect(B.ColStreamTabStripTabX, B.ColStreamTabStripTabY, B.ColStreamTabStripTabW, B.ColStreamTabStripTabH, t.Rect)
		tabs.Str(B.ColStreamTabStripTabNameText, B.ColStreamTabStripTabNameLen, t.Name)
		tabs.U8(B.ColStreamTabStripTabSelected, boolU8(t.Selected))
	}
	tabs.writeTo(c,
		B.ColStreamTabStripTabX, B.ColStreamTabStripTabY, B.ColStreamTabStripTabW, B.ColStreamTabStripTabH,
		B.ColStreamTabStripTabNameText, B.ColStreamTabStripTabNameLen,
		B.ColStreamTabStripTabSelected,
	)
}
