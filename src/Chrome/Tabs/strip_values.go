package Tabs

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/src/valuefile"
)

const ValueRelFile = "view/chrome/tab-strip.bin"

var StripValueNames = []string{
	"stripX", "stripY", "stripW", "stripH",
	"tabX", "tabY", "tabW", "tabH",
	"tabNameText", "tabNameLen",
	"tabSelected",
}

func ValueRelPath() string { return ValueRelFile }

type ValueWriter struct {
	*valuefile.BlobWriter
}

func NewValueWriter(sceneRoot string) *ValueWriter {
	path := filepath.Join(sceneRoot, filepath.FromSlash(ValueRelFile))
	return &ValueWriter{BlobWriter: valuefile.NewBlobWriter(path, StripValueNames)}
}

func (w *ValueWriter) Rect(xName, yName, wName, hName string, r Rect) {
	w.F32(xName, r.X)
	w.F32(yName, r.Y)
	w.F32(wName, r.W)
	w.F32(hName, r.H)
}
