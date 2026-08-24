package SliderPanel

import (
	"path/filepath"
)

const ValueRelPath = "view/slider-panel.bin"

var PanelValueNames = []string{
	"boxX", "boxY", "boxW", "boxH",
	"rectX", "rectY", "rectW", "rectH",
	"selected",
	"numText", "numLen",
	"denText", "denLen",
	"trackX", "trackY", "trackW", "trackH",
}

type ValueWriter struct {
	*BlobWriter
}

func NewValueWriter(sceneRoot string) *ValueWriter {
	path := filepath.Join(sceneRoot, filepath.FromSlash(ValueRelPath))
	return &ValueWriter{BlobWriter: NewBlobWriter(path, PanelValueNames)}
}

func (w *ValueWriter) Rect(xName, yName, wName, hName string, r Rect) {
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
