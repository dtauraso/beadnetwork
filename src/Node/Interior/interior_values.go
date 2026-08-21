package interior

import (
	"fmt"
	"path/filepath"

	"github.com/dtauraso/wirefold/src/valuefile"
)

const ValueRelTemplate = "view/nodes/{row}/interior.bin"

var InteriorValueNames = []string{
	"present", "value", "x", "y", "z",
}

func ValueRelPath(row int) string {
	return fmt.Sprintf("view/nodes/%d/interior.bin", row)
}

type ValueWriter struct {
	*valuefile.BlobWriter
}

func NewValueWriter(sceneRoot string, row int) *ValueWriter {
	path := filepath.Join(sceneRoot, filepath.FromSlash(ValueRelPath(row)))
	return &ValueWriter{BlobWriter: valuefile.NewBlobWriter(path, InteriorValueNames)}
}
