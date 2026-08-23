package Node

import (
	"fmt"
	"path/filepath"
	"strings"
)

func dragDir(root, id string) string {
	return filepath.Join(nodeDirPath(root, id), "drag")
}

func ReadDragIndex(root, id string) (phi, theta, r int, topTiltVectorPhiIdx int32, ok bool) {
	dir := dragDir(root, id)
	read := func(name string, dst *int) {
		if ReadIfExists(filepath.Join(dir, name), dst) {
			ok = true
		}
	}
	read(FileIndexPhi, &phi)
	read(FileIndexTheta, &theta)
	read(FileIndexR, &r)

	var tilt int
	read(FileTiltIdx, &tilt)
	topTiltVectorPhiIdx = int32(tilt)

	if !ok {
		return 0, 0, 0, 0, false
	}
	return phi, theta, r, topTiltVectorPhiIdx, true
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
