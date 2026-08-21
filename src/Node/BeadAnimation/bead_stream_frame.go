package beadanimation

import (
	T "github.com/dtauraso/wirefold/src/Trace"
	"encoding/binary"

	B "github.com/dtauraso/wirefold/src/Buffer"
)

func BuildBeadStreamFrame(tick uint32, nodeRow int32, events []T.RowEvent) []byte {
	buf := make([]byte, B.BufBeadStreamFrameHeaderSize)
	binary.LittleEndian.PutUint32(buf[0:], tick)
	binary.LittleEndian.PutUint32(buf[4:], uint32(nodeRow))

	T.NewLog(T.OwnerBead, nodeRow).Append(events)
	return buf
}
