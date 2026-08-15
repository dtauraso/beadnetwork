package polarindex

import (
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/spatial"
)

type vec3 = spatial.Vec3

type SceneConstants struct {
	ConstantR     float64 `json:"constantR"`
	ConstantPhi   float64 `json:"constantPhi"`
	ConstantTheta float64 `json:"constantTheta"`
}

type Index struct {
	Phi   int
	Theta int
	R     int
}

func MeasureScalars(centers map[string]vec3, ids map[string]bool, sceneCenter vec3, sc SceneConstants) map[string]Index {
	result := make(map[string]Index, len(ids))
	for id := range ids {
		pos, ok := centers[id]
		if !ok {
			continue
		}
		p := polar.Cart2polar(pos.Sub(sceneCenter))
		result[id] = Index{
			Phi:   int(math.Round(p.Phi / sc.ConstantPhi)),
			Theta: int(math.Round(p.Theta / sc.ConstantTheta)),
			R:     int(math.Round(p.R / sc.ConstantR)),
		}
	}
	return result
}

func MeasureScalar(p polar.Polar, sc SceneConstants) Index {
	return Index{
		Phi:   int(math.Round(p.Phi / sc.ConstantPhi)),
		Theta: int(math.Round(p.Theta / sc.ConstantTheta)),
		R:     int(math.Round(p.R / sc.ConstantR)),
	}
}

func offsetScenePolar(o Index, sc SceneConstants) polar.Polar {
	return polar.Polar{R: float64(o.R) * sc.ConstantR, Phi: float64(o.Phi) * sc.ConstantPhi, Theta: float64(o.Theta) * sc.ConstantTheta}
}

func DeriveCenters(scalars map[string]Index, sceneCenter vec3, sc SceneConstants) map[string]vec3 {
	derived := make(map[string]vec3, len(scalars))
	for id, o := range scalars {
		derived[id] = sceneCenter.Add(polar.Polar2cart(offsetScenePolar(o, sc)))
	}
	return derived
}

func Compose(base, drag Index) Index {
	return Index{
		Phi:   base.Phi + drag.Phi,
		Theta: base.Theta + drag.Theta,
		R:     base.R + drag.R,
	}
}

func Delta(composed, base Index) Index {
	return Index{
		Phi:   composed.Phi - base.Phi,
		Theta: composed.Theta - base.Theta,
		R:     composed.R - base.R,
	}
}
