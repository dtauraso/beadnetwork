package viewstate

import (
	"github.com/dtauraso/wirefold/Chrome/Panels/TiltPanel"
)

func (ui *UIState) writeTiltPanelValues(lay TiltPanel.Layout) {
	w := ui.tiltPanelValues
	if w == nil {
		return
	}
	w.Begin()

	w.Rect("boxX", "boxY", "boxW", "boxH", lay.Box)
	w.Rect("startX", "startY", "startW", "startH", lay.Start)
	w.Rect("resetX", "resetY", "resetW", "resetH", lay.Reset)
	w.Text("startText", TiltPanel.StartLabel)
	w.Text("resetText", TiltPanel.ResetLabel)

	for _, col := range lay.Columns {
		w.I32("colNodeRow", col.NodeRow)
		w.Str("colLabelText", "colLabelLen", col.Label)
		w.Rect("headX", "headY", "headW", "headH", col.Head)
		w.Rect("roundsX", "roundsY", "roundsW", "roundsH", col.Rounds)
		w.Rect("msgsX", "msgsY", "msgsW", "msgsH", col.Msgs)
	}

	if err := w.Flush(); err != nil {
		LogPersistErr("tilt_panel_values", "", err)
	}
}
