// headless_node_row_order_test.go — drives the REAL compiled binary headlessly against
// the real topology/ dir and asserts the Node block's row→id order (read back via each
// row's own inline Label bytes on its dedicated NODE stream frame — see
// Buffer.BuildNodeStreamFrame; topology/nodes/*/meta.json have no data.label so each node's
// label falls back to its id) equals spec order — the directory-sorted node id order
// LoadTopology reads the topology in (readDirNames + sort.Strings; lexicographic, so "10"
// sorts before "2").
//
// This is the end-to-end proof that row order is a deterministic PROJECTION OF THE
// DIAGRAM (CLAUDE.md/MODEL.md), not a discovery log built by racing node goroutines to
// their first geometry emit. See headless_stream_helpers_test.go for the spawn/cleanup
// pattern this reuses; NEVER run the sim in the foreground
// (memory/feedback_no_foreground_sim_runs.md).
//
// NOTE: this used to also assert row order was IDENTICAL across 5 separate spawns of the
// binary. That half was removed (memory/feedback_no_deferrals.md's sibling doctrine,
// docs/testing-shape.md): node row order comes from sort.Strings over a directory listing,
// which is deterministic by construction, so re-running the binary 4 more times to observe
// that a sort sorted the same way exercised Go's runtime, not this codebase. See
// docs/headless-test-latency.md.

package main

import (
	"os"
	"testing"

	B "github.com/dtauraso/wirefold/Buffer"
)

// nodeStreamRowID decodes ONE dedicated NODE-stream frame's own inline Label bytes (see
// Buffer.BuildNodeStreamFrame's header: [tick,labelLen,layoutLinkCount,chainBeadCount] =
// 4×u32, then the Node row, then this frame's own label bytes inline — no port section any
// more, docs/channels-not-ports.md).
func nodeStreamRowID(frame []byte) string {
	// The header width comes from Buffer, never a literal: this parsing duplicates
	// BuildNodeStreamFrame's layout, and a hardcoded copy silently reads the wrong
	// offset the moment a header field is added (it did — adding chainBeadCount made
	// these read the label from inside the Node row and assert on garbage bytes).
	const hdrSize = B.BufNodeStreamFrameHeaderSize
	labelLen := int(readU32(frame, 4))
	labelOff := hdrSize + B.BufNodeStride
	return string(frame[labelOff : labelOff+labelLen])
}

// TestHeadlessNodeRowOrderMatchesSpecOrder runs the real binary once against the real
// topology/ dir and asserts the node-row id order (each node row's own dedicated NODE-fd,
// keyed by fd position = row) equals directory/spec order.
func TestHeadlessNodeRowOrderMatchesSpecOrder(t *testing.T) {
	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	binPath := buildHeadlessBinary(t, repoRoot, "wirefold-headless-row-order-test")

	want := wantNodeRowOrder(t, repoRoot)

	ds := spawnDedicatedAllStreams(t, binPath, repoRoot)
	nodeFrames := readLastFrames(t, ds.nodeReads, "node")
	got := make([]string, len(nodeFrames))
	for row, frame := range nodeFrames {
		got[row] = nodeStreamRowID(frame)
	}

	if len(got) != len(want) {
		t.Fatalf("node row count %d != topology/nodes dir count %d", len(got), len(want))
	}
	for row := range want {
		if got[row] != want[row] {
			t.Fatalf("row %d: got id %q, want spec-order id %q (full: got=%v want=%v)", row, got[row], want[row], got, want)
		}
	}
}
