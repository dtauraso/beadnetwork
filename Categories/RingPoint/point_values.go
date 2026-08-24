package RingPoint

import (
	"path/filepath"
)

const ValueRelFile = "view/ring-points.bin"

var PointValueNames = []string{
	"nodeX", "nodeY", "nodeZ",
	"beadX", "beadY", "beadZ",
}

func ValueRelPath() string { return ValueRelFile }

type ValueWriter struct {
	*BlobWriter
}

func NewValueWriter(sceneRoot string) *ValueWriter {
	path := filepath.Join(sceneRoot, filepath.FromSlash(ValueRelFile))
	return &ValueWriter{BlobWriter: NewBlobWriter(path, PointValueNames)}
}

func (w *ValueWriter) Surface(xName, yName, zName string, pts []float32) {
	if len(pts)%3 != 0 {
		return
	}
	for i := 0; i < len(pts)/3; i++ {
		w.F32(xName, pts[i*3])
		w.F32(yName, pts[i*3+1])
		w.F32(zName, pts[i*3+2])
	}
}

// State is this piece's own: it carries the writer for its block, armed when
// the scene opens, so nothing outside can write it.
type State struct {
	w *ValueWriter
}

func (s *State) Arm(sceneRoot string) { s.w = NewValueWriter(sceneRoot) }

// W is this piece's writer, for the view's own column writing.
func (s *State) W() *ValueWriter { return s.w }
