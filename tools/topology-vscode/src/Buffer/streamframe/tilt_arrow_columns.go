package streamframe

import (
	"encoding/binary"
	"math"

	B "github.com/dtauraso/wirefold/tools/topology-vscode/src/Buffer"
	"github.com/dtauraso/wirefold/tools/topology-vscode/src/Buffer/colstream"
)

func WriteTiltArrowColumns(c *colstream.ColumnSet, arrows []TiltArrow) {
	if c == nil {
		return
	}
	n := len(arrows)

	received := make([]byte, 0, n)
	for _, a := range arrows {
		received = append(received, a.Received)
	}
	c.SetBytes(B.ColStreamTiltArrowReceived, received)

	for m := 0; m < 16; m++ {
		shaft := make([]byte, 0, n*4)
		head := make([]byte, 0, n*4)
		for _, a := range arrows {
			shaft = binary.LittleEndian.AppendUint32(shaft, math.Float32bits(a.Shaft[m]))
			head = binary.LittleEndian.AppendUint32(head, math.Float32bits(a.Head[m]))
		}
		c.SetBytes(tiltArrowShaftCols[m], shaft)
		c.SetBytes(tiltArrowHeadCols[m], head)
	}
}

var tiltArrowShaftCols = [16]int{
	B.ColStreamTiltArrowShaftM0, B.ColStreamTiltArrowShaftM1, B.ColStreamTiltArrowShaftM2, B.ColStreamTiltArrowShaftM3,
	B.ColStreamTiltArrowShaftM4, B.ColStreamTiltArrowShaftM5, B.ColStreamTiltArrowShaftM6, B.ColStreamTiltArrowShaftM7,
	B.ColStreamTiltArrowShaftM8, B.ColStreamTiltArrowShaftM9, B.ColStreamTiltArrowShaftM10, B.ColStreamTiltArrowShaftM11,
	B.ColStreamTiltArrowShaftM12, B.ColStreamTiltArrowShaftM13, B.ColStreamTiltArrowShaftM14, B.ColStreamTiltArrowShaftM15,
}

var tiltArrowHeadCols = [16]int{
	B.ColStreamTiltArrowHeadM0, B.ColStreamTiltArrowHeadM1, B.ColStreamTiltArrowHeadM2, B.ColStreamTiltArrowHeadM3,
	B.ColStreamTiltArrowHeadM4, B.ColStreamTiltArrowHeadM5, B.ColStreamTiltArrowHeadM6, B.ColStreamTiltArrowHeadM7,
	B.ColStreamTiltArrowHeadM8, B.ColStreamTiltArrowHeadM9, B.ColStreamTiltArrowHeadM10, B.ColStreamTiltArrowHeadM11,
	B.ColStreamTiltArrowHeadM12, B.ColStreamTiltArrowHeadM13, B.ColStreamTiltArrowHeadM14, B.ColStreamTiltArrowHeadM15,
}
