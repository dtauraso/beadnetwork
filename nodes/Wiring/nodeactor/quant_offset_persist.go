package nodeactor

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/dragfile"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/quantoffset"
)

func (nm *NodeGeometry) persistQuantOffset(off quantoffset.QuantizedOffset) {
	if nm.persistRoot == "" {
		return
	}

	if err := writeQuantOffset(nm.persistRoot, nm.id, off, nm.geom.DragPolar, nm.tilt.TopTiltVectorPhiIdx()); err != nil {
		jsonpersist.LogPersistErr("quant_offset_persist", nm.id, err)
	}
}

func (nm *NodeGeometry) persistTiltVectorAngle() {
	if nm.persistRoot == "" {
		return
	}
	if err := writeQuantOffset(nm.persistRoot, nm.id, nm.quant.Drag(), nm.geom.DragPolar, nm.tilt.TopTiltVectorPhiIdx()); err != nil {
		jsonpersist.LogPersistErr("quant_offset_persist", nm.id, err)
	}
}

func writeQuantOffset(root, id string, off quantoffset.QuantizedOffset, delta polar.Polar, topTiltVectorPhiIdx int32) error {
	if !jsonpersist.SafeTreePathComponent(id) {
		return fmt.Errorf("unsafe node id %q", id)
	}
	return dragfile.Write(root, id, dragfile.JSON{
		DragPolarR: delta.R, DragPolarPhi: delta.Phi, DragPolarTheta: delta.Theta,
		IPhi: off.IPhi, ITheta: off.ITheta, IR: off.IR,
		TopTiltVectorPhiIdx: topTiltVectorPhiIdx,
	})
}
