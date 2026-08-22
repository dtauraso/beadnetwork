package TiltPanel

import (
	"path/filepath"
)

const ValueRelFile = "view/chrome/tilt-panel.bin"

var PanelValueNames = []string{
	"boxX", "boxY", "boxW", "boxH",
	"startX", "startY", "startW", "startH",
	"resetX", "resetY", "resetW", "resetH",
	"startText", "resetText",
	"colNodeRow", "colLabelText", "colLabelLen",
	"headX", "headY", "headW", "headH",
	"roundsX", "roundsY", "roundsW", "roundsH",
	"msgsX", "msgsY", "msgsW", "msgsH",
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
