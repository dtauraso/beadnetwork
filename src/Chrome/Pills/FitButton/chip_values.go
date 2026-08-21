package FitButton

import (
	"path/filepath"

	Pills "github.com/dtauraso/wirefold/src/Chrome/Pills"
	"github.com/dtauraso/wirefold/src/valuefile"
)

const ValueRelDir = "view/chrome/fit-chip"

var ChipValueNames = []string{
	"x", "y", "w", "h",
	"labelText",
}

func ValueRelPath(name string) string {
	return ValueRelDir + "/" + valuefile.BlobFileName(name)
}

type ValueWriter struct {
	*valuefile.BlobWriter
}

func NewValueWriter(sceneRoot string) *ValueWriter {
	dir := filepath.Join(sceneRoot, filepath.FromSlash(ValueRelDir))
	return &ValueWriter{BlobWriter: valuefile.NewBlobWriter(dir, ChipValueNames)}
}

func (w *ValueWriter) Rect(xName, yName, wName, hName string, r Pills.Rect) {
	w.F32(xName, r.X)
	w.F32(yName, r.Y)
	w.F32(wName, r.W)
	w.F32(hName, r.H)
}
