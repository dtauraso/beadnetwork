package streamframe

import (
	"encoding/binary"
	"fmt"

	B "github.com/dtauraso/wirefold/Buffer"
)

type EdgeBead struct {
	X, Y, Z float32
	Value   int32
}

func BuildEdgeStreamFrame(tick uint32, sx, sy, sz, ex, ey, ez float32, srcNodeRow int32, label string, beads []EdgeBead, events []StreamEvent) []byte {
	labelBytes := []byte(label)
	size := B.BufEdgeStreamFrameHeaderSize + B.BufEdgeStride + len(labelBytes) + len(beads)*B.BufEdgeBeadStride
	buf := make([]byte, size)
	off := 0
	binary.LittleEndian.PutUint32(buf[off:], tick)
	off += 4
	binary.LittleEndian.PutUint32(buf[off:], uint32(len(labelBytes)))
	off += 4
	binary.LittleEndian.PutUint32(buf[off:], uint32(len(beads)))
	off += 4

	B.SetEdgeRow(buf[off:off+B.BufEdgeStride], 0, sx, sy, sz, ex, ey, ez, srcNodeRow, 0, uint32(len(labelBytes)))
	off += B.BufEdgeStride
	copy(buf[off:off+len(labelBytes)], labelBytes)
	off += len(labelBytes)

	for i, b := range beads {
		rowOff := off + i*B.BufEdgeBeadStride
		B.SetEdgeBeadRow(buf[rowOff:rowOff+B.BufEdgeBeadStride], 0, b.X, b.Y, b.Z, b.Value)
	}
	off += len(beads) * B.BufEdgeBeadStride

	if off != size {
		panic(fmt.Sprintf(
			"BuildEdgeStreamFrame: packed %d bytes for edge %q but allocated %d — the section walk and the size formula disagree",
			off, label, size))
	}
	return append(buf, BuildEventsSection(events)...)
}
