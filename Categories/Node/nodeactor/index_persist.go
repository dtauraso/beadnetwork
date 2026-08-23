package nodeactor

import (
	"fmt"

	"github.com/dtauraso/wirefold/Categories/Node/nodeactor/nodefiles"
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
)

func (nm *NodeGeometry) persistIndex(off polarindex.Offset) {
	if nm.persistRoot == "" {
		return
	}

	if err := writeIndex(nm.persistRoot, nm.id, off, nm.tilt.TopTiltVectorPhiIdx()); err != nil {
		LogPersistErr("index_persist", nm.id, err)
	}
}

func (nm *NodeGeometry) persistTiltVectorAngle() {
	if nm.persistRoot == "" {
		return
	}
	if err := writeIndex(nm.persistRoot, nm.id, nm.geom.DragIndex, nm.tilt.TopTiltVectorPhiIdx()); err != nil {
		LogPersistErr("index_persist", nm.id, err)
	}
}

func writeIndex(root, id string, off polarindex.Offset, topTiltVectorPhiIdx int32) error {
	if !SafeTreePathComponent(id) {
		return fmt.Errorf("unsafe node id %q", id)
	}
	return nodefiles.WriteDragIndex(root, id, off.Phi, off.Theta, off.R, topTiltVectorPhiIdx)
}
