package viewstate

import (
	"github.com/dtauraso/wirefold/src/Chrome/Pills/AngleDropdown"
	"github.com/dtauraso/wirefold/src/valuefile"
)

func (ui *UIState) writeAnglePillValues(lay AngleDropdown.Layout) {
	w := ui.anglePillValues
	if w == nil {
		return
	}
	w.Begin()

	w.Rect("pillX", "pillY", "pillW", "pillH", lay.Pill)
	w.Bool("open", lay.Open)
	w.Rect("popoverX", "popoverY", "popoverW", "popoverH", lay.Popover)
	w.Text("labelText", AngleDropdown.Label)

	addStep := func(s AngleDropdown.Stepper) {
		w.Rect("stepX", "stepY", "stepW", "stepH", s.Row)
		w.Str("stepNameText", "stepNameLen", s.Name)
		w.Str("stepShownText", "stepShownLen", s.Shown)
		w.I32("stepValueRow", s.ValueRow)
		w.I32("stepDenom", s.Denom)
		w.Rect("stepUpX", "stepUpY", "stepUpW", "stepUpH", s.Up)
		w.Rect("stepDownX", "stepDownY", "stepDownW", "stepDownH", s.Down)
		w.Bool("stepUpOn", s.UpEnabled)
		w.Bool("stepDownOn", s.DownEnabled)
	}

	if lay.Open {
		addStep(lay.Lattice)
		for _, g := range lay.Groups {
			w.Rect("groupX", "groupY", "groupW", "groupH", g.Head)
			w.Bool("groupOpen", g.Open)
			w.Str("groupHeadText", "groupHeadLen", g.Heading)
			if g.Open {
				addStep(g.Phi)
			}
		}
	}

	if err := w.Flush(); err != nil {
		valuefile.LogPersistErr("angle_pill_values", "", err)
	}
}
