package SceneVectors

import (
	"encoding/binary"
	"math"

	B "github.com/dtauraso/wirefold/src/schema/buffer-layout"
	"github.com/dtauraso/wirefold/src/schema/buffer-layout/colstream"
)

type ChannelVector struct {
	Shaft [16]float32
	Head  [16]float32
}

func WriteChannelVectorColumns(c *colstream.ColumnSet, vectors []ChannelVector) {
	if c == nil {
		return
	}
	n := len(vectors)

	for m := 0; m < 16; m++ {
		shaft := make([]byte, 0, n*4)
		head := make([]byte, 0, n*4)
		for _, v := range vectors {
			shaft = binary.LittleEndian.AppendUint32(shaft, math.Float32bits(v.Shaft[m]))
			head = binary.LittleEndian.AppendUint32(head, math.Float32bits(v.Head[m]))
		}
		c.SetBytes(channelVectorShaftCols[m], shaft)
		c.SetBytes(channelVectorHeadCols[m], head)
	}
}

var channelVectorShaftCols = [16]int{
	B.ColStreamChannelVectorShaftM0, B.ColStreamChannelVectorShaftM1, B.ColStreamChannelVectorShaftM2, B.ColStreamChannelVectorShaftM3,
	B.ColStreamChannelVectorShaftM4, B.ColStreamChannelVectorShaftM5, B.ColStreamChannelVectorShaftM6, B.ColStreamChannelVectorShaftM7,
	B.ColStreamChannelVectorShaftM8, B.ColStreamChannelVectorShaftM9, B.ColStreamChannelVectorShaftM10, B.ColStreamChannelVectorShaftM11,
	B.ColStreamChannelVectorShaftM12, B.ColStreamChannelVectorShaftM13, B.ColStreamChannelVectorShaftM14, B.ColStreamChannelVectorShaftM15,
}

var channelVectorHeadCols = [16]int{
	B.ColStreamChannelVectorHeadM0, B.ColStreamChannelVectorHeadM1, B.ColStreamChannelVectorHeadM2, B.ColStreamChannelVectorHeadM3,
	B.ColStreamChannelVectorHeadM4, B.ColStreamChannelVectorHeadM5, B.ColStreamChannelVectorHeadM6, B.ColStreamChannelVectorHeadM7,
	B.ColStreamChannelVectorHeadM8, B.ColStreamChannelVectorHeadM9, B.ColStreamChannelVectorHeadM10, B.ColStreamChannelVectorHeadM11,
	B.ColStreamChannelVectorHeadM12, B.ColStreamChannelVectorHeadM13, B.ColStreamChannelVectorHeadM14, B.ColStreamChannelVectorHeadM15,
}
