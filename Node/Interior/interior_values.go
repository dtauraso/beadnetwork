package interior

import (
	"fmt"
	"path/filepath"
)

const ValueRelTemplate = "view/nodes/{row}/interior.bin"

var InteriorValueNames = []string{
	"present", "value", "x", "y", "z",
}

func ValueRelPath(row int) string {
	return fmt.Sprintf("view/nodes/%d/interior.bin", row)
}

type ValueWriter struct {
	*BlobWriter
}

func NewValueWriter(sceneRoot string, row int) *ValueWriter {
	path := filepath.Join(sceneRoot, filepath.FromSlash(ValueRelPath(row)))
	return &ValueWriter{BlobWriter: NewBlobWriter(path, InteriorValueNames)}
}
