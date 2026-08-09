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
// The emitting side lives in nodes/Wiring, which must stay Buffer-independent (see
// PortRowResolver/EdgeRowResolver's existing interface-injection pattern), so the injected
// build funcs it holds can only close over values of ITS OWN types — that is why
// Wiring.NodeFrameInput and Buffer.NodeStreamFrame are two structs with the same fields,
// mapped field-by-name in main.go's injected closure (exactly as toStreamEvents already
// maps Wiring.RowEvent → Buffer.StreamEvent for the same reason). BuildInteriorStreamFrame
// and BuildEdgeStreamFrame keep plain parallel slices/scalars: their parameter lists are
// short and not a long run of same-typed values.
package Buffer

import (
	"encoding/binary"
	"fmt"
)

// NodeStreamFrame is BuildNodeStreamFrame's single named-field input — one node's own
// frame, assigned BY NAME at every call site.
//
// WHY A STRUCT: this used to be a ~35-parameter positional signature whose middle was a run
// of 20 consecutive float32s, then 5 consecutive uint8s, then 2 int32s. Transposing any two
// adjacent same-typed arguments (say TopTiltVectorLen/TopTiltVectorTheta, or
// BottomTiltVectorTheta/CoplanarNormalTheta) COMPILED CLEANLY and streamed a silently wrong
// scene — no compiler check, no test that obviously catches it. Named fields make that bug
// class unrepresentable (memory/feedback_make_bug_class_unrepresentable.md): the same
// mistake is now a wrong field NAME, which is either a compile error or plainly readable.
//
// The field order below MIRRORS the packed column order so a reader can diff it against
// Buffer/layout.go — but the WIRE FORMAT is defined by BuildNodeStreamFrame's packing walk
// and the generated SetNodeRow setter, NOT by this struct. Reordering fields here changes
// nothing on the wire, and must not be mistaken for a format change.
type NodeStreamFrame struct {
	// Tick is the emitting goroutine's own clock tick, stamped on this frame.
	Tick uint32
	// NodeRow is this node's stable buffer row — ROW ID = NODE ID - 1
	// (.claude/rules/persistence-ownership.md). Used here for the panic messages only;
	// the fd position is what identifies the node on the wire.
	NodeRow int32
	// NodeID is the node's own numeric identity (NodeRow+1 by construction), streamed so a
	// misrouted frame can be caught against its arrival row rather than trusted.
	NodeID int32
	// CX/CY/CZ are this node's own world-space centre.
	CX, CY, CZ float32
	// Radius is the node's own drawn radius; SphereR is the scene sphere's radius.
	Radius, SphereR float32
	// VRX/VRY/VRZ and FRX/FRY/FRZ are the vertical- and flat-ring normals.
	VRX, VRY, VRZ float32
	FRX, FRY, FRZ float32
	// PoleTheta/PolePhi are the node's own local-frame pole — its scene-polar direction
	// reversed, so the frame points back at the scene centre.
	PoleTheta, PolePhi float32
	// RingAxisTheta/RingAxisPhi are the DRAWN ring's axis, separate from the navigation
	// pole above (Buffer/layout.go).
	RingAxisTheta, RingAxisPhi float32
	// TopTiltVectorLen encodes WHETHER-AND-HOW-FAR: 0 means this node draws no tilt
	// vector at all; non-zero is the node's own radius (the vector runs centre → top).
	TopTiltVectorLen float32
	// TopTiltVectorTheta is the vector's OWN direction, separate from the ring axis, so a
	// scene can aim a node's vector somewhere other than its ring. θ-only, no φ:
	// index × the node's own lattice step (task/drop-tilt-vector-phi).
	TopTiltVectorTheta float32
	// BottomTiltVectorTheta is a half turn in θ from the top; it shares the top's length
	// column, so it is drawn exactly when TopTiltVectorLen is non-zero.
	BottomTiltVectorTheta float32
	// CoplanarNormalTheta is a fixed quarter turn in θ from the top tilt index — it turns
	// WITH the tilt rather than staying aimed at a partner.
	CoplanarNormalTheta float32
	// ReceivedVectorLen follows the same whether-and-how-far convention as
	// TopTiltVectorLen: ZERO when nothing has been received on this node's tilt-vector
	// channel (or a reset cleared it), the node's own radius otherwise — so "nothing
	// received" stays distinguishable from "received a direction that happens to be 0".
	ReceivedVectorLen float32
	// ReceivedVectorTheta is that received direction (meaningless when the length is 0).
	ReceivedVectorTheta float32
	// Selected/Hovered/LatchedSel are this node's OWN selection-UI bits (never a shared
	// selection map); KindID is its NODE_DEFS index, static after load.
	Selected, KindID, Hovered, LatchedSel uint8
	// LatticePoints is this node's own pair-lattice point count — the N the θ columns
	// above were converted against.
	LatticePoints uint8
	// RoundsToParallel is this node's own rounds-to-rest count (vector-exchange rounds
	// between START and its rule settling); MsgsToParallel is the same span in messages.
	RoundsToParallel, MsgsToParallel int32
	// Label is this node's own label bytes, packed INLINE in this frame (not a shared
	// section).
	Label string
	// ChainBeadOX/OY/OZ/Lit/LitValue are PARALLEL slices, one entry per placeholder chain
	// bead, NODE-LOCAL offsets, concatenated across this node's outgoing edges in order.
	// A length mismatch panics below.
	ChainBeadOX, ChainBeadOY, ChainBeadOZ []float32
	ChainBeadLit                          []uint8
	ChainBeadLitValue                     []int32
	// Events is whatever this node's own caller wants riding this frame's trailing EVENTS
	// section (nil from a plain tick-driven write).
	Events []StreamEvent
}

// BuildNodeStreamFrame packs one node's combined per-fd frame payload (no outer tag byte
// — the fd position already identifies which node this is):
//
//	[tick:u32]
//	[labelLen:u32]
//	[chainBeadCount:u32]
//	Node       BufNodeStride bytes (SAME SetNodeRow column writer buildNodeFrame uses;
//	           LabelOff=0 into this frame's own label bytes, NodeRow-local)
//	Label      labelLen bytes (this node's own label bytes — inline, not a shared section)
//	ChainBead  chainBeadCount × BufNodeStreamChainBeadStride bytes — this node's OWN
//	           placeholder chain beads: NODE-LOCAL offsets + the Lit animation flag
//	           ([OX,OY,OZ] f32 + [Lit] u8), concatenated
//	           across all of this node's outgoing edges in that order. The chain is the
//	           VISUAL of a traversal, never a picture of the node-to-node channels
//	           (docs/beads-are-the-edge.md); nothing here identifies a channel or a message.
//
// The Port block/section is GONE (docs/channels-not-ports.md): a port is a load-time
// channel-binding ROLE, never a place, so it has no row here any more. An edge's own
// endpoints ride the Edge block's SX..EZ instead (Buffer/edge_stream_frame.go).
func BuildNodeStreamFrame(f NodeStreamFrame) []byte {
	labelBytes := []byte(f.Label)
	chainBeadCount := len(f.ChainBeadOX)
	// INVARIANT: the chain-bead slices are PARALLEL, one entry per bead, same order —
	// a short slice is an opaque index panic naming no node; a long one is silently dropped.
	for _, s := range []struct {
		name string
		n    int
	}{{"ChainBeadOY", len(f.ChainBeadOY)}, {"ChainBeadOZ", len(f.ChainBeadOZ)}, {"ChainBeadLit", len(f.ChainBeadLit)}, {"ChainBeadLitValue", len(f.ChainBeadLitValue)}} {
		if s.n != chainBeadCount {
			panic(fmt.Sprintf(
				"BuildNodeStreamFrame: node row %d has %d chain-bead OX entries but %s has %d — the chain-bead slices are parallel, one entry per bead",
				f.NodeRow, chainBeadCount, s.name, s.n))
		}
	}

	size := BufNodeStreamFrameHeaderSize + BufNodeStride + len(labelBytes) +
		chainBeadCount*BufChainBeadStride
	buf := make([]byte, size)
	off := 0
	binary.LittleEndian.PutUint32(buf[off:], f.Tick)
	off += 4
	binary.LittleEndian.PutUint32(buf[off:], uint32(len(labelBytes)))
	off += 4
	binary.LittleEndian.PutUint32(buf[off:], uint32(chainBeadCount))
	off += 4

	SetNodeRow(buf[off:off+BufNodeStride], 0, f.NodeID, f.CX, f.CY, f.CZ, f.Radius, f.SphereR, f.VRX, f.VRY, f.VRZ, f.FRX, f.FRY, f.FRZ,
		f.PoleTheta, f.PolePhi, f.RingAxisTheta, f.RingAxisPhi, f.TopTiltVectorLen, f.TopTiltVectorTheta, f.BottomTiltVectorTheta, f.CoplanarNormalTheta, f.ReceivedVectorLen, f.ReceivedVectorTheta, f.Selected, f.KindID, 0, uint32(len(labelBytes)), f.Hovered, f.LatchedSel, f.LatticePoints, f.RoundsToParallel, f.MsgsToParallel)
	off += BufNodeStride

	copy(buf[off:off+len(labelBytes)], labelBytes)
	off += len(labelBytes)

	for i := 0; i < chainBeadCount; i++ {
		rowOff := off + i*BufChainBeadStride
		SetChainBeadRow(buf[rowOff:rowOff+BufChainBeadStride], 0, f.ChainBeadOX[i], f.ChainBeadOY[i], f.ChainBeadOZ[i], f.ChainBeadLit[i], f.ChainBeadLitValue[i])
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
			off, f.NodeRow, size))
	}

	return append(buf, BuildEventsSection(f.Events)...)
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
