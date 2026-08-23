package nodefiles

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	fileIndexPhi   = "index-phi.bin"
	fileIndexTheta = "index-theta.bin"
	fileIndexR     = "index-r.bin"
	fileTiltIdx    = "top-tilt-vector-phi-idx.bin"
)

func dragDir(root, id string) string {
	return filepath.Join(root, "nodes", id, "drag")
}

func WriteDragIndex(root, id string, phi, theta, r int, topTiltVectorPhiIdx int32) error {
	if id == "" || id == "." || id == ".." || strings.ContainsAny(id, `/\`) {
		return fmt.Errorf("unsafe node id %q", id)
	}
	dir := dragDir(root, id)
	for name, value := range map[string]int{
		fileIndexPhi:   phi,
		fileIndexTheta: theta,
		fileIndexR:     r,
		fileTiltIdx:    int(topTiltVectorPhiIdx),
	} {
		if err := WriteAtomicIfChanged(filepath.Join(dir, name), value); err != nil {
			return err
		}
	}
	return nil
}
