package runtopology

import (
	"fmt"
	"os"

	"github.com/dtauraso/wirefold/nodes/rowevent"

	SF "github.com/dtauraso/wirefold/Buffer/streamframe"
	W "github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
)

func wireEdgeStreams(streamFDs SF.StreamFDs, md *W.MoveDispatch) {
	if _, edgeFDsWired := streamFDs[SF.StreamKindEdge]; !edgeFDsWired {
		if n := len(md.GS.EdgeSeedsFn()); n > 0 {
			fmt.Fprintf(os.Stderr,
				"stream-fd mismatch: topology loaded %d edges but WIREFOLD_STREAM_FDS carries no %q entry; "+
					"every edgeMover's stream stays nil, so NO EDGES will be drawn. If the editor was open "+
					"across a rebuild, run \"Developer: Reload Window\" — reopening the file restarts only the "+
					"webview, not the extension host that allocates these fds.\n",
				n, SF.StreamKindEdge)
		}
	}
	if edgeBase, ok := streamFDs[SF.StreamKindEdge]; ok {

		md.Sw.SetEdgeStreams(md.GS.EdgeSeeds, md.MR.Edges(), md.MR.NodeGeoms(), edgeBase, md.RT.NodeRowFor,
			func(tick uint32, sx, sy, sz, ex, ey, ez float32, srcNodeRow, dstNodeRow int32, deltaR float32, label string, events []rowevent.RowEvent) []byte {
				return SF.BuildEdgeStreamFrame(tick, sx, sy, sz, ex, ey, ez, srcNodeRow, dstNodeRow, deltaR, label, toStreamEvents(events))
			})
	}
}
