package CardPanel

import (
	"path/filepath"
)

const ValueRelFile = "view/chrome/card-panel.bin"

var PanelValueNames = []string{
	"boxX", "boxY", "boxW", "boxH",
	"headX", "headY", "headW", "headH",
	"titleText", "titleLen",
	"fieldX", "fieldY", "fieldW", "fieldH",
	"keyText", "keyLen",
	"fieldValue", "fieldEditing",
	"draftText",
}

func ValueRelPath() string { return ValueRelFile }

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
