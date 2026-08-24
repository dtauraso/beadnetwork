package AngleDropdown

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

	w.Rect("pillX", "pillY", "pillW", "pillH", lay.Pill)
	w.Bool("open", lay.Open)
	w.Rect("popoverX", "popoverY", "popoverW", "popoverH", lay.Popover)
	w.Text("labelText", Label)

	addStep := func(s Stepper) {
		w.F32("stepX", s.Row.X)
		w.F32("stepY", s.Row.Y)
		w.F32("stepH", s.Row.H)
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
			w.F32("groupX", g.Head.X)
			w.F32("groupY", g.Head.Y)
			w.F32("groupH", g.Head.H)
			w.Bool("groupOpen", g.Open)
			w.Str("groupHeadText", "groupHeadLen", g.Heading)
			if g.Open {
				addStep(g.Phi)
			}
		}
	}

	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "angle_pill_values: %v\n", err)
	}
}
