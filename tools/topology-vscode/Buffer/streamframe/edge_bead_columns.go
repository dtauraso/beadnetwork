package streamframe

import (
	"encoding/binary"
	"math"

	B "github.com/dtauraso/wirefold/tools/topology-vscode/Buffer"
	"github.com/dtauraso/wirefold/tools/topology-vscode/Buffer/colstream"
)

func WriteEdgeBeadColumns(c *colstream.ColumnSet, beads []EdgeBead) {
	if c == nil {
		return
	}
	n := len(beads)

	xs := make([]byte, 0, n*4)
	ys := make([]byte, 0, n*4)
	zs := make([]byte, 0, n*4)
	values := make([]byte, 0, n*4)
	edgeRows := make([]byte, 0, n*4)
	for _, b := range beads {
		xs = binary.LittleEndian.AppendUint32(xs, math.Float32bits(b.X))
		ys = binary.LittleEndian.AppendUint32(ys, math.Float32bits(b.Y))
		zs = binary.LittleEndian.AppendUint32(zs, math.Float32bits(b.Z))
		values = binary.LittleEndian.AppendUint32(values, uint32(b.Value))
		edgeRows = binary.LittleEndian.AppendUint32(edgeRows, uint32(b.EdgeRow))
	}
	rings := make([]byte, 0, n*16*4)
	for _, b := range beads {
		for m := 0; m < 16; m++ {
			rings = binary.LittleEndian.AppendUint32(rings, math.Float32bits(b.RingMatrix[m]))
		}
	}

	c.SetBytes(B.ColStreamEdgeBeadX, xs)
	c.SetBytes(B.ColStreamEdgeBeadY, ys)
	c.SetBytes(B.ColStreamEdgeBeadZ, zs)
	c.SetBytes(B.ColStreamEdgeBeadValue, values)
	c.SetBytes(B.ColStreamEdgeBeadEdgeRow, edgeRows)
	c.SetBytes(B.ColStreamEdgeBeadRingMatrix, rings)
}
