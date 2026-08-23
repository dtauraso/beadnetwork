package nodefile

import (
	"fmt"
	"path/filepath"
	"strings"
)

func dragDir(root, id string) string {
	return filepath.Join(nodeDirPath(root, id), "drag")
}

func WriteDragIndex(root, id string, phi, theta, r int, topTiltVectorPhiIdx int32) error {
	if id == "" || id == "." || id == ".." || strings.ContainsAny(id, `/\`) {
		return fmt.Errorf("unsafe node id %q", id)
	}
	dir := dragDir(root, id)
	for name, value := range map[string]int{
		FileIndexPhi:   phi,
		FileIndexTheta: theta,
		FileIndexR:     r,
		FileTiltIdx:    int(topTiltVectorPhiIdx),
	} {
		if err := WriteAtomicIfChanged(filepath.Join(dir, name), value); err != nil {
			return err
		}
	}
	return nil
}
