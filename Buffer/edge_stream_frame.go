// Buffer/edge_stream_frame.go — the per-edge dedicated-stream frame packer (see
// Buffer/stream_fds.go's StreamKindEdge doc comment and
// memory/feedback_no_single_writer_bridge.md). This is the combined frame ONE edgeMover
// goroutine writes to ITS OWN fd every cycle it changes (geometry recompute OR a bead
// step) — the edge's own Edge-block row PLUS the beads currently in flight on that
// edge's wire, with no sub-tag byte (the fd position already identifies which edge).
//
// Wire bytes (no outer tag, unlike fd3's [len][tag][payload] — a dedicated fd's OWN
// position already identifies the stream, so there is nothing left to discriminate):
//
//	[tick:u32]
//	Edge     BufEdgeStride bytes (SX..EZ/Selected + EdgeLabelOff=0/Len,
//	         written via the SAME SetEdgeRow column writer buildEdgeFrame uses)
//	EdgeLabel labelLen bytes (this edge's own label bytes — inline, not a shared section:
//	         each edge's own stream carries its own label bytes)
//
// Injected into nodes/Wiring's MoveDispatch.SetEdgeStreams as a plain func (not a Buffer
// import in the Wiring package — mirrors PortRowResolver/EdgeRowResolver's existing
// interface-injection pattern, keeping Wiring Buffer-independent).
package Buffer

import (
	"encoding/binary"
	"fmt"
)

// BuildEdgeStreamFrame packs one edge's combined per-fd frame payload (see this file's
// header comment for the byte layout).
//
// There is NO bead section any more. The transit bead is not drawn: the animation is the LIT
// bead on the SOURCE NODE's own placeholder chain (docs/beads-are-the-edge.md), which that
// node computes and streams on its own node frame. Removing it also removed a real race —
// the bead rows were read via PacedWire.LiveBeadRows from the EDGE goroutine, while the wire
// is now stepped by its source node's goroutine, so that read no longer satisfied the
// single-goroutine ownership pw.inflight requires.
func BuildEdgeStreamFrame(tick uint32, sx, sy, sz, ex, ey, ez float32, selected uint8, label string, events []StreamEvent) []byte {
	labelBytes := []byte(label)
	size := BufEdgeStreamFrameHeaderSize + BufEdgeStride + len(labelBytes)
	buf := make([]byte, size)
	off := 0
	binary.LittleEndian.PutUint32(buf[off:], tick)
	off += 4
	// edgeLabelOff=0: this frame's own label bytes immediately follow the Edge row —
	// there is no shared EdgeLabel section on a dedicated per-edge stream.
	SetEdgeRow(buf[off:off+BufEdgeStride], 0, sx, sy, sz, ex, ey, ez, selected, 0, uint32(len(labelBytes)))
	off += BufEdgeStride
	copy(buf[off:off+len(labelBytes)], labelBytes)
	off += len(labelBytes)
	// INVARIANT: walk ends where the size formula says — the runtime half of buffer-layout
	// parity, same as BuildNodeStreamFrame's (see the full reasoning there).
	if off != size {
		panic(fmt.Sprintf(
			"BuildEdgeStreamFrame: packed %d bytes for edge %q but allocated %d — the section walk and the size formula disagree",
			off, label, size))
	}
	return append(buf, BuildEventsSection(events)...)
}
