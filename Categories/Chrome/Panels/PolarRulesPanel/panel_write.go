package PolarRulesPanel

import (
	"path/filepath"
)

type ValueWriter struct {
	*BlobWriter
}

func NewValueWriter(sceneRoot string) *ValueWriter {
	path := filepath.Join(sceneRoot, filepath.FromSlash(ValueRelFile))
	return &ValueWriter{BlobWriter: NewBlobWriter(path, PanelValueNames)}
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
