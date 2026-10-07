package Node

import (
	"github.com/dtauraso/beadnetwork/Categories/Node/TiltVectors"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

func startArrows(center Vec3, size float64, sc polarindex.SceneConstants, starts []polarindex.Offset) []TiltVectors.TiltArrow {
	var out []TiltVectors.TiltArrow
	for _, off := range starts {
		tip := center.Add(Vec3(polar.Polar2cart(polarindex.OffsetToPolar(off, sc))))
		if a, ok := TiltVectors.ArrowBetween(TiltVectors.Vec3(center), TiltVectors.Vec3(tip), size, TiltVectors.ArrowStart); ok {
			out = append(out, a)
		}
	}
	return out
}
