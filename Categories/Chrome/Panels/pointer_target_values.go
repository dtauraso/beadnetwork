package Panels

import (
	"path/filepath"
)

const ValueRelPath = "view/pointer-target.bin"

var PointerTargetValueNames = []string{
	"x", "y", "w", "h", "kind", "tipX", "tipY", "tipText",
}

type ValueWriter struct {
	*BlobWriter
}

func NewValueWriter(sceneRoot string) *ValueWriter {
	path := filepath.Join(sceneRoot, filepath.FromSlash(ValueRelPath))
	return &ValueWriter{BlobWriter: NewBlobWriter(path, PointerTargetValueNames)}
}

func (w *ValueWriter) Write(x, y, wd, h float32, kind uint8, tipX, tipY float32, tip string) error {
	w.Begin()
	w.F32("x", x)
	w.F32("y", y)
	w.F32("w", wd)
	w.F32("h", h)
	w.U8("kind", kind)
	w.F32("tipX", tipX)
	w.F32("tipY", tipY)
	w.Text("tipText", tip)
	return w.Flush()
}

// State is this piece's own: it carries the writer for its block, armed when
// the scene opens, so nothing outside can write it.
type State struct {
	w *ValueWriter
}

func (s *State) Arm(sceneRoot string) { s.w = NewValueWriter(sceneRoot) }

// W is this piece's writer, for the view's own column writing.
func (s *State) W() *ValueWriter { return s.w }
