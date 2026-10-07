package edgefile

import (
	"path/filepath"

	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

func edgeDragDir(root, src, label string) string {
	return filepath.Join(root, "nodes", src, "drag", "edges", label)
}

func ReadEdgeDragIndex(root, src, label string) (polarindex.Offset, bool) {
	dir := edgeDragDir(root, src, label)
	var off polarindex.Offset
	read := func(name string, dst *int) bool {
		return ReadIfExists(filepath.Join(dir, name), dst)
	}
	if !read(FileDragIndexPhi, &off.Phi) || !read(FileDragIndexTheta, &off.Theta) ||
		!read(FileDragIndexR, &off.R) {
		return polarindex.Offset{}, false
	}
	return off, true
}

func ReadEdgeDragTurn(root, src, label string) (phi, theta int, ok bool) {
	dir := edgeDragDir(root, src, label)
	if !ReadIfExists(filepath.Join(dir, FileDragTurnPhi), &phi) || !ReadIfExists(filepath.Join(dir, FileDragTurnTheta), &theta) {
		return 0, 0, false
	}
	return phi, theta, phi > 0 && theta > 0
}

func WriteEdgeDrag(root, src, label string, off polarindex.Offset, sc polarindex.SceneConstants) error {
	dir := edgeDragDir(root, src, label)
	for name, value := range map[string]int{
		FileDragIndexPhi:   off.Phi,
		FileDragIndexTheta: off.Theta,
		FileDragIndexR:     off.R,
		FileDragTurnPhi:    sc.MaxIndexPhi,
		FileDragTurnTheta:  sc.MaxIndexTheta,
	} {
		if err := WriteAtomicIfChanged(filepath.Join(dir, name), value); err != nil {
			return err
		}
	}
	return nil
}
