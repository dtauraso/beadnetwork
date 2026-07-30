// headless_settled_geometry_test.go — drives the REAL compiled binary headlessly
// against the real topology/ dir and asserts that the SETTLED per-owner NODE and EDGE
// stream frames already have REAL geometry in every row: every edge's segment is
// non-degenerate (start != end, not the 0,0,0->0,0,0 placeholder a bare address-only seed
// would produce). This is the row-completeness half of "rows come from the diagram"
// (CLAUDE.md/MODEL.md): a row is prefilled with the diagram's own state, not an empty row
// waiting on a node goroutine to fill it in later.
//
// See headless_stream_helpers_test.go for the spawn/cleanup pattern this reuses; NEVER run
// the sim in the foreground (memory/feedback_no_foreground_sim_runs.md).

package main

import (
	"os"
	"testing"

	B "github.com/dtauraso/wirefold/Buffer"
)

// TestHeadlessSettledFramesHaveRealGeometry asserts that the settled per-owner node/edge
// stream frames already have non-degenerate edge segments — proving the row-seed carries
// the diagram's REAL state, not an empty/placeholder row waiting on a later per-node emit.
// A port carries no geometry of its own any more (docs/channels-not-ports.md), so the
// edge's segment is read straight off its own frame's SX..EZ columns — no port-row
// indirection through a node frame's own Port block (that block is gone).
func TestHeadlessSettledFramesHaveRealGeometry(t *testing.T) {
	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	binPath := buildHeadlessBinary(t, repoRoot, "wirefold-headless-first-frame-geometry-test")
	ds := spawnDedicatedAllStreams(t, binPath, repoRoot)
	if ds.edgeN == 0 {
		t.Fatal("topology/ is expected to have edges; test cannot assert non-degeneracy without any")
	}

	nodeFrames := readLastFrames(t, ds.nodeReads, "node")
	edgeFrames := readLastFrames(t, ds.edgeReads, "edge")

	if len(nodeFrames) != len(ds.nodeIDs) {
		t.Fatalf("nodeFrames count %d != topology dir count %d", len(nodeFrames), len(ds.nodeIDs))
	}

	// Every edge row's segment must be non-degenerate: start != end. A bare address-only
	// seed (the bug this test targets) would leave every edge at 0,0,0 -> 0,0,0.
	for row, edgeFrame := range edgeFrames {
		if len(edgeFrame) < 4+B.BufEdgeStride+4 {
			t.Fatalf("edge row %d: frame too short (%d bytes)", row, len(edgeFrame))
		}
		edgeRowOff := 4
		sx := readF32(edgeFrame, edgeRowOff+B.BufEdgeColSX)
		sy := readF32(edgeFrame, edgeRowOff+B.BufEdgeColSY)
		sz := readF32(edgeFrame, edgeRowOff+B.BufEdgeColSZ)
		ex := readF32(edgeFrame, edgeRowOff+B.BufEdgeColEX)
		ey := readF32(edgeFrame, edgeRowOff+B.BufEdgeColEY)
		ez := readF32(edgeFrame, edgeRowOff+B.BufEdgeColEZ)
		if sx == 0 && sy == 0 && sz == 0 && ex == 0 && ey == 0 && ez == 0 {
			t.Fatalf("edge row %d: segment is the degenerate placeholder 0,0,0 -> 0,0,0", row)
		}
		if sx == ex && sy == ey && sz == ez {
			t.Fatalf("edge row %d: start == end (%v == %v) — degenerate zero-length segment", row, [3]float32{sx, sy, sz}, [3]float32{ex, ey, ez})
		}
	}
}
