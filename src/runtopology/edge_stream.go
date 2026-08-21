package runtopology

import (
	T "github.com/dtauraso/wirefold/src/Trace"
	"fmt"
	"os"


	W "github.com/dtauraso/wirefold/src/Input/dispatch"
	EdgeB "github.com/dtauraso/wirefold/src/Node/Edge"
	SW "github.com/dtauraso/wirefold/src/runtopology/streamwire"
)

func wireEdgeStreams(streamFDs SW.StreamFDs, md *W.MoveDispatch, sceneRoot string) {

	edgeValues := make([]*EdgeB.ValueWriter, len(md.RT.EdgeRowTable))
	for row := range edgeValues {
		edgeValues[row] = EdgeB.NewValueWriter(sceneRoot, row)
	}
	if _, edgeFDsWired := streamFDs[SW.StreamKindEdge]; !edgeFDsWired {
		if n := len(md.GS.EdgeSeedsFn()); n > 0 {
			fmt.Fprintf(os.Stderr,
				"stream-fd mismatch: topology loaded %d edges but WIREFOLD_STREAM_FDS carries no %q entry; "+
					"every edgeMover's stream stays nil, so NO EDGES will be drawn. If the editor was open "+
					"across a rebuild, run \"Developer: Reload Window\" — reopening the file restarts only the "+
					"webview, not the extension host that allocates these fds.\n",
				n, SW.StreamKindEdge)
		}
	}
	if edgeBase, ok := streamFDs[SW.StreamKindEdge]; ok {

		md.Sw.SetEdgeStreams(md.GS.EdgeSeeds, md.MR.Edges(), md.MR.NodeGeoms(), edgeBase, md.RT.NodeRowFor,
			func(tick uint32, edgeRow int32, sx, sy, sz, ex, ey, ez float32, srcNodeRow, dstNodeRow int32, deltaR float32, dragActive uint8, label string, events []T.RowEvent) []byte {
				if int(edgeRow) < len(edgeValues) {
					if err := edgeValues[edgeRow].Write(sx, sy, sz, ex, ey, ez, srcNodeRow, dstNodeRow, deltaR, dragActive, label); err != nil {
						fmt.Fprintf(os.Stderr, "edge values write (row %d): %v\n", edgeRow, err)
					}
				}
				T.NewLog(T.OwnerEdge, edgeRow).Append(events)
				return EdgeB.BuildEdgeStreamFrame(tick)
			})
	}
}
