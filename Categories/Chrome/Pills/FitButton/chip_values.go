package FitButton

import (
	"path/filepath"

	Pills "github.com/dtauraso/wirefold/Categories/Chrome/Pills"
)

const ValueRelFile = "view/chrome/fit-chip.bin"

var ChipValueNames = []string{
	"x", "y", "w", "h",
	"labelText",
}

func ValueRelPath() string { return ValueRelFile }

type ValueWriter struct {
	*BlobWriter
}

func NewValueWriter(sceneRoot string) *ValueWriter {
	path := filepath.Join(sceneRoot, filepath.FromSlash(ValueRelFile))
	return &ValueWriter{BlobWriter: NewBlobWriter(path, ChipValueNames)}
}

func (w *ValueWriter) Rect(xName, yName, wName, hName string, r Pills.Rect) {
	w.F32(xName, r.X)
	w.F32(yName, r.Y)
	w.F32(wName, r.W)
	w.F32(hName, r.H)
}

// State is this piece's own: it carries the writer for its block, armed when
// the scene opens, so nothing outside can write it.
type State struct {
	w *ValueWriter
}

func (s *State) Arm(sceneRoot string) { s.w = NewValueWriter(sceneRoot) }
