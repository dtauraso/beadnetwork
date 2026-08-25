package TopVector

import (
	"fmt"
	"path/filepath"
)

const ValueRelTemplate = "view/nodes/{row}/top-vector.bin"

var TopVectorValueNames = buildTopVectorValueNames()

func buildTopVectorValueNames() []string {
	names := []string{"drawn"}
	for m := range 16 {
		names = append(names, fmt.Sprintf("topShaftM%d", m))
	}
	for m := range 16 {
		names = append(names, fmt.Sprintf("topHeadM%d", m))
	}
	return names
}

func ShaftName(m int) string { return TopVectorValueNames[1+m] }
func HeadName(m int) string  { return TopVectorValueNames[17+m] }

func ValueRelPath(row int) string {
	return fmt.Sprintf("view/nodes/%d/top-vector.bin", row)
}

type ValueWriter struct {
	*BlobWriter
}

func NewValueWriter(sceneRoot string, row int) *ValueWriter {
	path := filepath.Join(sceneRoot, filepath.FromSlash(ValueRelPath(row)))
	return &ValueWriter{BlobWriter: NewBlobWriter(path, TopVectorValueNames)}
}

type Frame struct {
	Drawn bool

	Shaft [16]float32
	Head  [16]float32
}

func (w *ValueWriter) Write(f Frame) error {
	w.Begin()
	w.Bool("drawn", f.Drawn)
	for m := range 16 {
		w.F32(ShaftName(m), f.Shaft[m])
		w.F32(HeadName(m), f.Head[m])
	}
	return w.Flush()
}
