package Node

import (
	"github.com/dtauraso/beadnetwork/Categories/Node/TiltVectors"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

func TipPoint(center Vec3, vec polar.Polar) Vec3 {
	return center.Add(Vec3(polar.Polar2cart(vec)))
}

func TipIndex(tip, sceneCenter Vec3, sc polarindex.SceneConstants) polarindex.Index {
	return polarindex.MeasureIndex(polar.Cart2polar(polar.Vec3(tip.Sub(sceneCenter))), sc)
}

func startArrows(g NodeGeom, center Vec3, starts []polar.Polar) []TiltVectors.TiltArrow {
	size := NodeRadius(g.Kind)
	var out []TiltVectors.TiltArrow
	for _, vec := range starts {
		tip := TipPoint(center, vec)
		if a, ok := TiltVectors.ArrowBetween(TiltVectors.Vec3(center), TiltVectors.Vec3(tip), size, TiltVectors.ArrowStart); ok {
			out = append(out, a)
		}
	}
	return out
}
