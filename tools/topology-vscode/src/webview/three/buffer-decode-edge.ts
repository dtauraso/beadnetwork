// buffer-decode-edge.ts — decodeEdgeStreamFrame: one edge's dedicated per-fd EDGE_STREAM
// frame, plus edgeLabel, the string reader over the aggregate DecodedEdgeFrame that
// edge-stream-blocks.ts assembles from those per-edge frames.

import { EDGE_STRIDE, readEdgeEdgeLabelOff, readEdgeEdgeLabelLen } from "../../schema/buffer-layout";
import { BUF_EDGE_STREAM_FRAME_HEADER_SIZE } from "../../schema/frame-tags";
import { STR_DECODER, decodeTrailingEvents } from "./buffer-decode-shared";

/** Aggregate view assembled from per-edge EDGE_STREAM frames: the Edge block + EdgeLabel
 *  bytes. The Edge block carries its own SEGMENT (SX..EZ) directly — node surface to node
 *  surface (docs/channels-not-ports.md) — not a reference through a port row. */
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
 *  [tick:u32] + one EDGE_STRIDE row (this edge's own SX..EZ segment/Selected) + this edge's
 *  own label bytes (inline, not a shared section) + its trailing EVENTS section. No bead
 *  rows: the Bead block is gone with the moving bead. */
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

/**
 * Edge label for buffer edge row `row`: slice out of the decoded edge-label-bytes section.
 * Returns "" when the edge has no label. Used only by the buffer-decoded .probe logger — the
 * render/bridge path resolves an edge hit by row index, never by this string.
 */
export function edgeLabel(decoded: DecodedEdgeFrame, row: number): string {
  // Upper-bound the row too (see nodeLabel): a stale cached edge frame can have fewer
  // rows than the topology; reading row×EDGE_STRIDE past edgeView throws.
  if (row < 0 || row >= decoded.edgeCount) return "";
  const off = readEdgeEdgeLabelOff(decoded.edgeView, row);
  const len = readEdgeEdgeLabelLen(decoded.edgeView, row);
  if (len === 0) return "";
  if (off < 0 || len < 0 || off + len > decoded.edgeLabelBytes.byteLength) return "";
  return STR_DECODER.decode(decoded.edgeLabelBytes.subarray(off, off + len));
}
