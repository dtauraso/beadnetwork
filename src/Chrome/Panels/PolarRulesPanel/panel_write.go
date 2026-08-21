package PolarRulesPanel

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/src/valuefile"
)

type ValueWriter struct {
	*valuefile.BlobWriter
}

func NewValueWriter(sceneRoot string) *ValueWriter {
	dir := filepath.Join(sceneRoot, filepath.FromSlash(ValueRelDir))
	return &ValueWriter{BlobWriter: valuefile.NewBlobWriter(dir, PanelValueNames)}
}

func (w *ValueWriter) Rect(xName, yName, wName, hName string, r Rect) {
	w.F32(xName, r.X)
	w.F32(yName, r.Y)
	w.F32(wName, r.W)
	w.F32(hName, r.H)
}

func (w *ValueWriter) Point(xName, yName string, r Rect) {
	w.F32(xName, r.X)
	w.F32(yName, r.Y)
}
