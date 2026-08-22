package Scene

import (
	"path/filepath"
)

const CountsValueRelPath = "view/owner-counts.bin"

var CountsValueNames = []string{"nodes", "edges"}

type CountsValueWriter struct {
	*BlobWriter
}

func NewCountsValueWriter(sceneRoot string) *CountsValueWriter {
	path := filepath.Join(sceneRoot, filepath.FromSlash(CountsValueRelPath))
	return &CountsValueWriter{BlobWriter: NewBlobWriter(path, CountsValueNames)}
}

func (w *CountsValueWriter) Write(nodes, edges int32) error {
	w.Begin()
	w.I32("nodes", nodes)
	w.I32("edges", edges)
	return w.Flush()
}
