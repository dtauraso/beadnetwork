package edgefile

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/src/valuefile"
	"github.com/dtauraso/wirefold/src/Polar/polarindex"
)

func edgeDragDir(root, src, label string) string {
	return filepath.Join(root, "nodes", src, "drag", "edges", label)
}

func ReadEdgeDragIndex(root, src, label string) (polarindex.Offset, bool) {
	dir := edgeDragDir(root, src, label)
	var off polarindex.Offset
	read := func(name string, dst *int) bool {
		return valuefile.ReadIfExists(filepath.Join(dir, name), dst)
	}
	if !read(FileDragIndexPhi, &off.Phi) || !read(FileDragIndexTheta, &off.Theta) ||
		!read(FileDragIndexR, &off.R) {
		return polarindex.Offset{}, false
	}
	return off, true
}

func WriteEdgeDrag(root, src, label string, off polarindex.Offset) error {
	dir := edgeDragDir(root, src, label)
	for name, value := range map[string]int{
		FileDragIndexPhi:   off.Phi,
		FileDragIndexTheta: off.Theta,
		FileDragIndexR:     off.R,
	} {
		if err := valuefile.WriteAtomicIfChanged(filepath.Join(dir, name), value); err != nil {
			return err
		}
	}
	return nil
}
