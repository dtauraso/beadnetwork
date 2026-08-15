package quantoffset

import (
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/spatial"
	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"
)

type vec3 = spatial.Vec3

const (
	constantPhi   = math.Pi / 180
	constantTheta = math.Pi / 180

	constantR = lattice.BeadStepR
)

type QuantizedOffset struct {
	IPhi   int
	ITheta int
	IR     int

	ConstantPhi   float64
	ConstantTheta float64
	ConstantR     float64
}

func (o QuantizedOffset) EffectiveConstants() (t, p, r float64) {
	t, p, r = o.ConstantPhi, o.ConstantTheta, o.ConstantR
	if t == 0 {
		t = constantPhi
	}
	if p == 0 {
		p = constantTheta
	}
	if r == 0 {
		r = constantR
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
		t, p_, r := carried.EffectiveConstants()
		p := polar.Cart2polar(pos.Sub(sceneCenter))
		result[id] = QuantizedOffset{
			IPhi:          int(math.Round(p.Phi / t)),
			ITheta:        int(math.Round(p.Theta / p_)),
			IR:            int(math.Round(p.R / r)),
			ConstantPhi:   carried.ConstantPhi,
			ConstantTheta: carried.ConstantTheta,
			ConstantR:     carried.ConstantR,
		}
	}
	return result
}

func MeasureScalar(p polar.Polar, prior QuantizedOffset) QuantizedOffset {
	t, p_, r := prior.EffectiveConstants()
	return QuantizedOffset{
		IPhi:          int(math.Round(p.Phi / t)),
		ITheta:        int(math.Round(p.Theta / p_)),
		IR:            int(math.Round(p.R / r)),
		ConstantPhi:   prior.ConstantPhi,
		ConstantTheta: prior.ConstantTheta,
		ConstantR:     prior.ConstantR,
	}
}

func offsetScenePolar(o QuantizedOffset) polar.Polar {
	t, p, r := o.EffectiveConstants()
	return polar.Polar{R: float64(o.IR) * r, Phi: float64(o.IPhi) * t, Theta: float64(o.ITheta) * p}
}

func DeriveCenters(scalars map[string]QuantizedOffset, sceneCenter vec3) map[string]vec3 {
	derived := make(map[string]vec3, len(scalars))
	for id, o := range scalars {
		derived[id] = sceneCenter.Add(polar.Polar2cart(offsetScenePolar(o)))
	}
	return derived
}

func Compose(base, drag QuantizedOffset) QuantizedOffset {
	return QuantizedOffset{
		IPhi:          base.IPhi + drag.IPhi,
		ITheta:        base.ITheta + drag.ITheta,
		IR:            base.IR + drag.IR,
		ConstantPhi:   base.ConstantPhi,
		ConstantTheta: base.ConstantTheta,
		ConstantR:     base.ConstantR,
	}
}

func Delta(composed, base QuantizedOffset) QuantizedOffset {
	return QuantizedOffset{
		IPhi:   composed.IPhi - base.IPhi,
		ITheta: composed.ITheta - base.ITheta,
		IR:     composed.IR - base.IR,
	}
}

func NormalizeOffset(o QuantizedOffset) QuantizedOffset {
	if o.ConstantPhi != 0 && math.Abs(o.ConstantPhi-constantPhi) > 1e-9 {
		o.IPhi = int(math.Round(float64(o.IPhi) * o.ConstantPhi / constantPhi))
		o.ConstantPhi = constantPhi
	}
	if o.ConstantTheta != 0 && math.Abs(o.ConstantTheta-constantTheta) > 1e-9 {
		o.ITheta = int(math.Round(float64(o.ITheta) * o.ConstantTheta / constantTheta))
		o.ConstantTheta = constantTheta
	}
	if o.ConstantR != 0 && math.Abs(o.ConstantR-constantR) > 1e-9 {
		o.IR = int(math.Round(float64(o.IR) * o.ConstantR / constantR))
		o.ConstantR = constantR
	}
	return o
}
