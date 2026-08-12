package quantoffset

import (
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/wire"
	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"
)

type vec3 = wire.Vec3

const (
	stepTheta = math.Pi / 180
	stepPhi   = math.Pi / 180

	stepR = lattice.BeadStepR
)

type QuantizedOffset struct {
	ITheta int
	IPhi   int
	IR     int

	CTheta float64
	CPhi   float64
	CR     float64
}

func (o QuantizedOffset) EffectiveSteps() (t, p, r float64) {
	t, p, r = o.CTheta, o.CPhi, o.CR
	if t == 0 {
		t = stepTheta
	}
	if p == 0 {
		p = stepPhi
	}
	if r == 0 {
		r = stepR
	}
	return
}

func MeasureScalars(centers map[string]vec3, ids map[string]bool, sceneCenter vec3, prior map[string]QuantizedOffset) map[string]QuantizedOffset {
	result := make(map[string]QuantizedOffset, len(ids))
	for id := range ids {
		pos, ok := centers[id]
		if !ok {
			continue
		}
		carried := prior[id]
		t, p_, r := carried.EffectiveSteps()
		p := geom.Cart2polar(pos.Sub(sceneCenter))
		result[id] = QuantizedOffset{
			ITheta: int(math.Round(p.Theta / t)),
			IPhi:   int(math.Round(p.Phi / p_)),
			IR:     int(math.Round(p.R / r)),
			CTheta: carried.CTheta,
			CPhi:   carried.CPhi,
			CR:     carried.CR,
		}
	}
	return result
}

func MeasureScalar(p geom.Polar, prior QuantizedOffset) QuantizedOffset {
	t, p_, r := prior.EffectiveSteps()
	return QuantizedOffset{
		ITheta: int(math.Round(p.Theta / t)),
		IPhi:   int(math.Round(p.Phi / p_)),
		IR:     int(math.Round(p.R / r)),
		CTheta: prior.CTheta,
		CPhi:   prior.CPhi,
		CR:     prior.CR,
	}
}

func offsetScenePolar(o QuantizedOffset) geom.Polar {
	t, p, r := o.EffectiveSteps()
	return geom.Polar{R: float64(o.IR) * r, Theta: float64(o.ITheta) * t, Phi: float64(o.IPhi) * p}
}

func DeriveCenters(scalars map[string]QuantizedOffset, sceneCenter vec3) map[string]vec3 {
	derived := make(map[string]vec3, len(scalars))
	for id, o := range scalars {
		derived[id] = sceneCenter.Add(geom.Polar2cart(offsetScenePolar(o)))
	}
	return derived
}

func NormalizeOffset(o QuantizedOffset) QuantizedOffset {
	if o.CTheta != 0 && math.Abs(o.CTheta-stepTheta) > 1e-9 {
		o.ITheta = int(math.Round(float64(o.ITheta) * o.CTheta / stepTheta))
		o.CTheta = stepTheta
	}
	if o.CPhi != 0 && math.Abs(o.CPhi-stepPhi) > 1e-9 {
		o.IPhi = int(math.Round(float64(o.IPhi) * o.CPhi / stepPhi))
		o.CPhi = stepPhi
	}
	if o.CR != 0 && math.Abs(o.CR-stepR) > 1e-9 {
		o.IR = int(math.Round(float64(o.IR) * o.CR / stepR))
		o.CR = stepR
	}
	return o
}
