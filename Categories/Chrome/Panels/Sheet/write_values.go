package Sheet

import (
	"fmt"
	"os"
)

func (s *State) Write(lay Layout) {
	w := s.w
	if w == nil {
		return
	}
	w.Begin()

	flag := func(b bool) uint8 {
		if b {
			return 1
		}
		return 0
	}
	for _, p := range lay.Pills {
		w.Rect("pillX", "pillY", "pillW", "pillH", p.Pill)
		w.Str("pillText", "pillLen", p.Label)
		w.U8("pillOpen", flag(p.Open))
		w.U8("pillBar", flag(p.Bar))
		w.Rect("boxX", "boxY", "boxW", "boxH", p.Popover)
		w.F32("slideX", p.Slide.X)
		w.F32("slideW", p.Slide.W)
	}
	for i, p := range lay.Pills {
		for _, t := range p.Tabs {
			w.U32("tabPill", uint32(i))
			w.Rect("tabX", "tabY", "tabW", "tabH", t.Rect)
			w.Str("tabText", "tabLen", t.Label)
			w.U8("tabOn", flag(t.On))
		}
	}
	for _, h := range lay.Heads {
		w.U32("headBox", uint32(h.Box))
		w.F32("headX", h.Rect.X)
		w.F32("headY", h.Rect.Y)
		w.F32("headH", h.Rect.H)
		w.Str("titleText", "titleLen", h.Title)
	}
	for _, cb := range lay.Cells {
		w.U32("cellBox", uint32(cb.Box))
		w.Rect("cellX", "cellY", "cellW", "cellH", cb.Rect)
		w.U8("cellKind", uint8(cb.Cell.Kind))
		w.U8("cellLit", flag(cb.Cell.Lit))
		w.U8("cellEditing", flag(s.Edit.Active && s.Edit.Act == cb.Cell.Act && cb.Cell.Typed))
		w.Str("keyText", "keyLen", cb.Cell.Key)
		w.Str("valueText", "valueLen", cb.Cell.Value)
		if cb.Up.W > 0 {
			w.U32("upBox", uint32(cb.Box))
			w.Rect("upX", "upY", "upW", "upH", cb.Up)
			w.Rect("downX", "downY", "downW", "downH", cb.Down)
		}
	}
	w.Text("draftText", s.Edit.Draft)
	for _, cb := range lay.Cells {
		w.U8("cellOutlined", flag(cb.Cell.Outlined))
		w.U8("cellBoxed", flag(cb.Cell.Boxed))
		w.U8("cellSlides", flag(cb.Slides))
		w.U8("cellDim", flag(cb.Cell.Dim))
		if cb.Up.W > 0 {
			w.U8("upSlides", flag(cb.Slides))
			w.U8("upLit", flag(cb.Cell.Locked && cb.Cell.LockSign > 0))
			w.U8("downLit", flag(cb.Cell.Locked && cb.Cell.LockSign < 0))
			w.U8("upDim", flag(cb.Cell.Dim))
		}
	}

	for _, c := range lay.Cards {
		w.U32("cardBox", uint32(c.Box))
		w.Rect("cardX", "cardY", "cardW", "cardH", c.Rect)
	}

	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "sheet_values: %v\n", err)
	}
}
