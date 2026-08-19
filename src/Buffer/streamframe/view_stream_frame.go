package streamframe

import (
	"encoding/binary"
	"github.com/dtauraso/wirefold/src/Node/rowevent"

	B "github.com/dtauraso/wirefold/src/Buffer"
)

func BuildViewStreamFrame(tick uint32,
	tabNames []string, tabSelected uint16,
	events []rowevent.RowEvent,
) []byte {
	buf := make([]byte, B.BufViewFrameHeaderSize)
	binary.LittleEndian.PutUint32(buf[0:], tick)
	binary.LittleEndian.PutUint32(buf[4:], B.BufLayoutFingerprintHash)

	buf = append(buf, BuildSceneTabsSection(tabNames, tabSelected)...)
	return append(buf, BuildEventsSection(events)...)
}

func BuildSceneTabsSection(names []string, selected uint16) []byte {
	size := 4
	for _, n := range names {
		size += 2 + len(n)
	}
	buf := make([]byte, size)
	binary.LittleEndian.PutUint16(buf[0:], uint16(len(names)))
	binary.LittleEndian.PutUint16(buf[2:], selected)
	off := 4
	for _, n := range names {
		binary.LittleEndian.PutUint16(buf[off:], uint16(len(n)))
		off += 2
		off += copy(buf[off:], n)
	}
	return buf
}
