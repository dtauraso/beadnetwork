package AngleDropdown

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/src/valuefile"
)

const ValueRelFile = "view/chrome/angle-pill.bin"

var PillValueNames = []string{
	"pillX", "pillY", "pillW", "pillH",
	"open",
	"popoverX", "popoverY", "popoverW", "popoverH",
	"labelText",
	"stepX", "stepY", "stepW", "stepH",
	"stepNameText", "stepNameLen",
	"stepShownText", "stepShownLen",
	"stepValueRow", "stepDenom",
	"stepUpX", "stepUpY", "stepUpW", "stepUpH",
	"stepDownX", "stepDownY", "stepDownW", "stepDownH",
	"stepUpOn", "stepDownOn",
	"groupX", "groupY", "groupW", "groupH",
	"groupOpen",
	"groupHeadText", "groupHeadLen",
}

func ValueRelPath() string { return ValueRelFile }

type ValueWriter struct {
	*valuefile.BlobWriter
}

func NewValueWriter(sceneRoot string) *ValueWriter {
	path := filepath.Join(sceneRoot, filepath.FromSlash(ValueRelFile))
	return &ValueWriter{BlobWriter: valuefile.NewBlobWriter(path, PillValueNames)}
}

func (w *ValueWriter) Rect(xName, yName, wName, hName string, r Rect) {
	w.F32(xName, r.X)
	w.F32(yName, r.Y)
	w.F32(wName, r.W)
	w.F32(hName, r.H)
}
