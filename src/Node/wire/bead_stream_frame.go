package wire

import (
	"encoding/binary"
	"github.com/dtauraso/wirefold/src/Node/rowevent"

	B "github.com/dtauraso/wirefold/src/schema/buffer-layout"
)

func BuildBeadStreamFrame(tick uint32, nodeRow int32, events []rowevent.RowEvent) []byte {
	buf := make([]byte, B.BufBeadStreamFrameHeaderSize)
	binary.LittleEndian.PutUint32(buf[0:], tick)
	binary.LittleEndian.PutUint32(buf[4:], uint32(nodeRow))
	return append(buf, B.BuildEventsSection(events)...)
}
