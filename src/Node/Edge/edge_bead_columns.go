package edge

import (
	bead "github.com/dtauraso/wirefold/src/Ring/Bead"
)

type EdgeBead struct {
	X, Y, Z    float32
	Value      int32
	EdgeRow    int32
	RingMatrix [16]float32
}

func WriteEdgeBeadValues(w *bead.ValueWriter, beads []EdgeBead) error {
	if w == nil {
		return nil
	}
	w.Begin()
	for _, b := range beads {
		w.F32("x", b.X)
		w.F32("y", b.Y)
		w.F32("z", b.Z)
		w.I32("value", b.Value)
	}
	for m := 0; m < 16; m++ {
		name := bead.RingName(m)
		for _, b := range beads {
			w.F32(name, b.RingMatrix[m])
		}
	}
	return w.Flush()
}
