package nodeactor

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/positionfile"
	"github.com/dtauraso/wirefold/nodes/Wiring/quantoffset"
)

func (nm *NodeGeometry) persistQuantOffset(off quantoffset.QuantizedOffset, scene polar.Polar) {
	if nm.persistRoot == "" {
		return
	}

	if err := writeQuantOffset(nm.persistRoot, nm.id, off, scene, nm.tilt.TopTiltVectorThetaIdx()); err != nil {
		jsonpersist.LogPersistErr("quant_offset_persist", nm.id, err)
	}
}

func (nm *NodeGeometry) persistTiltVectorAngle() {
	if nm.persistRoot == "" {
		return
	}
	if err := writeQuantOffset(nm.persistRoot, nm.id, nm.quantOffset, nm.geom.ScenePolar, nm.tilt.TopTiltVectorThetaIdx()); err != nil {
		jsonpersist.LogPersistErr("quant_offset_persist", nm.id, err)
	}
}

func writeQuantOffset(root, id string, off quantoffset.QuantizedOffset, scene polar.Polar, topTiltVectorThetaIdx int32) error {
	if !jsonpersist.SafeTreePathComponent(id) {
		return fmt.Errorf("unsafe node id %q", id)
	}
	t, p, r := off.EffectiveSteps()
	return jsonpersist.WriteJSONAtomic(positionfile.FilePath(root, id), positionfile.JSON{
		ScenePolarR: scene.R, ScenePolarPhi: scene.Phi, ScenePolarTheta: scene.Theta,
		QuantIPhi: off.IPhi, QuantITheta: off.ITheta, QuantIR: off.IR,
		StepPhi: t, StepTheta: p, StepR: r,
		TopTiltVectorThetaIdx: topTiltVectorThetaIdx,
	})
}
