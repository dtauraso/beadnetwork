package TiltVectors

import (
	"fmt"
	"path/filepath"
)

const ValueRelTemplate = "view/nodes/{row}/tilt-arrows.bin"

var TiltValueNames = buildTiltValueNames()

func buildTiltValueNames() []string {
	names := []string{"received"}
	for m := range 16 {
		names = append(names, fmt.Sprintf("shaftM%d", m))
	}
	for m := range 16 {
		names = append(names, fmt.Sprintf("headM%d", m))
	}
	return names
}

func ShaftName(m int) string { return TiltValueNames[1+m] }
func HeadName(m int) string  { return TiltValueNames[17+m] }

func ValueRelPath(row int) string {
	return fmt.Sprintf("view/nodes/%d/tilt-arrows.bin", row)
}

type ValueWriter struct {
	*BlobWriter
}

func NewValueWriter(sceneRoot string, row int) *ValueWriter {
	path := filepath.Join(sceneRoot, filepath.FromSlash(ValueRelPath(row)))
	return &ValueWriter{BlobWriter: NewBlobWriter(path, TiltValueNames)}
}
