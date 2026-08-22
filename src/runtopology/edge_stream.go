package runtopology

import (
	"fmt"
	"os"

	W "github.com/dtauraso/wirefold/src/Input/dispatch"
	EdgeB "github.com/dtauraso/wirefold/src/Node/Edge"
	T "github.com/dtauraso/wirefold/src/Trace"
)

func wireEdgeStreams(md *W.MoveDispatch, sceneRoot string) {
	edgeValues := make([]*EdgeB.ValueWriter, len(md.RT.EdgeRowTable))
	for row := range edgeValues {
		edgeValues[row] = EdgeB.NewValueWriter(sceneRoot, row)
	}

	md.Sw.SetEdgeStreams(md.GS.EdgeSeeds, md.MR.Edges(), md.MR.NodeGeoms(), md.RT.NodeRowFor,
		func(tick uint32, edgeRow int32, sx, sy, sz, ex, ey, ez float32, srcNodeRow, dstNodeRow int32, deltaR float32, dragActive uint8, label string, events []T.RowEvent) {
			if int(edgeRow) < len(edgeValues) {
				if err := edgeValues[edgeRow].Write(sx, sy, sz, ex, ey, ez, srcNodeRow, dstNodeRow, deltaR, dragActive, label); err != nil {
					fmt.Fprintf(os.Stderr, "edge values write (row %d): %v\n", edgeRow, err)
				}
			}
			T.NewLog(T.OwnerEdge, edgeRow).Append(events)
		})
}
