package TiltPanel

import (
	"fmt"
	"os"
)

// WriteValues writes this piece's own block, from its own writer and its own
// layout — not from the view state, which would let it write anything.
func WriteValues(w *ValueWriter, lay Layout) {
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
