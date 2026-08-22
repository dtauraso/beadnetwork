package edge

import (
	"fmt"
	"os"
)

type EdgeSink func(tick uint32, edgeRow int32, sx, sy, sz, ex, ey, ez float32,
	srcNodeRow, dstNodeRow int32, deltaR float32, dragActive uint8, label string)

func NewEdgeSink(sceneRoot string, rows int) EdgeSink {
	values := make([]*ValueWriter, rows)
	for row := range values {
		values[row] = NewValueWriter(sceneRoot, row)
	}
	return func(tick uint32, edgeRow int32, sx, sy, sz, ex, ey, ez float32,
		srcNodeRow, dstNodeRow int32, deltaR float32, dragActive uint8, label string) {
		if int(edgeRow) < len(values) {
			if err := values[edgeRow].Write(sx, sy, sz, ex, ey, ez, srcNodeRow, dstNodeRow, deltaR, dragActive, label); err != nil {
				fmt.Fprintf(os.Stderr, "edge values write (row %d): %v\n", edgeRow, err)
			}
		}
	}
}
