package Node

import (
	"math"

	"github.com/dtauraso/beadnetwork/Categories/Node/TiltVectors"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Camera"
)

type TickMark struct {
	From, To Vec3
}

var TickValueNames = []string{"tickX0", "tickY0", "tickZ0", "tickX1", "tickY1", "tickZ1"}

func TickMarks(center Vec3, kind string, count int32) []TickMark {
	if count <= 0 {
		return nil
	}
	reach := NodeRadius(kind) * (1 - ShadingParamNodeRingTubeRatio)
	out := make([]TickMark, count)
	for i := range out {
		phi := 2 * math.Pi * float64(i) / float64(count)
		dir := Vec3(Camera.AnglesToWorldOffset(1, phi, TiltVectors.ArrowRingDiskTheta).Normalize())
		out[i] = TickMark{From: center, To: center.Add(dir.Scale(reach))}
	}
	return out
}

type tickBlock interface {
	F32(name string, v float32)
}

func WriteTickValues(w tickBlock, ticks []TickMark) {
	for _, t := range ticks {
		w.F32("tickX0", float32(t.From.X))
		w.F32("tickY0", float32(t.From.Y))
		w.F32("tickZ0", float32(t.From.Z))
		w.F32("tickX1", float32(t.To.X))
		w.F32("tickY1", float32(t.To.Y))
		w.F32("tickZ1", float32(t.To.Z))
	}
}
