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
		w.Str("stepNameText", "stepNameLen", s.Name)
		w.Str("stepShownText", "stepShownLen", s.Shown)
		w.Rect("stepUpX", "stepUpY", "stepUpW", "stepUpH", s.Up)
		w.Rect("stepDownX", "stepDownY", "stepDownW", "stepDownH", s.Down)
		w.Bool("stepUpOn", s.UpEnabled)
		w.Bool("stepDownOn", s.DownEnabled)
	}

	if lay.Open {
		addStep(lay.Lattice)
	}

	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "angle_pill_values: %v\n", err)
	}
}
