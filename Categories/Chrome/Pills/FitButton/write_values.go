package FitButton

import (
	"fmt"
	"os"

	Pills "github.com/dtauraso/beadnetwork/Categories/Chrome/Pills"
)

func (s *State) Write(r Pills.Rect) {
	w := s.w
	if w == nil {
		return
	}
	w.Begin()
	w.Rect("x", "y", "w", "h", r)
	w.Text("labelText", FitLabel)
	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "fit_chip_values: %v\n", err)
	}
}

const FitLabel = "⌂ fit"
