package NodeShape

import (
	"path/filepath"
)

const RingPointValueRelFile = "view/node-ring-points.bin"

var RingPointValueNames = []string{"x", "y", "z"}

func RingPointValueRelPath() string { return RingPointValueRelFile }

type RingPointValueWriter struct {
	*BlobWriter
}

func NewRingPointValueWriter(sceneRoot string) *RingPointValueWriter {
	path := filepath.Join(sceneRoot, filepath.FromSlash(RingPointValueRelFile))
	return &RingPointValueWriter{BlobWriter: NewBlobWriter(path, RingPointValueNames)}
}

func (w *RingPointValueWriter) Surface(pts []float32) {
	if len(pts)%3 != 0 {
		return
	}
	for i := 0; i < len(pts)/3; i++ {
		w.F32("x", pts[i*3])
		w.F32("y", pts[i*3+1])
		w.F32("z", pts[i*3+2])
	}
}

type RingPointState struct {
	w *RingPointValueWriter
}

func (s *RingPointState) Arm(sceneRoot string) { s.w = NewRingPointValueWriter(sceneRoot) }

func (s *RingPointState) W() *RingPointValueWriter { return s.w }
