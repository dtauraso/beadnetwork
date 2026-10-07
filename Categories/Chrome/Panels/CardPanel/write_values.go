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

	if lay.Pill.W > 0 {
		w.Rect("pillX", "pillY", "pillW", "pillH", lay.Pill)
		w.Text("pillText", PillLabel)
		open := uint8(0)
		if lay.Open {
			open = 1
		}
		w.U8("open", open)
	}

	for _, p := range lay.Panels {
		w.Rect("boxX", "boxY", "boxW", "boxH", p.Box)
		w.Rect("headX", "headY", "headW", "headH", p.Head)
		w.Str("titleText", "titleLen", p.Title)
	}
	for _, fb := range lay.Fields {
		w.Rect("fieldX", "fieldY", "fieldW", "fieldH", fb.Rect)
		w.Str("keyText", "keyLen", fb.Key)
		w.Str("valueText", "valueLen", s.Shown(fb.Field))
		editing := uint8(0)
		if s.Edit.Active && s.Edit.Field == fb.Field {
			editing = 1
		}
		w.U8("fieldEditing", editing)
		if fb.Up.W > 0 {
			w.Rect("upX", "upY", "upW", "upH", fb.Up)
			w.Rect("downX", "downY", "downW", "downH", fb.Down)
		}
	}
	w.Text("draftText", s.Edit.Draft)

	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "card_panel_values: %v\n", err)
	}
}
