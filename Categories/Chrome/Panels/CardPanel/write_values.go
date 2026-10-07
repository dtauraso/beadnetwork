package CardPanel

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

	for _, p := range lay.Panels {
		w.Rect("boxX", "boxY", "boxW", "boxH", p.Box)
		w.Rect("headX", "headY", "headW", "headH", p.Head)
		w.Str("titleText", "titleLen", p.Title)
	}
	for _, fb := range lay.Fields {
		w.Rect("fieldX", "fieldY", "fieldW", "fieldH", fb.Rect)
		w.Str("keyText", "keyLen", fb.Key)
		w.I32("fieldValue", int32(s.Value(fb.Field)))
		editing := uint8(0)
		if s.Edit.Active && s.Edit.Field == fb.Field {
			editing = 1
		}
		w.U8("fieldEditing", editing)
	}
	w.Text("draftText", s.Edit.Draft)

	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "card_panel_values: %v\n", err)
	}
}
