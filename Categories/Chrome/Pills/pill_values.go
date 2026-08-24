package Pills

import (
	"path/filepath"
)

const ValueRelFile = "view/chrome/overlays-pill.bin"

var PillValueNames = []string{
	"scrollY",
	"pillX", "pillY", "pillW", "pillH",
	"open", "active",
	"popoverX", "popoverY", "popoverW", "popoverH",
	"labelText",
	"rowKind", "rowDepth",
	"rowX", "rowY", "rowW", "rowH",
	"rowTextData", "rowTextLen",
	"rowIconData", "rowIconLen",
	"rowOn", "rowDisabled",
	"rowCountOn", "rowCountAll",
	"countX", "countY", "countW", "countH",
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

type State struct {
	w *ValueWriter // this piece's own writer, armed when the scene opens

	Scroll float32
}

func (s *State) Arm(sceneRoot string) { s.w = NewValueWriter(sceneRoot) }
