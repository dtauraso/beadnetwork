package streamframe

import (
	"encoding/binary"
	"math"

	B "github.com/dtauraso/wirefold/tools/topology-vscode/Buffer"
	"github.com/dtauraso/wirefold/tools/topology-vscode/Buffer/colstream"
)

func WriteTiltArrowColumns(c *colstream.ColumnSet, arrows []TiltArrow) {
	if c == nil {
		return
	}
	n := len(arrows)

	received := make([]byte, 0, n)
	shaft := make([]byte, 0, n*16*4)
	head := make([]byte, 0, n*16*4)
	for _, a := range arrows {
		received = append(received, a.Received)
		for m := 0; m < 16; m++ {
			shaft = binary.LittleEndian.AppendUint32(shaft, math.Float32bits(a.Shaft[m]))
			head = binary.LittleEndian.AppendUint32(head, math.Float32bits(a.Head[m]))
		}
	}
	c.SetBytes(B.ColStreamTiltArrowReceived, received)
	c.SetBytes(B.ColStreamTiltArrowShaft, shaft)
	c.SetBytes(B.ColStreamTiltArrowHead, head)
}
