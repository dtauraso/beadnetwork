package NodesDropdown

import (
	"path/filepath"
)

const ValueRelFile = "view/chrome/nodes-pill.bin"

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
	"descX", "descY", "descW",
	"refusedCount", "refusedX", "refusedY", "refusedW", "refusedH", "refusedText",
}

func ValueRelPath() string { return ValueRelFile }

type ValueWriter struct {
	*BlobWriter
}

func NewValueWriter(sceneRoot string) *ValueWriter {
	path := filepath.Join(sceneRoot, filepath.FromSlash(ValueRelFile))
	return &ValueWriter{BlobWriter: NewBlobWriter(path, PillValueNames)}
}

func (w *ValueWriter) Rect(xName, yName, wName, hName string, r Rect) {
	w.F32(xName, r.X)
	w.F32(yName, r.Y)
	w.F32(wName, r.W)
	w.F32(hName, r.H)
}
