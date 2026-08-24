package TiltPanel

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

	w.Rect("boxX", "boxY", "boxW", "boxH", lay.Box)
	w.Rect("startX", "startY", "startW", "startH", lay.Start)
	w.Rect("resetX", "resetY", "resetW", "resetH", lay.Reset)
	w.Text("startText", StartLabel)
	w.Text("resetText", ResetLabel)

	for _, col := range lay.Columns {
		w.I32("colNodeRow", col.NodeRow)
		w.Str("colLabelText", "colLabelLen", col.Label)
		w.Rect("headX", "headY", "headW", "headH", col.Head)
		w.Rect("roundsX", "roundsY", "roundsW", "roundsH", col.Rounds)
		w.Rect("msgsX", "msgsY", "msgsW", "msgsH", col.Msgs)
	}

	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "tilt_panel_values: %v\n", err)
	}
}
