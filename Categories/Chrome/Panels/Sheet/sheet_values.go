package Sheet

import (
	"path/filepath"
)

const ValueRelFile = "view/chrome/sheets.bin"

var SheetValueNames = []string{
	"pillX", "pillY", "pillW", "pillH",
	"pillText", "pillLen", "pillOpen",
	"boxX", "boxY", "boxW", "boxH",
	"headBox", "headX", "headY", "headH",
	"titleText", "titleLen",
	"cellBox", "cellX", "cellY", "cellW", "cellH",
	"cellKind", "cellLit", "cellEditing",
	"keyText", "keyLen",
	"valueText", "valueLen",
	"upBox", "upX", "upY", "upW", "upH",
	"downX", "downY", "downW", "downH",
	"draftText",
	"cellOutlined", "cellBoxed",
	"slideX", "slideW", "cellSlides", "upSlides",
}

func ValueRelPath() string { return ValueRelFile }

type ValueWriter struct {
	*BlobWriter
}

func NewValueWriter(sceneRoot string) *ValueWriter {
	path := filepath.Join(sceneRoot, filepath.FromSlash(ValueRelFile))
	return &ValueWriter{BlobWriter: NewBlobWriter(path, SheetValueNames)}
}

func (w *ValueWriter) Rect(xName, yName, wName, hName string, r Rect) {
	w.F32(xName, r.X)
	w.F32(yName, r.Y)
	w.F32(wName, r.W)
	w.F32(hName, r.H)
}
