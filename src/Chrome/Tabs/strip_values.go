package Tabs

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/src/valuefile"
)

const ValueRelDir = "view/chrome/tab-strip"

var StripValueNames = []string{
	"stripX", "stripY", "stripW", "stripH",
	"tabX", "tabY", "tabW", "tabH",
	"tabNameText", "tabNameLen",
	"tabSelected",
}

func ValueRelPath(name string) string {
	return ValueRelDir + "/" + valuefile.BlobFileName(name)
}

type ValueWriter struct {
	*valuefile.BlobWriter
}

func NewValueWriter(sceneRoot string) *ValueWriter {
	dir := filepath.Join(sceneRoot, filepath.FromSlash(ValueRelDir))
	return &ValueWriter{BlobWriter: valuefile.NewBlobWriter(dir, StripValueNames)}
}

func (w *ValueWriter) Rect(xName, yName, wName, hName string, r Rect) {
	w.F32(xName, r.X)
	w.F32(yName, r.Y)
	w.F32(wName, r.W)
	w.F32(hName, r.H)
}
