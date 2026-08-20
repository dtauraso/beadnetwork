package runtopology

import (
	"fmt"
	"os"

	B "github.com/dtauraso/wirefold/src/schema/buffer-layout"

	EdgeB "github.com/dtauraso/wirefold/src/Node/Edge"
	W "github.com/dtauraso/wirefold/src/Input/dispatch"
	SW "github.com/dtauraso/wirefold/src/runtopology/streamwire"
	"github.com/dtauraso/wirefold/src/schema/buffer-layout/colstream"
)

func wireEdgeStreams(streamFDs SW.StreamFDs, md *W.MoveDispatch) {
	cols := SW.NewColumnStreams(streamFDs, len(md.RT.NodeRowTable), len(md.RT.EdgeRowTable))

	edgeSets := map[int32]*colstream.ColumnSet{}
	edgeCols := func(row int32) *colstream.ColumnSet {
		if s, ok := edgeSets[row]; ok {
			return s
		}
		s := cols.EdgeColumns(int(row))
		edgeSets[row] = s
		return s
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
			func(tick uint32, edgeRow int32, sx, sy, sz, ex, ey, ez float32, srcNodeRow, dstNodeRow int32, deltaR float32, dragActive uint8, label string, events []B.RowEvent) []byte {
				EdgeB.WriteEdgeColumns(edgeCols(edgeRow), sx, sy, sz, ex, ey, ez, srcNodeRow, dstNodeRow, deltaR, dragActive, label)
				return EdgeB.BuildEdgeStreamFrame(tick, events)
			})
	}
}
