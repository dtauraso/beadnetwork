// Buffer/node_stream_frame.go — the per-node dedicated-stream frame packers (see
// Buffer/stream_fds.go's StreamKindNode/StreamKindInterior doc comments and
// memory/feedback_no_single_writer_bridge.md). Two frames, one per emitting goroutine:
//
//   - BuildNodeStreamFrame is written by ONE node's nodeMover goroutine — its own Node row
//     (center/radius/ring-normals/selection-UI columns) + its own ports + its own inline
//     label bytes. No sub-tag byte (the fd position already identifies which node).
//   - BuildInteriorStreamFrame is written by ONE node's OWN Update goroutine (the second
//     emitting goroutine per node) — its own fixed 4-slot interior-bead grid.
//
// Both take plain parallel slices/scalars rather than any Buffer-side struct, mirroring
// BuildEdgeStreamFrame's shape: the emitting side lives in nodes/Wiring, which must stay
// Buffer-independent (see PortRowResolver/EdgeRowResolver's existing interface-injection
// pattern), so the injected build funcs it holds can only close over plain Go values.
package Buffer

import (
	"encoding/binary"
	"fmt"
)

// BufNodeStreamLayoutLinkStride now lives in Buffer/frame_tags.go (envelope constants,
// generated into frame-tags.ts alongside the other frame envelope sizes).

// BuildNodeStreamFrame packs one node's combined per-fd frame payload (no outer tag byte
// — the fd position already identifies which node this is):
//
//	[tick:u32]
//	[portCount:u32]
//	[labelLen:u32]
//	[portNameBytesCount:u32]
//	[layoutLinkCount:u32]
//	[chainBeadCount:u32]
//	Node       BufNodeStride bytes (SAME SetNodeRow column writer buildNodeFrame uses;
//	           LabelOff=0 into this frame's own label bytes, NodeRow-local — nodeRow is
//	           carried separately below for the Port rows' NodeRow column)
//	Label      labelLen bytes (this node's own label bytes — inline, not a shared section)
//	Port       portCount × BufPortStride bytes (SAME SetPortRow column writer buildNodeFrame
//	           uses; every row's NodeRow = nodeRow, PortNameOff/Len into this frame's own
//	           port-name bytes)
//	PortName   portNameBytesCount bytes (this node's own ports' name bytes, concatenated in
//	           the same order as the Port rows above)
//	LayoutLink layoutLinkCount × BufNodeStreamLayoutLinkStride bytes — the cascade-link
//	           pairs for which THIS node is the lexicographically-smaller endpoint (see
//	           nodes/Wiring/node_mover.go's cascadeEdges doc comment): each row is
//	           [DstNodeRow:i32], dstNodeRows a single parallel slice. The overlay draws
//	           between the two nodes' CENTERS (Node block), never a bead edge — no
//	           edge-row travels here.
//	ChainBead  chainBeadCount × BufNodeStreamChainBeadStride bytes — this node's OWN
//	           placeholder chain beads: NODE-LOCAL offsets + the Lit animation flag
//	           ([OX,OY,OZ] f32 + [Lit] u8), concatenated
//	           across all of this node's outgoing edges in that order. The chain is the
//	           VISUAL of a traversal, never a picture of the node-to-node channels
//	           (docs/beads-are-the-edge.md); nothing here identifies a channel or a message.
func BuildNodeStreamFrame(
	tick uint32, nodeRow int32,
	cx, cy, cz, radius, sphereR float32,
	vrx, vry, vrz, frx, fry, frz float32,
	selected, kindID, hovered, latchedSel, gotDragMsg uint8,
	dragDeltaA, dragDeltaB, dragDeltaC, dragRequantCount int32,
	gotForwardMsg uint8,
	forwardDeltaA, forwardDeltaB, forwardDeltaC, forwardFromRow int32,
	cascadeRelay uint8,
	label string,
	portNames []string,
	portDX, portDY, portDZ, portPX, portPY, portPZ []float32,
	portIsInput, portHovered []uint8,
	dstNodeRows []int32,
	chainBeadOX, chainBeadOY, chainBeadOZ []float32,
	chainBeadLit []uint8,
	events []StreamEvent,
) []byte {
	labelBytes := []byte(label)
	portCount := len(portNames)
	// INVARIANT: the port slices are PARALLEL — one entry per port, same order. The doc
	// comment says so and the loop below indexes all nine at i without checking, so a
	// short slice is either an opaque "index out of range" naming no port and no node, or
	// (for a LONG slice) silently dropped columns nobody notices. Named here instead.
	for _, s := range []struct {
		name string
		n    int
	}{
		{"portDX", len(portDX)}, {"portDY", len(portDY)}, {"portDZ", len(portDZ)},
		{"portPX", len(portPX)}, {"portPY", len(portPY)}, {"portPZ", len(portPZ)},
		{"portIsInput", len(portIsInput)}, {"portHovered", len(portHovered)},
	} {
		if s.n != portCount {
			panic(fmt.Sprintf(
				"BuildNodeStreamFrame: node row %d has %d port names but %s has %d entries — the port slices are parallel, one entry per port",
				nodeRow, portCount, s.name, s.n))
		}
	}
	portNameBytes := make([]byte, 0, portCount*8)
	portNameOffs := make([]uint32, portCount)
	portNameLens := make([]uint32, portCount)
	for i, n := range portNames {
		portNameOffs[i] = uint32(len(portNameBytes))
		nb := []byte(n)
		portNameLens[i] = uint32(len(nb))
		portNameBytes = append(portNameBytes, nb...)
	}
	layoutLinkCount := len(dstNodeRows)
	chainBeadCount := len(chainBeadOX)
	// INVARIANT: the three chain-bead slices are PARALLEL, one entry per bead, same order —
	// same reasoning as the port slices above (a short slice is an opaque index panic naming
	// no node; a long one is silently dropped).
	for _, s := range []struct {
		name string
		n    int
	}{{"chainBeadOY", len(chainBeadOY)}, {"chainBeadOZ", len(chainBeadOZ)}, {"chainBeadLit", len(chainBeadLit)}} {
		if s.n != chainBeadCount {
			panic(fmt.Sprintf(
				"BuildNodeStreamFrame: node row %d has %d chain-bead OX entries but %s has %d — the chain-bead slices are parallel, one entry per bead",
				nodeRow, chainBeadCount, s.name, s.n))
		}
	}

	size := BufNodeStreamFrameHeaderSize + BufNodeStride + len(labelBytes) + portCount*BufPortStride + len(portNameBytes) +
		layoutLinkCount*BufNodeStreamLayoutLinkStride + chainBeadCount*BufChainBeadStride
	buf := make([]byte, size)
	off := 0
	binary.LittleEndian.PutUint32(buf[off:], tick)
	off += 4
	binary.LittleEndian.PutUint32(buf[off:], uint32(portCount))
	off += 4
	binary.LittleEndian.PutUint32(buf[off:], uint32(len(labelBytes)))
	off += 4
	binary.LittleEndian.PutUint32(buf[off:], uint32(len(portNameBytes)))
	off += 4
	binary.LittleEndian.PutUint32(buf[off:], uint32(layoutLinkCount))
	off += 4
	binary.LittleEndian.PutUint32(buf[off:], uint32(chainBeadCount))
	off += 4

	SetNodeRow(buf[off:off+BufNodeStride], 0, cx, cy, cz, radius, sphereR, vrx, vry, vrz, frx, fry, frz,
		selected, kindID, 0, uint32(len(labelBytes)), hovered, latchedSel, gotDragMsg,
		dragDeltaA, dragDeltaB, dragDeltaC, dragRequantCount,
		gotForwardMsg, forwardDeltaA, forwardDeltaB, forwardDeltaC, forwardFromRow, cascadeRelay)
	off += BufNodeStride

	copy(buf[off:off+len(labelBytes)], labelBytes)
	off += len(labelBytes)

	portBuf := buf[off : off+portCount*BufPortStride]
	for i := range portNames {
		SetPortRow(portBuf, i, nodeRow, portDX[i], portDY[i], portDZ[i], portPX[i], portPY[i], portPZ[i],
			portIsInput[i], portHovered[i], portNameOffs[i], portNameLens[i])
	}
	off += portCount * BufPortStride

	copy(buf[off:off+len(portNameBytes)], portNameBytes)
	off += len(portNameBytes)

	for i := 0; i < layoutLinkCount; i++ {
		rowOff := off + i*BufNodeStreamLayoutLinkStride
		binary.LittleEndian.PutUint32(buf[rowOff:], uint32(dstNodeRows[i]))
	}
	off += layoutLinkCount * BufNodeStreamLayoutLinkStride

	for i := 0; i < chainBeadCount; i++ {
		rowOff := off + i*BufChainBeadStride
		SetChainBeadRow(buf[rowOff:rowOff+BufChainBeadStride], 0, chainBeadOX[i], chainBeadOY[i], chainBeadOZ[i], chainBeadLit[i])
	}
	off += chainBeadCount * BufChainBeadStride

	// INVARIANT: the walk that WRITES the frame ends exactly where the `size` formula that
	// ALLOCATED it says it should. This is the runtime half of buffer-layout parity —
	// check-buffer-layout-parity.sh compares the two GENERATED files' fingerprints and
	// check-generated.sh catches a stale regen, but neither reads this function, so adding
	// or reordering a SECTION here without updating `size` (or the reverse) is caught by
	// nothing. The failure mode is a frame with trailing zero bytes or a truncated tail,
	// which the TS decoder reads as real columns — a wrong scene that still renders.
	// (Column-level drift inside a row is a different question, covered by the generated
	// setters; this pins the section walk.)
	if off != size {
		panic(fmt.Sprintf(
			"BuildNodeStreamFrame: packed %d bytes for node row %d but allocated %d — the section walk and the size formula disagree; a section was added, reordered, or resized in one of the two and not the other",
			off, nodeRow, size))
	}

	return append(buf, BuildEventsSection(events)...)
}

// BuildInteriorStreamFrame packs one node's fixed-slot interior-bead frame payload (no
// outer tag byte): [tick:u32] followed by len(present) Interior rows (SAME SetInteriorRow
// column writer buildNodeFrame uses) — no count, the decoder derives the length from the
// fixed per-node slot count (BufInteriorSlotsPerNode), same as the combined Interior
// block. present/value/ox/oy/oz are parallel slices, same length, same slot order.
func BuildInteriorStreamFrame(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []StreamEvent) []byte {
	n := len(present)
	// INVARIANT: present/value/ox/oy/oz are parallel, same length, same slot order — the
	// doc comment above says exactly that, and the loop below indexes all five at i. Same
	// reasoning as BuildNodeStreamFrame's port slices: unchecked, a short slice is an
	// opaque index panic and a long one is silently truncated.
	for _, s := range []struct {
		name string
		n    int
	}{{"value", len(value)}, {"ox", len(ox)}, {"oy", len(oy)}, {"oz", len(oz)}} {
		if s.n != n {
			panic(fmt.Sprintf(
				"BuildInteriorStreamFrame: %d present slots but %s has %d entries — the interior slot slices are parallel, one entry per slot",
				n, s.name, s.n))
		}
	}
	buf := make([]byte, BufInteriorStreamFrameHeaderSize+n*BufInteriorStride)
	binary.LittleEndian.PutUint32(buf[0:], tick)
	interiorBuf := buf[4:]
	for i := 0; i < n; i++ {
		SetInteriorRow(interiorBuf, i, present[i], value[i], ox[i], oy[i], oz[i])
	}
	return append(buf, BuildEventsSection(events)...)
}
