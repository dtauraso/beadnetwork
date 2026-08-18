package streamframe

import (
	"encoding/binary"
	"fmt"

	B "github.com/dtauraso/wirefold/tools/topology-vscode/Buffer"
)

type EdgeBead struct {
	X, Y, Z    float32
	Value      int32
	EdgeRow    int32
	RingMatrix [16]float32
}

func BuildBeadStreamFrame(tick uint32, nodeRow int32, beads []EdgeBead, events []StreamEvent) []byte {
	size := B.BufBeadStreamFrameHeaderSize + len(beads)*B.BufEdgeBeadStride
	buf := make([]byte, size)
	off := 0
	binary.LittleEndian.PutUint32(buf[off:], tick)
	off += 4
	binary.LittleEndian.PutUint32(buf[off:], uint32(nodeRow))
	off += 4
	binary.LittleEndian.PutUint32(buf[off:], uint32(len(beads)))
	off += 4

	for i, b := range beads {
		rowOff := off + i*B.BufEdgeBeadStride
		m := b.RingMatrix
		B.SetEdgeBeadRow(buf[rowOff:rowOff+B.BufEdgeBeadStride], 0, b.X, b.Y, b.Z, b.Value, b.EdgeRow,
			m[0], m[1], m[2], m[3], m[4], m[5], m[6], m[7],
			m[8], m[9], m[10], m[11], m[12], m[13], m[14], m[15])
	}
	off += len(beads) * B.BufEdgeBeadStride

	if off != size {
		panic(fmt.Sprintf(
			"BuildBeadStreamFrame: packed %d bytes for node row %d but allocated %d — the section walk and the size formula disagree",
			off, nodeRow, size))
	}
	return append(buf, BuildEventsSection(events)...)
}
