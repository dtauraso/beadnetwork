package RingPoint

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/src/valuefile"
)

const ValueRelFile = "view/ring-points.bin"

var PointValueNames = []string{
	"nodeX", "nodeY", "nodeZ",
	"beadX", "beadY", "beadZ",
}

func ValueRelPath() string { return ValueRelFile }

type ValueWriter struct {
	*valuefile.BlobWriter
}

func NewValueWriter(sceneRoot string) *ValueWriter {
	path := filepath.Join(sceneRoot, filepath.FromSlash(ValueRelFile))
	return &ValueWriter{BlobWriter: valuefile.NewBlobWriter(path, PointValueNames)}
}

// Surface takes the flat x,y,z triples the canonical ring is computed as and
// writes them as three runs, the shape the mesh builder reads.
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
