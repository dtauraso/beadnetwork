package streamframe

import (
	"encoding/binary"
	"fmt"

	B "github.com/dtauraso/wirefold/tools/topology-vscode/Buffer"
)

func BuildEdgeStreamFrame(tick uint32, sx, sy, sz, ex, ey, ez float32, srcNodeRow, dstNodeRow int32, deltaR float32, dragActive uint8, label string, events []StreamEvent) []byte {
	labelBytes := []byte(label)
	size := B.BufEdgeStreamFrameHeaderSize + B.BufEdgeStride + len(labelBytes)
	buf := make([]byte, size)
	off := 0
	binary.LittleEndian.PutUint32(buf[off:], tick)
	off += 4
	binary.LittleEndian.PutUint32(buf[off:], uint32(len(labelBytes)))
	off += 4

	B.SetEdgeRow(buf[off:off+B.BufEdgeStride], 0, sx, sy, sz, ex, ey, ez, srcNodeRow, dstNodeRow, deltaR, dragActive, 0, uint32(len(labelBytes)))
	off += B.BufEdgeStride
	copy(buf[off:off+len(labelBytes)], labelBytes)
	off += len(labelBytes)

	if off != size {
		panic(fmt.Sprintf(
			"BuildEdgeStreamFrame: packed %d bytes for edge %q but allocated %d — the section walk and the size formula disagree",
			off, label, size))
	}
	return append(buf, BuildEventsSection(events)...)
}
