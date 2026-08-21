package edge

import (
	"encoding/binary"

	B "github.com/dtauraso/wirefold/src/Buffer"
)

func BuildEdgeStreamFrame(tick uint32, events []B.RowEvent) []byte {
	buf := make([]byte, B.BufEdgeStreamFrameHeaderSize)
	binary.LittleEndian.PutUint32(buf[0:], tick)
	return append(buf, B.BuildEventsSection(events)...)
}
