// headless_node_fd_test.go — drives the REAL compiled binary headlessly and proves the
// per-node dedicated-stream migration end-to-end (memory/feedback_no_single_writer_bridge.md,
// Buffer/stream_fds.go's StreamKindNode/StreamKindInterior): with every per-owner fd wired
// (mandatory — no fallback path left, memory/feedback_no_single_writer_bridge.md's final step), every node's
// own combined frame (Node fields + ports + label) arrives on its OWN "node" fd, and every
// node's own interior-bead frame arrives on its OWN "interior" fd.
//
// See headless_stream_helpers_test.go for the spawn/cleanup pattern this reuses; NEVER run
// the sim in the foreground (memory/feedback_no_foreground_sim_runs.md).
package main

import (
	"os"
	"sort"
	"testing"

	B "github.com/dtauraso/wirefold/Buffer"
)

// TestHeadlessNodeFdDedicatedStream proves each node's combined geometry+ports+label frame
// arrives on its OWN "node" fd with resolvable content, and each node's interior frame
// arrives on its OWN "interior" fd.
func TestHeadlessNodeFdDedicatedStream(t *testing.T) {
	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	binPath := buildHeadlessBinary(t, repoRoot, "wirefold-headless-node-fd-test")
	ds := spawnDedicatedAllStreams(t, binPath, repoRoot)

	nodeFrames := readLastFrames(t, ds.nodeReads, "node")
	interiorFrames := readLastFrames(t, ds.interiorReads, "interior")

	wantLabels := wantNodeRowOrder(t, repoRoot)
	sort.Strings(wantLabels) // wantNodeRowOrder already returns sorted spec order

	totalLayoutLinks := 0
	for row, frame := range nodeFrames {
		// Combined frame layout (Buffer.BuildNodeStreamFrame): [tick:u32][portCount:u32]
		// [labelLen:u32][portNameBytesCount:u32][layoutLinkCount:u32][chainBeadCount:u32] +
		// Node row + label bytes + Port rows + port-name bytes + LayoutLink rows
		// ([DstNodeRow:i32] each, BufNodeStreamLayoutLinkStride bytes — no edge-row column;
		// the cascade-link overlay draws node-center to node-center, never along a bead edge)
		// + ChainBead rows ([OX,OY,OZ] f32 node-local offsets, this node's own placeholder
		// chain — docs/beads-are-the-edge.md).
		// The header width comes from Buffer, never a literal: this parsing duplicates
		// BuildNodeStreamFrame's layout, and a hardcoded copy silently reads the wrong
		// offset the moment a header field is added (it did — adding chainBeadCount made
		// these read the label from inside the Node row and assert on garbage bytes).
		const hdrSize = B.BufNodeStreamFrameHeaderSize
		if len(frame) < hdrSize+B.BufNodeStride {
			t.Fatalf("node row %d: frame too short (%d bytes) to hold header+Node row", row, len(frame))
		}
		portCount := readU32(frame, 4)
		labelLen := readU32(frame, 8)
		portNameBytesCount := readU32(frame, 12)
		layoutLinkCount := readU32(frame, 16)
		chainBeadCount := readU32(frame, 20)
		nodeOff := hdrSize
		labelOff := nodeOff + B.BufNodeStride
		if labelOff+int(labelLen) > len(frame) {
			t.Fatalf("node row %d: label overruns frame (labelLen=%d, frameLen=%d)", row, labelLen, len(frame))
		}
		label := string(frame[labelOff : labelOff+int(labelLen)])
		if label == "" {
			t.Fatalf("node row %d: empty inline label", row)
		}
		if row < len(wantLabels) && label != wantLabels[row] {
			t.Fatalf("node row %d: label = %q, want %q (falls back to node id)", row, label, wantLabels[row])
		}
		portsOff := labelOff + int(labelLen)
		portNamesOff := portsOff + int(portCount)*B.BufPortStride
		layoutLinksOff := portNamesOff + int(portNameBytesCount)
		chainBeadsOff := layoutLinksOff + int(layoutLinkCount)*B.BufNodeStreamLayoutLinkStride
		eventsOff := chainBeadsOff + int(chainBeadCount)*B.BufChainBeadStride
		if eventsOff+4 > len(frame) {
			t.Fatalf("node row %d: frame too short (%d bytes) to hold the trailing EVENTS section count", row, len(frame))
		}
		eventCount := readU32(frame, eventsOff)
		wantLen := eventsOff + 4 + int(eventCount)*B.BufEventStride
		if wantLen != len(frame) {
			t.Fatalf("node row %d: frame length %d does not match computed layout end %d",
				row, len(frame), wantLen)
		}
		// Every Port row's NodeRow column must equal this frame's own node row (the
		// tear-free cross-stream contract: EdgeTube resolves a port row to (nodeRow,
		// portIndex) and looks up coordinates in THIS node's own cell).
		for p := 0; p < int(portCount); p++ {
			portRowOff := portsOff + p*B.BufPortStride
			gotNodeRow := int32(readU32(frame, portRowOff))
			if int(gotNodeRow) != row {
				t.Fatalf("node row %d: port %d's NodeRow column = %d, want %d", row, p, gotNodeRow, row)
			}
		}
		// Every LayoutLink row's DstNodeRow must be a valid, DIFFERENT node row (never
		// this node's own row — a node is never its own layout-link partner). No EdgeRow
		// column to check: the cascade-link overlay draws between the two nodes' CENTERS
		// (Node block), never along a bead edge.
		for l := 0; l < int(layoutLinkCount); l++ {
			rowOff := layoutLinksOff + l*B.BufNodeStreamLayoutLinkStride
			dstNodeRow := int32(readU32(frame, rowOff))
			if dstNodeRow < 0 || int(dstNodeRow) >= len(nodeFrames) {
				t.Fatalf("node row %d: layout-link %d DstNodeRow=%d out of range [0,%d)", row, l, dstNodeRow, len(nodeFrames))
			}
			if int(dstNodeRow) == row {
				t.Fatalf("node row %d: layout-link %d DstNodeRow equals its own row (self-link)", row, l)
			}
		}
		totalLayoutLinks += int(layoutLinkCount)
	}
	// This topology's cascade-edges data (topology/nodes/*/cascade-edges.json) declares
	// real cascade-link pairs — the per-node streams must actually carry SOME
	// layout-links, not silently zero every row (a real regression: e.g. cascadeEdges
	// never wired, or every dst id failing to resolve a node row).
	if totalLayoutLinks == 0 {
		t.Fatalf("no node row streamed any layout-link — expected this topology's cascade-edges pairs to appear on their source node's own fd")
	}

	for row, frame := range interiorFrames {
		// Fixed-slot frame (Buffer.BuildInteriorStreamFrame): [tick:u32] + 4 Interior rows +
		// a trailing EVENTS section ([count:u32] + count NodeBead rows).
		fixedLen := 4 + 4*B.BufInteriorStride
		if len(frame) < fixedLen+4 {
			t.Fatalf("interior row %d: frame length %d, want >= %d (fixed 4-slot layout + EVENTS count)", row, len(frame), fixedLen+4)
		}
		eventCount := readU32(frame, fixedLen)
		want := fixedLen + 4 + int(eventCount)*B.BufEventStride
		if len(frame) != want {
			t.Fatalf("interior row %d: frame length %d, want %d (fixed 4-slot layout + %d events)", row, len(frame), want, eventCount)
		}
	}
}
