package NodesDropdown

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/src/valuefile"
)

const ValueRelDir = "view/chrome/nodes-pill"

var PillValueNames = []string{
	"pillX", "pillY", "pillW", "pillH",
	"open",
	"popoverX", "popoverY", "popoverW", "popoverH",
	"labelText",
	"rowX", "rowY", "rowW", "rowH",
	"rowOpen",
	"rowKindText", "rowKindLen",
	"rowFillText", "rowFillLen",
	"rowStrokeText", "rowStrokeLen",
	"swatchX", "swatchY", "swatchW", "swatchH",
	"rowDescText", "rowDescLen",
	"descX", "descY", "descW", "descH",
	"refusedCount", "refusedX", "refusedY", "refusedW", "refusedH", "refusedText",
}

func ValueRelPath(name string) string {
	return ValueRelDir + "/" + valuefile.BlobFileName(name)
}

type ValueWriter struct {
	*valuefile.BlobWriter
}

func NewValueWriter(sceneRoot string) *ValueWriter {
	dir := filepath.Join(sceneRoot, filepath.FromSlash(ValueRelDir))
	return &ValueWriter{BlobWriter: valuefile.NewBlobWriter(dir, PillValueNames)}
}

func (w *ValueWriter) Rect(xName, yName, wName, hName string, r Rect) {
	w.F32(xName, r.X)
	w.F32(yName, r.Y)
	w.F32(wName, r.W)
	w.F32(hName, r.H)
}
