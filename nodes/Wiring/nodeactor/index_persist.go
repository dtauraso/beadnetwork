package nodeactor

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/dragfile"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/polarindex"
)

func (nm *NodeGeometry) persistIndex(off polarindex.Offset) {
	if nm.persistRoot == "" {
		return
	}

	if err := writeIndex(nm.persistRoot, nm.id, off, nm.tilt.TopTiltVectorPhiIdx()); err != nil {
		jsonpersist.LogPersistErr("index_persist", nm.id, err)
	}
}

func (nm *NodeGeometry) persistTiltVectorAngle() {
	if nm.persistRoot == "" {
		return
	}
	if err := writeIndex(nm.persistRoot, nm.id, nm.geom.DragIndex, nm.tilt.TopTiltVectorPhiIdx()); err != nil {
		jsonpersist.LogPersistErr("index_persist", nm.id, err)
	}
}

func writeIndex(root, id string, off polarindex.Offset, topTiltVectorPhiIdx int32) error {
	if !jsonpersist.SafeTreePathComponent(id) {
		return fmt.Errorf("unsafe node id %q", id)
	}
	return dragfile.Write(root, id, dragfile.JSON{
		IndexPhi: off.Phi, IndexTheta: off.Theta, IndexR: off.R,
		TopTiltVectorPhiIdx: topTiltVectorPhiIdx,
	})
}
