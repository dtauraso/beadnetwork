package ChannelVectors

type ChannelVector struct {
	Shaft [16]float32
	Head  [16]float32
}

type block interface {
	F32(name string, v float32)
}

func WriteChannelVectorValues(w block, vectors []ChannelVector) {
	if w == nil {
		return
	}
	for m := 0; m < 16; m++ {
		shaft := ShaftName(m)
		head := HeadName(m)
		for _, v := range vectors {
			w.F32(shaft, v.Shaft[m])
			w.F32(head, v.Head[m])
		}
	}
}
