package Edge

import (
	"encoding/binary"
	"github.com/dtauraso/wirefold/src/Node/rowevent"

	T "github.com/dtauraso/wirefold/src/Trace"
	B "github.com/dtauraso/wirefold/src/schema/buffer-layout"
	"github.com/dtauraso/wirefold/src/schema/buffer-layout/colstream"
)

func WriteEdgeColumns(c *colstream.ColumnSet,
	sx, sy, sz, ex, ey, ez float32,
	srcNodeRow, dstNodeRow int32, deltaR float32, dragActive uint8, label string,
) {
	if c == nil {
		return
	}
	c.SetF32(B.ColStreamEdgeSX, sx)
	c.SetF32(B.ColStreamEdgeSY, sy)
	c.SetF32(B.ColStreamEdgeSZ, sz)
	c.SetF32(B.ColStreamEdgeEX, ex)
	c.SetF32(B.ColStreamEdgeEY, ey)
	c.SetF32(B.ColStreamEdgeEZ, ez)
	c.SetI32(B.ColStreamEdgeSrcNodeRow, srcNodeRow)
	c.SetI32(B.ColStreamEdgeDstNodeRow, dstNodeRow)
	c.SetF32(B.ColStreamEdgeDeltaR, deltaR)
	c.SetU8(B.ColStreamEdgeDragActive, dragActive)
	c.SetBytes(B.ColStreamEdgeLabel, []byte(label))
}

func BuildEdgeStreamFrame(tick uint32, events []rowevent.RowEvent) []byte {
	buf := make([]byte, B.BufEdgeStreamFrameHeaderSize)
	binary.LittleEndian.PutUint32(buf[0:], tick)
	return append(buf, T.BuildEventsSection(events)...)
}
