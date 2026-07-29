// buffer-decode.ts — pure view/stream-frame decoders.
//
// decodeViewFrame takes the dedicated VIEW-stream frame and returns the Camera/Overlay/
// Scene blocks. decodeEdgeStreamFrame/decodeNodeStreamFrame/decodeInteriorStreamFrame
// decode each per-owner (per-edge / per-node) stream frame. There is no combined
// BEAD/NODE/EDGE frame decoder — the central accumulator that used to write that combined
// frame was deleted entirely — memory/feedback_no_single_writer_bridge.md's final step;
// WIREFOLD_STREAM_FDS is mandatory. `DecodedNodeFrame`/`DecodedEdgeFrame` below remain as
// the AGGREGATE types the per-node/per-edge stream frames are assembled into (see
// node-stream-blocks.ts's aggregator and buffer-log.ts's decodeStreamFrameEvents) — they
// are no longer decoded directly off a single combined wire frame.

import {
  NODE_STRIDE,
  CHAIN_BEAD_STRIDE,
  INTERIOR_STRIDE,
  INTERIOR_SLOTS_PER_NODE,
  EDGE_STRIDE,
  PORT_STRIDE,
  CAMERA_STRIDE,
  OVERLAY_STRIDE,
  SCENE_STRIDE,
  EVENT_STRIDE,
  readNodeLabelOff,
  readNodeLabelLen,
  readPortPortNameOff,
  readPortPortNameLen,
  readEdgeEdgeLabelOff,
  readEdgeEdgeLabelLen,
} from "../../schema/buffer-layout";
import { BUF_VIEW_FRAME_HEADER_SIZE, BUF_EDGE_STREAM_FRAME_HEADER_SIZE, BUF_NODE_STREAM_FRAME_HEADER_SIZE, BUF_INTERIOR_STREAM_FRAME_HEADER_SIZE, NODE_STREAM_LAYOUT_LINK_STRIDE } from "../../schema/frame-tags";
// Generated (part of BUF_LAYOUT_FINGERPRINT) — re-exported here so existing consumers
// (buffer-scene.tsx, InteriorBeadInstances.tsx, buffer-log.ts) keep importing it from the
// decode module rather than reaching into schema/buffer-layout directly.
export { INTERIOR_SLOTS_PER_NODE } from "../../schema/buffer-layout";

/** Shared UTF-8 decoder for the label / port-name / edge-label sections. */
const STR_DECODER = new TextDecoder();

/** Decodes a trailing EVENTS section ([count:u32] + count × EVENT_STRIDE rows) appended
 *  after `offset` bytes of already-known content in ANY per-owner frame (NODE/EDGE/
 *  INTERIOR/VIEW — memory/feedback_no_single_writer_bridge.md). The view/scene frame no
 *  longer carries an EVENT block at all — each per-owner stream carries its own instead.
 *  Returns {count:0, view: empty} when the buffer is too short to hold even the count
 *  (never null — callers can always safely iterate 0 times). */
// decodeTrailingEvents decodes [count:u32] + count × EVENT_STRIDE event rows, plus the
// single sanctioned free-form event-text-bytes section that follows immediately after
// (see Buffer.BuildEventsSection — TextOff/TextLen on each event row address into
// textView). textView spans whatever bytes remain to the end of the frame; a frame with
// no breadcrumb events has an empty (but valid) textView.
export function decodeTrailingEvents(buf: ArrayBuffer, offset: number): { count: number; view: DataView; textView: DataView } {
  const empty = { count: 0, view: new DataView(buf, buf.byteLength, 0), textView: new DataView(buf, buf.byteLength, 0) };
  if (buf.byteLength < offset + 4) return empty;
  const count = new DataView(buf, offset, 4).getUint32(0, true);
  const bytes = count * EVENT_STRIDE;
  if (buf.byteLength < offset + 4 + bytes) return empty;
  const textStart = offset + 4 + bytes;
  return {
    count,
    view: new DataView(buf, offset + 4, bytes),
    textView: new DataView(buf, textStart, buf.byteLength - textStart),
  };
}

/** Aggregate view assembled from per-node NODE_STREAM frames (see node-stream-blocks.ts):
 *  the Node/Interior/Port blocks + Label/PortName bytes — the node-owner-group blocks,
 *  which share one owner (the node movers). */
export interface DecodedNodeFrame {
  tick: number;
  nodeCount: number;
  /** DataView over the node block only; byteLength = nodeCount × NODE_STRIDE. */
  nodeView: DataView;
  /** Interior grid rows (nodeCount × INTERIOR_SLOTS_PER_NODE); row = nodeRow*slots + slot. */
  interiorCount: number;
  /** DataView over the interior block; byteLength = interiorCount × INTERIOR_STRIDE. */
  interiorView: DataView;
  /** Total port rows across all nodes (self-sizing via the header portCount field). */
  portCount: number;
  /** DataView over the port block only; byteLength = portCount × PORT_STRIDE. Row i is the
   *  buffer port row i — the same index a port InstancedMesh instanceId carries for picking. */
  portView: DataView;
  /** Total bytes in the trailing label section (self-sizing via the header labelBytesCount). */
  labelBytesCount: number;
  /** Uint8 view over the label-bytes section: every node's label UTF-8 bytes concatenated in
   *  node-row order. A node's label is labelBytes[LabelOff : LabelOff+LabelLen) — see nodeLabel. */
  labelBytes: Uint8Array;
  /** Uint8 view over the port-name-bytes section (flattened port-row order). See portName. */
  portNameBytes: Uint8Array;
}

/**
 * Human label for buffer node row `row`: slice [LabelOff, LabelOff+LabelLen) out of the
 * decoded label-bytes section and UTF-8 decode it. Returns "" when the node has no label
 * (LabelLen == 0). Pure — reads only the decoded node frame. This is the row-keyed
 * replacement for the removed id/label sidecar: the label rides the binary buffer.
 */
export function nodeLabel(decoded: DecodedNodeFrame, row: number): string {
  // Bound the row against THIS frame's node count BEFORE indexing the node block:
  // reading the off/len columns at row×NODE_STRIDE throws (nodeView is exactly
  // nodeCount×NODE_STRIDE bytes) when a VIEW-bucket event carries a node row valid for
  // the topology but beyond a STALE cached per-node stream frame's count (a cross-generation
  // skew inherent to per-owner streaming). Degrade to "" — the graceful-empty contract
  // this function's callers already document (decodeBufferLog, buffer-log.ts).
  if (row < 0 || row >= decoded.nodeCount) return "";
  const off = readNodeLabelOff(decoded.nodeView, row);
  const len = readNodeLabelLen(decoded.nodeView, row);
  if (len === 0) return "";
  if (off < 0 || len < 0 || off + len > decoded.labelBytes.byteLength) return "";
  return STR_DECODER.decode(decoded.labelBytes.subarray(off, off + len));
}

/**
 * Port name for buffer port row `row`: slice out of the decoded port-name-bytes section.
 * Returns "" when the port has no name. Used only by the buffer-decoded .probe logger — the
 * render/bridge path resolves a port hit by row index, never by this string.
 */
export function portName(decoded: DecodedNodeFrame, row: number): string {
  // Upper-bound the row too (not just row<0): a stale cached frame can have fewer port
  // rows than the topology, and reading row×PORT_STRIDE past portView throws. Same
  // graceful-empty contract as nodeLabel.
  if (row < 0 || row >= decoded.portCount) return "";
  const off = readPortPortNameOff(decoded.portView, row);
  const len = readPortPortNameLen(decoded.portView, row);
  if (len === 0) return "";
  if (off < 0 || len < 0 || off + len > decoded.portNameBytes.byteLength) return "";
  return STR_DECODER.decode(decoded.portNameBytes.subarray(off, off + len));
}

/** Aggregate view assembled from per-edge EDGE_STREAM frames: the Edge block + EdgeLabel
 *  bytes — the Edge block carries NO endpoint coordinates; it references its two port rows
 *  (SrcPortRow/DstPortRow), resolved against the SAME TICK's node-frame Port block (see
 *  EdgeTube.tsx). */
export interface DecodedEdgeFrame {
  tick: number;
  edgeCount: number;
  /** DataView over the edge block only; byteLength = edgeCount × EDGE_STRIDE. */
  edgeView: DataView;
  /** Uint8 view over the edge-label-bytes section (edge-row order). See edgeLabel. */
  edgeLabelBytes: Uint8Array;
}

/** Decoded view over ONE edge's dedicated per-fd stream frame (BUF_BLOCK_TAG_EDGE_STREAM —
 *  see frame-tags.ts's BUF_EDGE_STREAM_FRAME_HEADER_SIZE doc comment for the byte layout):
 *  [tick:u32] + one EDGE_STRIDE row (this edge's own SrcPortRow/DstPortRow/Selected) + this
 *  edge's own label bytes (inline, not a shared section) +
 *  its trailing EVENTS section. No bead rows: the Bead block is gone with the moving bead. */
export interface DecodedEdgeStreamFrame {
  tick: number;
  /** DataView over the single Edge row (row 0); byteLength = EDGE_STRIDE. */
  edgeView: DataView;
  /** This edge's own label, decoded straight from its inline bytes (no shared section / no
   *  Off into a foreign frame — unlike the combined Edge block's EdgeLabelOff/Len). */
  label: string;
  // No beadCount/beadView: the Bead block is gone with the moving bead it carried. A
  // traversal renders as the LIT bead of the source node's own chain — docs/beads-are-the-edge.md.
  /** This edge's own trailing EVENTS section (.probe log only; see decodeTrailingEvents). */
  eventCount: number;
  eventView: DataView;
  eventTextView: DataView;
}

// Per-edge-row memo (keyed by row, not a single lastBuf — many edge streams arrive
// concurrently, one per fd, so a single-entry memo would thrash across rows every frame).
const lastEdgeStreamBufByRow = new Map<number, ArrayBuffer>();
const lastDecodedEdgeStreamByRow = new Map<number, DecodedEdgeStreamFrame | null>();

/**
 * Decode ONE edge row's BUF_BLOCK_TAG_EDGE_STREAM frame ArrayBuffer into a typed view.
 * Returns null if the buffer is too small to be a valid frame. Pure — no side effects
 * beyond this function's own per-row memo. Views alias the original buffer (zero-copy).
 */
export function decodeEdgeStreamFrame(row: number, buf: ArrayBuffer): DecodedEdgeStreamFrame | null {
  if (lastEdgeStreamBufByRow.get(row) === buf) {
    return lastDecodedEdgeStreamByRow.get(row) ?? null;
  }
  const decoded = decodeEdgeStreamFrameUncached(buf);
  lastEdgeStreamBufByRow.set(row, buf);
  lastDecodedEdgeStreamByRow.set(row, decoded);
  return decoded;
}

function decodeEdgeStreamFrameUncached(buf: ArrayBuffer): DecodedEdgeStreamFrame | null {
  if (buf.byteLength < BUF_EDGE_STREAM_FRAME_HEADER_SIZE + EDGE_STRIDE) return null;
  const hdr = new DataView(buf, 0, 4);
  const tick = hdr.getUint32(0, true);

  let off = BUF_EDGE_STREAM_FRAME_HEADER_SIZE;
  const edgeView = new DataView(buf, off, EDGE_STRIDE);
  off += EDGE_STRIDE;

  // EdgeLabelLen lives on the Edge row itself (readEdgeEdgeLabelLen) — EdgeLabelOff is
  // always 0 on a dedicated per-edge stream (this frame's own label bytes immediately
  // follow the row, no shared section — see Buffer/edge_stream_frame.go).
  const labelLen = readEdgeEdgeLabelLen(edgeView, 0);
  if (buf.byteLength < off + labelLen) return null;
  const labelBytes = new Uint8Array(buf, off, labelLen);
  const label = STR_DECODER.decode(labelBytes);
  off += labelLen;

  const { count: eventCount, view: eventView, textView: eventTextView } = decodeTrailingEvents(buf, off);

  return { tick, edgeView, label, eventCount, eventView, eventTextView };
}

/** Decoded view over a BUF_BLOCK_TAG_VIEW frame (see frame-tags.ts for its byte layout):
 *  [tick:u32] followed by the Camera, Overlay, and Scene blocks. */
export interface DecodedViewFrame {
  tick: number;
  cameraView: DataView;
  overlayView: DataView;
  sceneView: DataView;
  /** This VIEW stream's own trailing EVENTS section (camera/overlay/scene events —
   *  every other kind is decentralized to its own owner fd). */
  eventCount: number;
  eventView: DataView;
  eventTextView: DataView;
}

// Single-entry memo, mirroring the other per-frame decoders below — the view frame
// arrives on its own dedicated fd, decoded independently of every other stream.
let lastViewBuf: ArrayBuffer | null = null;
let lastDecodedView: DecodedViewFrame | null = null;

/**
 * Decode a BUF_BLOCK_TAG_VIEW frame ArrayBuffer (the dedicated view-fd stream) into
 * typed camera/overlay/scene views. Returns null if the buffer is too small to be a
 * valid view frame. Pure — no side effects, no store reads/writes. Views alias the
 * original buffer (zero-copy). Memoized on `buf`'s identity.
 */
export function decodeViewFrame(buf: ArrayBuffer): DecodedViewFrame | null {
  if (buf === lastViewBuf) return lastDecodedView;
  const decoded = decodeViewFrameUncached(buf);
  lastViewBuf = buf;
  lastDecodedView = decoded;
  return decoded;
}

function decodeViewFrameUncached(buf: ArrayBuffer): DecodedViewFrame | null {
  const expectedLen = BUF_VIEW_FRAME_HEADER_SIZE + CAMERA_STRIDE + OVERLAY_STRIDE + SCENE_STRIDE;
  if (buf.byteLength < expectedLen) return null;

  const tick = new DataView(buf, 0, BUF_VIEW_FRAME_HEADER_SIZE).getUint32(0, true);
  let off = BUF_VIEW_FRAME_HEADER_SIZE;

  const cameraView = new DataView(buf, off, CAMERA_STRIDE);
  off += CAMERA_STRIDE;

  const overlayView = new DataView(buf, off, OVERLAY_STRIDE);
  off += OVERLAY_STRIDE;

  const sceneView = new DataView(buf, off, SCENE_STRIDE);
  off += SCENE_STRIDE;

  const { count: eventCount, view: eventView, textView: eventTextView } = decodeTrailingEvents(buf, off);

  return { tick, cameraView, overlayView, sceneView, eventCount, eventView, eventTextView };
}

/**
 * Edge label for buffer edge row `row`: slice out of the decoded edge-label-bytes section.
 * Returns "" when the edge has no label. Used only by the buffer-decoded .probe logger — the
 * render/bridge path resolves an edge hit by row index, never by this string.
 */
export function edgeLabel(decoded: DecodedEdgeFrame, row: number): string {
  // Upper-bound the row too (see nodeLabel/portName): a stale cached edge frame can have
  // fewer rows than the topology; reading row×EDGE_STRIDE past edgeView throws.
  if (row < 0 || row >= decoded.edgeCount) return "";
  const off = readEdgeEdgeLabelOff(decoded.edgeView, row);
  const len = readEdgeEdgeLabelLen(decoded.edgeView, row);
  if (len === 0) return "";
  if (off < 0 || len < 0 || off + len > decoded.edgeLabelBytes.byteLength) return "";
  return STR_DECODER.decode(decoded.edgeLabelBytes.subarray(off, off + len));
}

/** Decoded view over ONE node's dedicated per-fd NODE-stream frame (BUF_BLOCK_TAG_NODE_STREAM
 *  — see frame-tags.ts's BUF_NODE_STREAM_FRAME_HEADER_SIZE doc comment for the byte layout):
 *  [tick:u32][portCount:u32][labelLen:u32][portNameBytesCount:u32][layoutLinkCount:u32]
 *  [chainBeadCount:u32] +
 *  this node's own single NODE_STRIDE row (index 0) + its own inline label bytes + its own
 *  Port rows (each row's NodeRow column already the global node row) + its own inline
 *  port-name bytes + its own outbound cascade-link rows (this node is always the
 *  lexicographically-smaller endpoint — see node_mover.go's cascadeEdges doc comment;
 *  each row is [DstNodeRow:i32], NODE_STREAM_LAYOUT_LINK_STRIDE bytes — no edge-row: the
 *  overlay draws node-center to node-center, never along a bead edge). */
export interface DecodedNodeStreamFrame {
  tick: number;
  /** DataView over this node's single Node row; byteLength = NODE_STRIDE. */
  nodeView: DataView;
  /** This node's own label, decoded straight from its inline bytes (LabelOff is always 0
   *  into THIS frame's own bytes — unlike the combined Node block's shared label section). */
  label: string;
  portCount: number;
  /** DataView over this node's own port rows; byteLength = portCount × PORT_STRIDE. */
  portView: DataView;
  /** Uint8 view over this node's own port-name bytes (flattened port-row order). */
  portNameBytes: Uint8Array;
  layoutLinkCount: number;
  /** DataView over this node's own outbound LayoutLink rows; byteLength = layoutLinkCount
   *  × NODE_STREAM_LAYOUT_LINK_STRIDE. Read with readNodeStreamLayoutLinkDstNodeRow
   *  below — this node's own row is the SrcNodeRow. */
  layoutLinkView: DataView;
  /** Number of this node's own chain-bead rows (docs/beads-are-the-edge.md) — the
   *  placeholder sequence that is the VISUAL of a traversal along its outgoing edges,
   *  concatenated across those edges in order. Not a count of anything on a channel: the
   *  node-to-node channels are the real connection and are never drawn. */
  chainBeadCount: number;
  /** DataView over this node's own ChainBead rows; byteLength = chainBeadCount ×
   *  CHAIN_BEAD_STRIDE. Offsets are NODE-LOCAL — add this node's own center to
   *  get a world position, the same convention as the Interior block. Go owns the offsets. */
  chainBeadView: DataView;
  /** This node's own trailing EVENTS section (.probe log only; see decodeTrailingEvents). */
  eventCount: number;
  eventView: DataView;
  eventTextView: DataView;
}

// No bespoke chain-bead row reader here: a chain-bead row on a node stream is byte-identical
// to a ChainBead BLOCK row, so the GENERATED readChainBeadOX/OY/OZ read these rows directly. A
// hand-rolled reader would be a second decoder for the same bytes, and would leave the
// generated ones with no production consumer (check-no-dead-buffer-column.sh).

/** Reads DstNodeRow (i32) from row `row` of a node stream frame's LayoutLink section. */
export function readNodeStreamLayoutLinkDstNodeRow(view: DataView, row: number): number {
  return view.getInt32(row * NODE_STREAM_LAYOUT_LINK_STRIDE, true);
}

// Per-node-row memo (keyed by row), mirroring decodeEdgeStreamFrame's per-row memo.
const lastNodeStreamBufByRow = new Map<number, ArrayBuffer>();
const lastDecodedNodeStreamByRow = new Map<number, DecodedNodeStreamFrame | null>();

/**
 * Decode ONE node row's BUF_BLOCK_TAG_NODE_STREAM frame ArrayBuffer into a typed view.
 * Returns null if the buffer is too small to be a valid frame. Pure — no side effects
 * beyond this function's own per-row memo. Views alias the original buffer (zero-copy).
 */
export function decodeNodeStreamFrame(row: number, buf: ArrayBuffer): DecodedNodeStreamFrame | null {
  if (lastNodeStreamBufByRow.get(row) === buf) {
    return lastDecodedNodeStreamByRow.get(row) ?? null;
  }
  const decoded = decodeNodeStreamFrameUncached(buf);
  lastNodeStreamBufByRow.set(row, buf);
  lastDecodedNodeStreamByRow.set(row, decoded);
  return decoded;
}

function decodeNodeStreamFrameUncached(buf: ArrayBuffer): DecodedNodeStreamFrame | null {
  if (buf.byteLength < BUF_NODE_STREAM_FRAME_HEADER_SIZE) return null;
  const hdr = new DataView(buf, 0, BUF_NODE_STREAM_FRAME_HEADER_SIZE);
  const tick               = hdr.getUint32(0,  true);
  const portCount          = hdr.getUint32(4,  true);
  const labelLen           = hdr.getUint32(8,  true);
  const portNameBytesCount = hdr.getUint32(12, true);
  const layoutLinkCount    = hdr.getUint32(16, true);
  const chainBeadCount     = hdr.getUint32(20, true);

  const portBytes = portCount * PORT_STRIDE;
  const layoutLinkBytes = layoutLinkCount * NODE_STREAM_LAYOUT_LINK_STRIDE;
  const chainBeadBytes = chainBeadCount * CHAIN_BEAD_STRIDE;
  const expectedLen = BUF_NODE_STREAM_FRAME_HEADER_SIZE + NODE_STRIDE + labelLen + portBytes + portNameBytesCount + layoutLinkBytes + chainBeadBytes;
  if (buf.byteLength < expectedLen) return null;

  let off = BUF_NODE_STREAM_FRAME_HEADER_SIZE;
  const nodeView = new DataView(buf, off, NODE_STRIDE);
  off += NODE_STRIDE;

  const labelBytes = new Uint8Array(buf, off, labelLen);
  const label = STR_DECODER.decode(labelBytes);
  off += labelLen;

  const portView = new DataView(buf, off, portBytes);
  off += portBytes;

  const portNameBytes = new Uint8Array(buf, off, portNameBytesCount);
  off += portNameBytesCount;

  const layoutLinkView = new DataView(buf, off, layoutLinkBytes);
  off += layoutLinkBytes;

  const chainBeadView = new DataView(buf, off, chainBeadBytes);
  off += chainBeadBytes;

  const { count: eventCount, view: eventView, textView: eventTextView } = decodeTrailingEvents(buf, off);

  return { tick, nodeView, label, portCount, portView, portNameBytes, layoutLinkCount, layoutLinkView, chainBeadCount, chainBeadView, eventCount, eventView, eventTextView };
}

/** Decoded view over ONE node's dedicated per-fd INTERIOR-stream frame
 *  (BUF_BLOCK_TAG_INTERIOR_STREAM): [tick:u32] followed by a FIXED
 *  INTERIOR_SLOTS_PER_NODE × INTERIOR_STRIDE bytes (that node's own interior-bead grid). */
export interface DecodedInteriorStreamFrame {
  tick: number;
  /** DataView over this node's own INTERIOR_SLOTS_PER_NODE interior rows. */
  interiorView: DataView;
  /** This goroutine's own trailing EVENTS section (.probe log only; see decodeTrailingEvents). */
  eventCount: number;
  eventView: DataView;
  eventTextView: DataView;
}

const lastInteriorStreamBufByRow = new Map<number, ArrayBuffer>();
const lastDecodedInteriorStreamByRow = new Map<number, DecodedInteriorStreamFrame | null>();

/**
 * Decode ONE node row's BUF_BLOCK_TAG_INTERIOR_STREAM frame ArrayBuffer into a typed view.
 * Returns null if the buffer is too small to be a valid frame. Pure, per-row memoized.
 */
export function decodeInteriorStreamFrame(row: number, buf: ArrayBuffer): DecodedInteriorStreamFrame | null {
  if (lastInteriorStreamBufByRow.get(row) === buf) {
    return lastDecodedInteriorStreamByRow.get(row) ?? null;
  }
  const decoded = decodeInteriorStreamFrameUncached(buf);
  lastInteriorStreamBufByRow.set(row, buf);
  lastDecodedInteriorStreamByRow.set(row, decoded);
  return decoded;
}

function decodeInteriorStreamFrameUncached(buf: ArrayBuffer): DecodedInteriorStreamFrame | null {
  const interiorBytes = INTERIOR_SLOTS_PER_NODE * INTERIOR_STRIDE;
  const expectedLen = BUF_INTERIOR_STREAM_FRAME_HEADER_SIZE + interiorBytes;
  if (buf.byteLength < expectedLen) return null;
  const tick = new DataView(buf, 0, BUF_INTERIOR_STREAM_FRAME_HEADER_SIZE).getUint32(0, true);
  const interiorView = new DataView(buf, BUF_INTERIOR_STREAM_FRAME_HEADER_SIZE, interiorBytes);
  const { count: eventCount, view: eventView, textView: eventTextView } = decodeTrailingEvents(buf, expectedLen);
  return { tick, interiorView, eventCount, eventView, eventTextView };
}
