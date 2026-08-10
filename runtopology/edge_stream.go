package runtopology

import (
	"fmt"
	"os"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	SF "github.com/dtauraso/wirefold/Buffer/streamframe"
	W "github.com/dtauraso/wirefold/nodes/Wiring"
)

// wireEdgeStreams reports the edge-fd asymmetry (loaded edges but no "edge" entry) and,
// when the entry IS present, wires every edgeMover to its own dedicated fd. Report first,
// then wire — the same order as before this was a named phase.
//
// The per-edge dedicated stream (memory/feedback_no_single_writer_bridge.md): when
// WIREFOLD_STREAM_FDS carries an "edge" entry, wire every edgeMover to its OWN fd
// (fd = baseFd + edgeRow, edgeRow = the stable seed order — see
// MoveDispatch.SetEdgeStreams). No edges (edgeBase absent) leaves every edgeMover's
// streamOut at its zero value (nil) — there is nothing to stream.
//
// REPORT the asymmetry rather than skipping it silently. A graph that loaded N edges
// but received no edge fds streams nothing, so the editor draws no edges — and without
// this message that is indistinguishable from a broken edge path, sending the reader
// through the code (disk layout, bundle freshness, recent merges) when the cause is
// operational: a VS Code extension host left running across a change, holding older fd
// plumbing. Reopening a file does not restart it; only "Developer: Reload Window" does
// (memory/feedback_two_process_editor_reload.md).
//
// This is the same class runCommand.ts's MAX_EDGE_STREAMS overflow was fixed for in
// 93d2e9b6, and its reasoning applies verbatim: silently disabling every dedicated
// per-edge stream is "the quietest possible failure for the loudest consequence."
// That fix covered the count-too-large case on the TS side; this covers the fds-absent
// case on the Go side, which is the one an operator actually hits.
//
// Not fatal: a deliberate no-fd launch (headless runs, tools with no dedicated pipes)
// is legitimate input, exactly like a large topology. Loud, not dead.
func wireEdgeStreams(streamFDs SF.StreamFDs, md *W.MoveDispatch) {
	if _, edgeFDsWired := streamFDs[SF.StreamKindEdge]; !edgeFDsWired {
		if n := len(md.EdgeSeeds()); n > 0 {
			fmt.Fprintf(os.Stderr,
				"stream-fd mismatch: topology loaded %d edges but WIREFOLD_STREAM_FDS carries no %q entry; "+
					"every edgeMover's stream stays nil, so NO EDGES will be drawn. If the editor was open "+
					"across a rebuild, run \"Developer: Reload Window\" — reopening the file restarts only the "+
					"webview, not the extension host that allocates these fds.\n",
				n, SF.StreamKindEdge)
		}
	}
	if edgeBase, ok := streamFDs[SF.StreamKindEdge]; ok {
		// Edge selection is no longer an injected lookup: each edgeMover owns its OWN
		// selected bit, set via a moveMsgKindSelect message the gesture goroutine sends
		// on select/deselect (MoveDispatch.sendEdgeSelect).
		md.SetEdgeStreams(edgeBase, md.RT.NodeRowFor,
			func(tick uint32, sx, sy, sz, ex, ey, ez float32, selected uint8, label string, events []wire.RowEvent) []byte {
				return SF.BuildEdgeStreamFrame(tick, sx, sy, sz, ex, ey, ez, selected, label, toStreamEvents(events))
			})
	}
}
