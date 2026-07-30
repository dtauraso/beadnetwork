// Buffer/frame_tags.go — SYNTHETIC ext-host-side tags for the dedicated per-owner stream
// frames (memory/feedback_no_single_writer_bridge.md, Buffer/stream_fds.go), plus the
// per-stream frame ENVELOPE constants (header sizes, the node-stream layout-link row
// width). This file is the SINGLE Go-side source for all of it; the mirrored TS file,
// tools/topology-vscode/src/schema/frame-tags.ts, is GENERATED from it by
// tools/gen-node-defs (see tools/gen-node-defs/frame_tags.go) — every const below that
// carries a `//frametag:ts=NAME` marker comment is emitted verbatim as `export const NAME`
// in that generated file, so this file's doc comments ARE the TS JSDoc.
//
// This is deliberately NOT part of the generated column-layout schema
// (Buffer/layout.go / buffer_layout_gen.go / buffer-layout.ts): it is not a column
// inside a block, it is envelope-level plumbing (frame header sizes and synthetic
// relay tags), so it has its own small generator pipeline instead of the block/column
// one — mirroring the existing input_codec.go / input-layout.ts split (binary wire
// ENVELOPE constants live separately from the column-layout schema).
//
// Every dedicated stream fd (view/edge/node/interior) carries frames as
// [len:u32-LE][payload] with NO tag byte on the wire — the fd POSITION already
// identifies which stream/row it is (see Buffer/stream_fds.go). The four BufBlockTag*
// constants below exist ONLY so the ext host can relay a decoded frame to the webview
// under one uniform "buffer-snapshot" message shape (tag + optional row), letting the
// render tree route by cell without a second message shape. They are NEVER written as a
// wire tag byte by Go.
//
//   - BufBlockTagView: a decoded VIEW-stream frame (camera + overlay + scene-sphere,
//     built by BuildViewStreamFrame). Singleton — no row.
//   - BufBlockTagEdgeStream: a decoded per-edge stream frame (BuildEdgeStreamFrame),
//     plus a `row` (that edge's stable seed-order row).
//   - BufBlockTagNodeStream: a decoded per-node NODE stream frame
//     (BuildNodeStreamFrame — geometry+ports+label), plus a `row` (that node's stable
//     seed-order row).
//   - BufBlockTagInteriorStream: a decoded per-node INTERIOR stream frame
//     (BuildInteriorStreamFrame — that node's own interior beads), plus a `row` (same
//     node-row numbering as BufBlockTagNodeStream, a SEPARATE goroutine's fd).
package Buffer

// BufBlockTagView is the SYNTHETIC ext-host-side tag for a decoded VIEW-stream frame
// (camera+overlay+scene), relayed to the webview under the same message shape as the other
// stream tags below. NEVER a wire tag byte on the dedicated view fd itself — that fd's
// frames carry no tag byte at all (see this file's header comment). Its payload layout
// (dedicated-fd wire bytes, no tag): BufViewFrameHeaderSize (4) bytes [tick:u32], then a
// Camera row, an Overlay row, a Scene row.
//
//frametag:ts=BUF_BLOCK_TAG_VIEW
const BufBlockTagView byte = 4

// BufViewFrameHeaderSize is the byte width of the VIEW stream's own frame header on its
// dedicated fd: [tick:u32]. Hand-authored here (envelope-level) rather than generated from
// the column-layout schema, mirroring BufHeaderSize's split from that schema.
//
//frametag:ts=BUF_VIEW_FRAME_HEADER_SIZE
const BufViewFrameHeaderSize = 4

// BufBlockTagEdgeStream is the SYNTHETIC ext-host-side tag for a decoded per-edge
// dedicated-stream frame (one edgeMover writes ITS OWN combined edge+bead frame to its own
// fd — see Buffer/stream_fds.go's StreamKindEdge). NEVER a wire tag byte on the dedicated
// per-edge fd itself (the fd POSITION already identifies which edge). Relayed to the
// webview under the same "buffer-snapshot" shape as BufBlockTagView, plus a `row` field
// (there are many edge streams, not a singleton).
//
//frametag:ts=BUF_BLOCK_TAG_EDGE_STREAM
const BufBlockTagEdgeStream byte = 5

// BufEdgeStreamFrameHeaderSize is the byte width of the leading header on one edge's
// combined per-fd frame (Buffer.BuildEdgeStreamFrame), before the Edge row: [tick:u32].
// The rest of that frame's byte layout: one BufEdgeStride row (SX..EZ/Selected,
// EdgeLabelOff=0/Len) + that edge's own label bytes (labelLen, from the row).
//
//frametag:ts=BUF_EDGE_STREAM_FRAME_HEADER_SIZE
const BufEdgeStreamFrameHeaderSize = 4

// BufBlockTagNodeStream is the SYNTHETIC ext-host-side tag for a decoded per-node
// dedicated-stream frame (one nodeMover writes ITS OWN node geometry + ports + label to
// its own fd — see Buffer/stream_fds.go's StreamKindNode / Buffer/node_stream_frame.go's
// BuildNodeStreamFrame). NEVER a wire tag byte on the dedicated per-node fd itself (the fd
// POSITION already identifies which node). Relayed under the same "buffer-snapshot" shape
// as BufBlockTagEdgeStream, plus a `row` field (one per node row).
//
//frametag:ts=BUF_BLOCK_TAG_NODE_STREAM
const BufBlockTagNodeStream byte = 6

// BufBlockTagInteriorStream is the SYNTHETIC ext-host-side tag for a decoded per-node
// INTERIOR stream frame (that node's OWN Update goroutine writes its interior beads to its
// own fd — see Buffer/stream_fds.go's StreamKindInterior / Buffer/node_stream_frame.go's
// BuildInteriorStreamFrame). NEVER a wire tag byte on the dedicated fd itself. Relayed
// under the same "buffer-snapshot" shape as BufBlockTagNodeStream, plus a `row` field (one
// per node row, same numbering as BufBlockTagNodeStream, a SEPARATE goroutine's fd).
//
//frametag:ts=BUF_BLOCK_TAG_INTERIOR_STREAM
const BufBlockTagInteriorStream byte = 7

// BufNodeStreamFrameHeaderSize is the byte width of the leading header on one node's
// combined per-fd frame (Buffer.BuildNodeStreamFrame), before the Node row:
// [tick:u32][labelLen:u32][layoutLinkCount:u32][chainBeadCount:u32]. No port section any
// more (docs/channels-not-ports.md — a port carries no geometry, so there is no portCount/
// portNameBytesCount to size).
// The rest of that frame's layout: one BufNodeStride row (LabelOff=0 into this frame's own
// label bytes) + labelLen label bytes + layoutLinkCount × BufNodeStreamLayoutLinkStride
// layout-link rows (this node's OWN outbound layout-links — see buffer-decode.ts's
// DecodedNodeStreamFrame doc comment) + chainBeadCount chain-bead rows.
//
//frametag:ts=BUF_NODE_STREAM_FRAME_HEADER_SIZE
const BufNodeStreamFrameHeaderSize = 16

// BufNodeStreamLayoutLinkStride is the byte width of ONE layout-link (cascade-link
// overlay) row within a node stream frame: [DstNodeRow:i32]. Narrower than the combined
// LayoutLink block's BufLayoutLinkStride (8 bytes, SrcNodeRow+DstNodeRow) because on a
// per-node stream the source IS this node — its own row is implicit (the fd position /
// the aggregator's row index), so only the dst endpoint travels. No edge-row column: the
// cascade-link overlay draws between the two nodes' CENTERS, never along a bead edge.
//
//frametag:ts=NODE_STREAM_LAYOUT_LINK_STRIDE
const BufNodeStreamLayoutLinkStride = 4

// NOTE: there is deliberately NO BufNodeStreamChainBeadStride here. A chain-bead row on a
// node stream is byte-identical to a ChainBead BLOCK row, so BufChainBeadStride (generated
// from Buffer/layout.go, and CHAIN_BEAD_STRIDE on the TS side) is the single source for both.
// A second copy existed briefly and immediately went stale when the Lit column was added:
// the packer allocated 12 bytes per bead and wrote 13, panicking on a slice bound. The
// LayoutLink stride above is a genuinely DIFFERENT width from its block (no SrcNodeRow on a
// per-node stream), which is why that one has to exist and this one must not.

// BufInteriorStreamFrameHeaderSize is the byte width of the leading header on one node's
// INTERIOR per-fd frame (Buffer.BuildInteriorStreamFrame), before the interior rows:
// [tick:u32]. Followed by a FIXED BufInteriorSlotsPerNode × BufInteriorStride bytes (no
// count — the decoder derives the length from the fixed per-node slot count, same as the
// combined Interior block).
//
//frametag:ts=BUF_INTERIOR_STREAM_FRAME_HEADER_SIZE
const BufInteriorStreamFrameHeaderSize = 4
