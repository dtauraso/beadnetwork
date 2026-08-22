package SceneVectors

import (
	"fmt"
	"path/filepath"
)

const ValueRelTemplate = "view/nodes/{row}/channel-vectors.bin"

var VectorValueNames = buildVectorValueNames()

func buildVectorValueNames() []string {
	names := make([]string, 0, 32)
	for m := 0; m < 16; m++ {
		names = append(names, fmt.Sprintf("shaftM%d", m))
	}
	for m := 0; m < 16; m++ {
		names = append(names, fmt.Sprintf("headM%d", m))
	}
	return names
}

func ShaftName(m int) string { return VectorValueNames[m] }
func HeadName(m int) string  { return VectorValueNames[16+m] }

func ValueRelPath(row int) string {
	return fmt.Sprintf("view/nodes/%d/channel-vectors.bin", row)
}

type ValueWriter struct {
	*BlobWriter
}

func NewValueWriter(sceneRoot string, row int) *ValueWriter {
	path := filepath.Join(sceneRoot, filepath.FromSlash(ValueRelPath(row)))
	return &ValueWriter{BlobWriter: NewBlobWriter(path, VectorValueNames)}
}
