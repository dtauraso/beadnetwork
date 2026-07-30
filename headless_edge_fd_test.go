// headless_edge_fd_test.go — drives the REAL compiled binary headlessly and proves the
// per-edge dedicated-stream migration end-to-end (memory/feedback_no_single_writer_bridge.md,
// Buffer/stream_fds.go's StreamKindEdge): with every per-owner fd wired (mandatory — no
// fallback path left, memory/feedback_no_single_writer_bridge.md's final step), every edge's own combined
// frame (Edge fields + its wire's live beads) arrives on its OWN fd.
//
// See headless_stream_helpers_test.go for the spawn/cleanup pattern this reuses; NEVER run
// the sim in the foreground (memory/feedback_no_foreground_sim_runs.md).
package main

import (
	"os"
	"testing"

	B "github.com/dtauraso/wirefold/Buffer"
)

// TestHeadlessEdgeFdDedicatedStream proves each edge's combined frame (Edge fields + its
// wire's own live beads) arrives on its OWN fd with resolvable geometry.
func TestHeadlessEdgeFdDedicatedStream(t *testing.T) {
	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	binPath := buildHeadlessBinary(t, repoRoot, "wirefold-headless-edge-fd-test")
	ds := spawnDedicatedAllStreams(t, binPath, repoRoot)
	if ds.edgeN == 0 {
		t.Fatal("topology/nodes/*/edges has 0 edges — test cannot assert per-edge frames without any")
	}

	edgeFrames := readLastFrames(t, ds.edgeReads, "edge")

	for row, frame := range edgeFrames {
		// Combined frame layout (Buffer.BuildEdgeStreamFrame): [tick:u32] + one
		// BufEdgeStride row (SX..EZ/Selected, EdgeLabelOff=0/Len) + label bytes. SX..EZ is
		// the edge's own SEGMENT, node surface to node surface (docs/channels-not-ports.md
		// — there is no port row to reference any more).
		if len(frame) < 4+B.BufEdgeStride+4 {
			t.Fatalf("edge row %d: frame too short (%d bytes) to hold [tick][EdgeRow][beadCount]", row, len(frame))
		}
		edgeRowOff := 4
		sx := readF32(frame, edgeRowOff+B.BufEdgeColSX)
		sy := readF32(frame, edgeRowOff+B.BufEdgeColSY)
		sz := readF32(frame, edgeRowOff+B.BufEdgeColSZ)
		ex := readF32(frame, edgeRowOff+B.BufEdgeColEX)
		ey := readF32(frame, edgeRowOff+B.BufEdgeColEY)
		ez := readF32(frame, edgeRowOff+B.BufEdgeColEZ)
		if sx == 0 && sy == 0 && sz == 0 && ex == 0 && ey == 0 && ez == 0 {
			t.Fatalf("edge row %d: segment is degenerate 0,0,0->0,0,0 — want a resolvable node-surface-to-node-surface segment", row)
		}
		labelLenOff := edgeRowOff + B.BufEdgeStride - 4 // EdgeLabelLen is the last u32 column of the Edge row
		labelLen := readU32(frame, labelLenOff)
		labelStart := edgeRowOff + B.BufEdgeStride
		if labelStart+int(labelLen) > len(frame) {
			t.Fatalf("edge row %d: label overruns frame (labelLen=%d, frameLen=%d)", row, labelLen, len(frame))
		}
		label := string(frame[labelStart : labelStart+int(labelLen)])
		if label == "" {
			t.Fatalf("edge row %d: empty inline label", row)
		}
		// No bead section to walk: the Bead block is gone with the moving bead it carried
		// (docs/beads-are-the-edge.md). This test still earns its keep — it proves each
		// edge's own frame reaches its OWN dedicated fd, which is the invariant in
		// memory/feedback_no_single_writer_bridge.md and is unchanged.
	}
}
