package Node

import (
	"math"

	"github.com/dtauraso/beadnetwork/Categories/Node/TiltVectors"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
)

func HeadPoint(center Vec3, vec polar.Polar) Vec3 {
	return center.Add(Vec3(polar.Polar2cart(vec)))
}

func startArrows(g NodeGeom, center Vec3, starts []polar.Polar) []TiltVectors.TiltArrow {
	size := NodeRadius(g.Kind)
	var out []TiltVectors.TiltArrow
	for _, vec := range starts {
		head := HeadPoint(center, vec)
		if a, ok := TiltVectors.ArrowBetween(TiltVectors.Vec3(center), TiltVectors.Vec3(head), size, TiltVectors.ArrowStart); ok {
			out = append(out, a)
		}
	}
	return out
}

const arcSegments = 24

func startArcs(kind string, center Vec3, starts []polar.Polar) []TickMark {
	outer := NodeTorusOuterR(kind)
	var out []TickMark
	for _, vec := range starts {
		if vec.R < 1e-9 {
			continue
		}
		theta := math.Mod(vec.Theta+2*math.Pi, 2*math.Pi)
		out = append(out, arc(center, func(t float64) Vec3 {
			return Vec3(polar.Polar2cart(polar.Polar{R: 1.6 * outer, Phi: t * vec.Phi, Theta: vec.Theta}))
		})...)
		out = append(out, arc(center, func(t float64) Vec3 {
			return Vec3{X: 1.3 * outer * math.Cos(t*theta), Z: 1.3 * outer * math.Sin(t*theta)}
		})...)
	}
	return out
}

func arc(center Vec3, at func(t float64) Vec3) []TickMark {
	out := make([]TickMark, 0, arcSegments+2)
	prev := at(0)
	for i := 1; i <= arcSegments; i++ {
		p := at(float64(i) / arcSegments)
		out = append(out, TickMark{From: center.Add(prev), To: center.Add(p)})
		prev = p
	}
	end, back := at(1), at(1-1.0/arcSegments)
	along := end.Sub(back)
	if along.Length() < 1e-9 {
		return out
	}
	along = along.Normalize().Scale(end.Length() * 0.12)
	side := end.Normalize().Scale(along.Length() * 0.6)
	out = append(out,
		TickMark{From: center.Add(end), To: center.Add(end.Sub(along).Add(side))},
		TickMark{From: center.Add(end), To: center.Add(end.Sub(along).Sub(side))})
	return out
}
