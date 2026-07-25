// frame-tags.ts — hand-authored frame ENVELOPE discriminators. NOT generated (it's the
// outer frame's tag byte, not a column-layout block). The SYNTHETIC per-owner-stream tags
// 4-7 (VIEW/EDGE_STREAM/NODE_STREAM/INTERIOR_STREAM) are the only tags that exist — each
// has a Go counterpart, Buffer/frame_tags.go's BufBlockTagView/…/BufBlockTagInteriorStream —
// kept in lockstep by hand (the same way input-layout.ts mirrors input_codec.go). Tags 0-3
// (the retired combined SCENE/BEAD/NODE/EDGE frame layouts) are gone: Go now streams only
// per-owner dedicated frames (one per goroutine, over its own inherited pipe —
// memory/feedback_no_single_writer_bridge.md), never a shared combined frame.
//
// Current frame format (see runCommand.ts's handleViewFd/handleEdgeFd/handleNodeFd/
// handleInteriorFd for the per-goroutine dedicated-stream decoders): each dedicated fd's
// wire bytes carry NO outer tag byte at all — the fd POSITION identifies which stream (and,
// for edge/node/interior fds, which row). The tags below are SYNTHETIC: the ext host
// attaches one when relaying a decoded frame to the webview under the shared
// "buffer-snapshot" message shape, so all four stream kinds can ride one message shape
// instead of four.
//
//   - BUF_BLOCK_TAG_VIEW: SYNTHETIC ext-host-side tag for a decoded VIEW-stream frame
//     (camera + overlay + scene-sphere), streamed over its own dedicated inherited pipe
//     (see runCommand.ts's stream-fd allocation and Buffer/stream_fds.go —
//     memory/feedback_no_single_writer_bridge.md). The wire bytes on that dedicated fd
//     carry NO tag byte (the fd POSITION identifies the stream); this tag exists only so
//     the ext host can relay a decoded view frame to the webview under the SAME
//     "buffer-snapshot" message shape as the other synthetic stream tags below,
//     extending the existing tag-routed-cell pattern to a fifth cell instead of adding a
//     second message shape. Its payload layout (dedicated-fd wire bytes, no tag):
//
//     BUF_VIEW_FRAME_HEADER_SIZE (4) bytes: [tick:u32]
//     Camera   CAMERA_STRIDE bytes
//     Overlay  OVERLAY_STRIDE bytes
//     Scene    SCENE_STRIDE bytes
//
// The tag values below (4-7) are the only ones that exist; each goroutine streams its own
// frame over its own dedicated inherited pipe (WIREFOLD_STREAM_FDS) rather than a single
// shared pipe, and the ext host tags the decoded result on relay to the webview.

/** SYNTHETIC ext-host-side tag for a decoded VIEW-stream frame (camera+overlay+scene),
 * relayed to the webview under the same message shape as the other stream tags. NEVER a wire
 * tag byte on the dedicated view fd itself — see this file's header comment. Mirrors
 * Buffer/frame_tags.go's BufBlockTagView. */
export const BUF_BLOCK_TAG_VIEW = 4;

/** Byte width of the VIEW stream's own frame header on its dedicated fd: [tick:u32].
 * Hand-authored (envelope-level), mirroring the other per-stream header-size constants'
 * split from the generated column-layout schema. */
export const BUF_VIEW_FRAME_HEADER_SIZE = 4;

/** SYNTHETIC ext-host-side tag for a decoded per-edge dedicated-stream frame (one edgeMover
 * writes ITS OWN combined edge+bead frame to its own fd — see runCommand.ts's edge-fd
 * range and Buffer/stream_fds.go's StreamKindEdge). NEVER a wire tag byte on the dedicated
 * per-edge fd itself (the fd POSITION already identifies which edge). Relayed to the
 * webview under the same "buffer-snapshot" shape as BUF_BLOCK_TAG_VIEW, plus a `row` field
 * (there are many edge streams, not a singleton). Mirrors Buffer/frame_tags.go's
 * BufBlockTagEdgeStream. */
export const BUF_BLOCK_TAG_EDGE_STREAM = 5;

/** Byte layout of one edge's combined per-fd frame (Buffer.BuildEdgeStreamFrame), no outer
 * tag: [tick:u32] + one EDGE_STRIDE row (SrcPortRow/DstPortRow/Selected, EdgeLabelOff=0/Len)
 * + that edge's own label bytes (labelLen, from the row) + [beadCount:u32] + beadCount ×
 * BEAD_STRIDE bead rows. Header before the Edge row is just the tick. */
export const BUF_EDGE_STREAM_FRAME_HEADER_SIZE = 4;

/** SYNTHETIC ext-host-side tag for a decoded per-node dedicated-stream frame (one
 * nodeMover writes ITS OWN node geometry + ports + label to its own fd — see
 * runCommand.ts's node-fd range and Buffer/stream_fds.go's StreamKindNode). NEVER a wire
 * tag byte on the dedicated per-node fd itself (the fd POSITION already identifies which
 * node). Relayed to the webview under the same "buffer-snapshot" shape as
 * BUF_BLOCK_TAG_EDGE_STREAM, plus a `row` field (one per node row). Mirrors
 * Buffer/frame_tags.go's BufBlockTagNodeStream (ext-host-only; not carried on any Go wire
 * byte — this numbering only needs to stay distinct from the other synthetic tags here). */
export const BUF_BLOCK_TAG_NODE_STREAM = 6;

/** SYNTHETIC ext-host-side tag for a decoded per-node dedicated INTERIOR-stream frame (that
 * node's OWN Update goroutine writes its interior beads to its own fd — see
 * runCommand.ts's interior-fd range and Buffer/stream_fds.go's StreamKindInterior). NEVER a
 * wire tag byte on the dedicated fd itself. Relayed under the same "buffer-snapshot" shape,
 * plus a `row` field (one per node row, same numbering as BUF_BLOCK_TAG_NODE_STREAM). */
export const BUF_BLOCK_TAG_INTERIOR_STREAM = 7;

/** Byte layout of one node's combined per-fd frame (Buffer.BuildNodeStreamFrame), no outer
 * tag: [tick:u32][portCount:u32][labelLen:u32][portNameBytesCount:u32][layoutLinkCount:u32]
 * + one NODE_STRIDE row (LabelOff=0 into this frame's own label bytes) + labelLen label
 * bytes + portCount × PORT_STRIDE port rows (each row's NodeRow column already the global
 * node row) + portNameBytesCount port-name bytes + layoutLinkCount ×
 * NODE_STREAM_LAYOUT_LINK_STRIDE layout-link rows (this node's OWN outbound layout-links —
 * see buffer-decode.ts's DecodedNodeStreamFrame doc comment). */
export const BUF_NODE_STREAM_FRAME_HEADER_SIZE = 20;

/** Byte width of ONE layout-link row within a node stream frame:
 * [DstNodeRow:i32][EdgeRow:i32]. Narrower than LAYOUT_LINK_STRIDE (the combined
 * LayoutLink block's 12-byte row, SrcNodeRow+DstNodeRow+EdgeRow) because on a per-node
 * stream the source IS this node — implicit from the fd position / aggregate row index. */
export const NODE_STREAM_LAYOUT_LINK_STRIDE = 8;

/** Byte layout of one node's INTERIOR per-fd frame (Buffer.BuildInteriorStreamFrame), no
 * outer tag: [tick:u32] followed by a FIXED INTERIOR_SLOTS_PER_NODE × INTERIOR_STRIDE bytes
 * (no count — the decoder derives the length from the fixed per-node slot count, same as
 * the combined Interior block). */
export const BUF_INTERIOR_STREAM_FRAME_HEADER_SIZE = 4;
