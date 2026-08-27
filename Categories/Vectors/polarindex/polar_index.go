package polarindex

import (
	"math"

	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
)

type SceneConstants struct {
	ConstantR     float64
	MaxIndexPhi   int
	MaxIndexTheta int
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
	phi := ((o.Phi % turnPhi) + turnPhi) % turnPhi
	theta := ((o.Theta % turnTheta) + turnTheta) % turnTheta
	return Index{Phi: phi, Theta: theta, R: o.R}
}

func MeasureScalars(centers map[string]Vec3, ids map[string]bool, sceneCenter Vec3, sc SceneConstants) map[string]Index {
	result := make(map[string]Index, len(ids))
	for id := range ids {
		pos, ok := centers[id]
		if !ok {
			continue
		}
		p := polar.Cart2polar(polar.Vec3(pos.Sub(sceneCenter)))
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

func DeriveCenters(scalars map[string]Index, sceneCenter Vec3, sc SceneConstants) map[string]Vec3 {
	derived := make(map[string]Vec3, len(scalars))
	for id, o := range scalars {
		derived[id] = sceneCenter.Add(Vec3(polar.Polar2cart(ToPolar(o, sc))))
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
