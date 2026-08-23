package TiltVectors

type TiltArrow struct {
	Received uint8
	Shaft    [16]float32
	Head     [16]float32
}

func WriteTiltArrowValues(w *ValueWriter, arrows []TiltArrow) error {
	if w == nil {
		return nil
	}
	w.Begin()
	for _, a := range arrows {
		w.U8("received", a.Received)
	}
	for m := range 16 {
		shaft := ShaftName(m)
		head := HeadName(m)
		for _, a := range arrows {
			w.F32(shaft, a.Shaft[m])
			w.F32(head, a.Head[m])
		}
	}
	return w.Flush()
}
