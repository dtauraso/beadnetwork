package streamframe

import (
	"encoding/binary"

	B "github.com/dtauraso/wirefold/tools/topology-vscode/Buffer"
)

type EdgeBead struct {
	X, Y, Z    float32
	Value      int32
	EdgeRow    int32
	RingMatrix [16]float32
}

func BuildBeadStreamFrame(tick uint32, nodeRow int32, events []StreamEvent) []byte {
	buf := make([]byte, B.BufBeadStreamFrameHeaderSize)
	binary.LittleEndian.PutUint32(buf[0:], tick)
	binary.LittleEndian.PutUint32(buf[4:], uint32(nodeRow))
	return append(buf, BuildEventsSection(events)...)
}
