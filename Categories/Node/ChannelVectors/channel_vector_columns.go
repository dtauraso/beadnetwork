package ChannelVectors

type ChannelVector struct {
	Shaft [16]float32
	Head  [16]float32
}

func WriteChannelVectorValues(w *ValueWriter, vectors []ChannelVector) error {
	if w == nil {
		return nil
	}
	w.Begin()
	for m := 0; m < 16; m++ {
		shaft := ShaftName(m)
		head := HeadName(m)
		for _, v := range vectors {
			w.F32(shaft, v.Shaft[m])
			w.F32(head, v.Head[m])
		}
	}
	return w.Flush()
}
