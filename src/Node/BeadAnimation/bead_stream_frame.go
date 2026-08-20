package beadanimation

import (
	"encoding/binary"

	B "github.com/dtauraso/wirefold/src/Buffer"
)

func BuildBeadStreamFrame(tick uint32, nodeRow int32, events []B.RowEvent) []byte {
	buf := make([]byte, B.BufBeadStreamFrameHeaderSize)
	binary.LittleEndian.PutUint32(buf[0:], tick)
	binary.LittleEndian.PutUint32(buf[4:], uint32(nodeRow))
	return append(buf, B.BuildEventsSection(events)...)
}
