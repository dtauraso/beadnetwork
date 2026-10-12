package Node

import (
	"github.com/dtauraso/beadnetwork/Categories/Node/TiltVectors"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
)

type TipCopy struct {
	At  Vec3
	Row int32
}

var TipCopyValueNames = []string{"copyX", "copyY", "copyZ", "copyRow"}

func TipPoint(center Vec3, vec polar.Polar) Vec3 {
	return center.Add(Vec3(polar.Polar2cart(vec)))
}

func startArrows(g NodeGeom, center Vec3, starts []TiltVectors.StartVector) []TiltVectors.TiltArrow {
	size := NodeRadius(g.Kind)
	var out []TiltVectors.TiltArrow
	for _, s := range starts {
		tip := TipPoint(center, s.Vec)
		if a, ok := TiltVectors.ArrowBetween(TiltVectors.Vec3(center), TiltVectors.Vec3(tip), size, TiltVectors.ArrowStart); ok {
			out = append(out, a)
		}
	}
	return out
}

func TipCopies(center Vec3, starts []TiltVectors.StartVector) []TipCopy {
	out := make([]TipCopy, 0, len(starts))
	for _, s := range starts {
		out = append(out, TipCopy{At: TipPoint(center, s.Vec), Row: s.ToRow})
	}
	return out
}

type tipCopyBlock interface {
	U8(name string, v uint8)
	F32(name string, v float32)
}

func WriteTipCopyValues(w tipCopyBlock, copies []TipCopy) {
	for _, c := range copies {
		w.F32("copyX", float32(c.At.X))
		w.F32("copyY", float32(c.At.Y))
		w.F32("copyZ", float32(c.At.Z))
		w.U8("copyRow", uint8(c.Row))
	}
}
