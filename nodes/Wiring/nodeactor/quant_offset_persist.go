package nodeactor

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/positionfile"
	"github.com/dtauraso/wirefold/nodes/Wiring/quantoffset"
)

func (nm *NodeGeometry) persistQuantOffset(off quantoffset.QuantizedOffset, scene geom.Polar) {
	if nm.persistRoot == "" {
		return
	}

	if err := writeQuantOffset(nm.persistRoot, nm.id, off, scene, nm.tilt.topTiltVectorThetaIdx); err != nil {
		jsonpersist.LogPersistErr("quant_offset_persist", nm.id, err)
	}
}

func (nm *NodeGeometry) persistTiltVectorAngle() {
	if nm.persistRoot == "" {
		return
	}
	if err := writeQuantOffset(nm.persistRoot, nm.id, nm.quantOffset, nm.geom.ScenePolar, nm.tilt.topTiltVectorThetaIdx); err != nil {
		jsonpersist.LogPersistErr("quant_offset_persist", nm.id, err)
	}
}

func writeQuantOffset(root, id string, off quantoffset.QuantizedOffset, scene geom.Polar, topTiltVectorThetaIdx int32) error {
	if !jsonpersist.SafeTreePathComponent(id) {
		return fmt.Errorf("unsafe node id %q", id)
	}
	t, p, r := off.EffectiveSteps()
	return jsonpersist.WriteJSONAtomic(positionfile.FilePath(root, id), positionfile.JSON{
		ScenePolarR: scene.R, ScenePolarTheta: scene.Theta, ScenePolarPhi: scene.Phi,
		QuantITheta: off.ITheta, QuantIPhi: off.IPhi, QuantIR: off.IR,
		StepTheta: t, StepPhi: p, StepR: r,
		TopTiltVectorThetaIdx: topTiltVectorThetaIdx,
	})
}
