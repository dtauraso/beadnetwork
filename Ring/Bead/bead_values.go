package bead

import (
	"fmt"
	"path/filepath"
)

const ValueRelTemplate = "view/nodes/{row}/beads.bin"

var BeadValueNames = []string{
	"x", "y", "z", "value",
	"ringM0", "ringM1", "ringM2", "ringM3",
	"ringM4", "ringM5", "ringM6", "ringM7",
	"ringM8", "ringM9", "ringM10", "ringM11",
	"ringM12", "ringM13", "ringM14", "ringM15",
}

func ValueRelPath(row int) string {
	return fmt.Sprintf("view/nodes/%d/beads.bin", row)
}

type ValueWriter struct {
	*BlobWriter
}

func NewValueWriter(sceneRoot string, row int) *ValueWriter {
	path := filepath.Join(sceneRoot, filepath.FromSlash(ValueRelPath(row)))
	return &ValueWriter{BlobWriter: NewBlobWriter(path, BeadValueNames)}
}

func RingName(m int) string { return BeadValueNames[4+m] }
