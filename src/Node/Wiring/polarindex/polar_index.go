package polarindex

import (
	"math"

	"github.com/dtauraso/wirefold/src/Node/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/src/Node/spatial"
)

type vec3 = spatial.Vec3

type SceneConstants struct {
	ConstantR     float64 `json:"constantR"`
	MaxIndexPhi   int     `json:"maxIndexPhi"`
	MaxIndexTheta int     `json:"maxIndexTheta"`
}

func (sc SceneConstants) ConstantPhi() float64 {
	if sc.MaxIndexPhi == 0 {
		return 0
	}
	return 2 * math.Pi / float64(sc.MaxIndexPhi)
}

func (sc SceneConstants) ConstantTheta() float64 {
	if sc.MaxIndexTheta == 0 {
		return 0
	}
	return 2 * math.Pi / float64(sc.MaxIndexTheta)
}

type Index struct {
	Phi   int
	Theta int
	R     int
}

type Offset struct {
	Phi   int
	Theta int
	R     int
}

func Canonical(o Index, sc SceneConstants) Index {
	turnPhi, turnTheta := sc.MaxIndexPhi, sc.MaxIndexTheta
	if turnPhi == 0 || turnTheta == 0 {
		return o
	}
	halfTheta := turnTheta / 2

	phi := ((o.Phi % turnPhi) + turnPhi) % turnPhi
	theta := ((o.Theta+halfTheta)%turnTheta+turnTheta)%turnTheta - halfTheta
	return Index{Phi: phi, Theta: theta, R: o.R}
}

func MeasureScalars(centers map[string]vec3, ids map[string]bool, sceneCenter vec3, sc SceneConstants) map[string]Index {
	result := make(map[string]Index, len(ids))
	for id := range ids {
		pos, ok := centers[id]
		if !ok {
			continue
		}
		p := polar.Cart2polar(pos.Sub(sceneCenter))
		result[id] = Canonical(Index{
			Phi:   int(math.Round(p.Phi / sc.ConstantPhi())),
			Theta: int(math.Round(p.Theta / sc.ConstantTheta())),
			R:     int(math.Round(p.R / sc.ConstantR)),
		}, sc)
	}
	return result
}

func MeasureIndex(p polar.Polar, sc SceneConstants) Index {
	return Canonical(Index{
		Phi:   int(math.Round(p.Phi / sc.ConstantPhi())),
		Theta: int(math.Round(p.Theta / sc.ConstantTheta())),
		R:     int(math.Round(p.R / sc.ConstantR)),
	}, sc)
}

func MeasureOffset(p polar.Polar, sc SceneConstants) Offset {
	return Offset{
		Phi:   int(math.Round(p.Phi / sc.ConstantPhi())),
		Theta: int(math.Round(p.Theta / sc.ConstantTheta())),
		R:     int(math.Round(p.R / sc.ConstantR)),
	}
}

func ToPolar(o Index, sc SceneConstants) polar.Polar {
	return polar.Polar{R: float64(o.R) * sc.ConstantR, Phi: float64(o.Phi) * sc.ConstantPhi(), Theta: float64(o.Theta) * sc.ConstantTheta()}
}

func OffsetToPolar(o Offset, sc SceneConstants) polar.Polar {
	return polar.Polar{R: float64(o.R) * sc.ConstantR, Phi: float64(o.Phi) * sc.ConstantPhi(), Theta: float64(o.Theta) * sc.ConstantTheta()}
}

func DeriveCenters(scalars map[string]Index, sceneCenter vec3, sc SceneConstants) map[string]vec3 {
	derived := make(map[string]vec3, len(scalars))
	for id, o := range scalars {
		derived[id] = sceneCenter.Add(polar.Polar2cart(ToPolar(o, sc)))
	}
	return derived
}

func Compose(base Index, off Offset, sc SceneConstants) Index {
	return Canonical(Index{
		Phi:   base.Phi + off.Phi,
		Theta: base.Theta + off.Theta,
		R:     base.R + off.R,
	}, sc)
}

func Sum(a, b Offset) Offset {
	return Offset{Phi: a.Phi + b.Phi, Theta: a.Theta + b.Theta, R: a.R + b.R}
}

func Delta(composed, base Index) Offset {
	return Offset{
		Phi:   composed.Phi - base.Phi,
		Theta: composed.Theta - base.Theta,
		R:     composed.R - base.R,
	}
}

func Neg(o Offset) Offset {
	return Offset{Phi: -o.Phi, Theta: -o.Theta, R: -o.R}
}
