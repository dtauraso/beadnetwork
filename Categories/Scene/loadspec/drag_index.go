package loadspec

import (
	"path/filepath"
)

const (
	fileIndexPhi   = "index-phi.bin"
	fileIndexTheta = "index-theta.bin"
	fileIndexR     = "index-r.bin"
	fileTiltIdx    = "top-tilt-vector-phi-idx.bin"
)

type dragIndex struct {
	Phi   int
	Theta int
	R     int

	TopTiltVectorPhiIdx int32
}

func readDragIndex(root, id string) (dragIndex, bool) {
	dir := filepath.Join(root, "nodes", id, "drag")

	var d dragIndex
	found := false
	read := func(name string, dst *int) {
		if ReadIfExists(filepath.Join(dir, name), dst) {
			found = true
		}
	}
	read(fileIndexPhi, &d.Phi)
	read(fileIndexTheta, &d.Theta)
	read(fileIndexR, &d.R)

	var tilt int
	read(fileTiltIdx, &tilt)
	d.TopTiltVectorPhiIdx = int32(tilt)

	if !found {
		return dragIndex{}, false
	}
	return d, true
}
