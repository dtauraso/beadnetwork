package nodeactor

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/dragfile"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/polarindex"
)

func (nm *NodeGeometry) persistQuantOffset(off polarindex.Index) {
	if nm.persistRoot == "" {
		return
	}

	if err := writeQuantOffset(nm.persistRoot, nm.id, off, nm.tilt.TopTiltVectorPhiIdx()); err != nil {
		jsonpersist.LogPersistErr("quant_offset_persist", nm.id, err)
	}
}

func (nm *NodeGeometry) persistTiltVectorAngle() {
	if nm.persistRoot == "" {
		return
	}
	if err := writeQuantOffset(nm.persistRoot, nm.id, nm.quant.Drag(), nm.tilt.TopTiltVectorPhiIdx()); err != nil {
		jsonpersist.LogPersistErr("quant_offset_persist", nm.id, err)
	}
}

func writeQuantOffset(root, id string, off polarindex.Index, topTiltVectorPhiIdx int32) error {
	if !jsonpersist.SafeTreePathComponent(id) {
		return fmt.Errorf("unsafe node id %q", id)
	}
	return dragfile.Write(root, id, dragfile.JSON{
		IndexPhi: off.Phi, IndexTheta: off.Theta, IndexR: off.R,
		TopTiltVectorPhiIdx: topTiltVectorPhiIdx,
	})
}
