package edge

import (
	"fmt"
	"path/filepath"

	"github.com/dtauraso/wirefold/src/valuefile"
)

const ValueRelTemplate = "view/edges/{row}/edge.bin"

var EdgeValueNames = []string{
	"sx", "sy", "sz", "ex", "ey", "ez",
	"srcNodeRow", "dstNodeRow", "deltaR", "dragActive", "label",
}

func ValueRelPath(row int) string {
	return fmt.Sprintf("view/edges/%d/edge.bin", row)
}

type ValueWriter struct {
	*valuefile.BlobWriter
}

func NewValueWriter(sceneRoot string, row int) *ValueWriter {
	path := filepath.Join(sceneRoot, filepath.FromSlash(ValueRelPath(row)))
	return &ValueWriter{BlobWriter: valuefile.NewBlobWriter(path, EdgeValueNames)}
}

func (w *ValueWriter) Write(
	sx, sy, sz, ex, ey, ez float32,
	srcNodeRow, dstNodeRow int32, deltaR float32, dragActive uint8, label string,
) error {
	w.Begin()
	w.F32("sx", sx)
	w.F32("sy", sy)
	w.F32("sz", sz)
	w.F32("ex", ex)
	w.F32("ey", ey)
	w.F32("ez", ez)
	w.I32("srcNodeRow", srcNodeRow)
	w.I32("dstNodeRow", dstNodeRow)
	w.F32("deltaR", deltaR)
	w.U8("dragActive", dragActive)
	w.Text("label", label)
	return w.Flush()
}
