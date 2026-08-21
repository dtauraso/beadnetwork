package Scene

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/src/valuefile"
)

const CountsValueRelPath = "view/owner-counts.bin"

var CountsValueNames = []string{"nodes", "edges"}

type CountsValueWriter struct {
	*valuefile.BlobWriter
}

func NewCountsValueWriter(sceneRoot string) *CountsValueWriter {
	path := filepath.Join(sceneRoot, filepath.FromSlash(CountsValueRelPath))
	return &CountsValueWriter{BlobWriter: valuefile.NewBlobWriter(path, CountsValueNames)}
}

func (w *CountsValueWriter) Write(nodes, edges int32) error {
	w.Begin()
	w.I32("nodes", nodes)
	w.I32("edges", edges)
	return w.Flush()
}
