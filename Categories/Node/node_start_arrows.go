package Node

import (
	"github.com/dtauraso/beadnetwork/Categories/Node/TiltVectors"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
)

func HeadPoint(center Vec3, vec polar.Polar) Vec3 {
	return center.Add(Vec3(polar.Polar2cart(vec)))
}

func startArrows(g NodeGeom, center Vec3, starts []polar.Polar, picked []bool) []TiltVectors.TiltArrow {
	size := NodeRadius(g.Kind)
	var out []TiltVectors.TiltArrow
	for i, vec := range starts {
		head := HeadPoint(center, vec)
		kind := TiltVectors.ArrowStartStill
		if i < len(picked) && picked[i] {
			kind = TiltVectors.ArrowStart
		}
		if a, ok := TiltVectors.ArrowBetween(TiltVectors.Vec3(center), TiltVectors.Vec3(head), size, kind); ok {
			out = append(out, a)
		}
	}
	return out
}
