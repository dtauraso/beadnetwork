package FitButton

import (
	"fmt"
	"os"

	Pills "github.com/dtauraso/wirefold/Categories/Chrome/Pills"
)

func WriteValues(w *ValueWriter, r Pills.Rect) {
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
