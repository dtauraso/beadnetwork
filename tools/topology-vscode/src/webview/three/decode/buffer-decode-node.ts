// buffer-decode-node.ts — decodeNodeStreamFrame: one node's dedicated per-fd NODE_STREAM
// frame, plus nodeLabel, the string reader over the aggregate DecodedNodeFrame that
// node-stream-blocks.ts assembles from those per-node frames.

import {
  NODE_STRIDE,
  CHAIN_BEAD_STRIDE,
  readNodeLabelOff,
  readNodeLabelLen,
  readNodeNodeId,
} from "../../../schema/buffer-layout";
import { BUF_NODE_STREAM_FRAME_HEADER_SIZE } from "../../../schema/frame-tags";
import { STR_DECODER, decodeTrailingEvents } from "./buffer-decode-shared";

// reportNodeIdMismatch is deliberately NOT a static import of ../log/post: that module
// transitively imports vscode-api.ts, which reads `window` unconditionally at module-eval
// time (acquireVsCodeApi() caching) — fine in the real webview, but this decoder is also
// imported directly by node-environment unit tests (stream-fixture.test.ts) with no
// `window` global, and a static import would crash the whole test file just for importing
// the decoder. Dynamic-import it, gated on `window` existing, so the webview still gets a
// real postLog("load-error", …) (routed to .probe/ts-errors.jsonl) while a headless test
// gets a plain console.error instead — same reasoning as postLog's own window guard.
function reportNodeIdMismatch(row: number, expectedId: number, statedId: number): void {
  const message = `node stream frame arrived on row ${row} (expected id ${expectedId}) but carries NodeId ${statedId}`;
  if (typeof window === "undefined") {
    // eslint-disable-next-line no-console
    console.error(`[wirefold] node-id-row-mismatch: ${message}`);
    return;
  }
  void import("../../log/post").then(({ postLog }) => {
    postLog("load-error", { reason: "node-id-row-mismatch", message, arrivalRow: row, statedNodeId: statedId, expectedNodeId: expectedId });
  });
}

/** Aggregate view assembled from per-node NODE_STREAM frames (see node-stream-blocks.ts):
 *  the Node/Interior blocks + Label bytes — the node-owner-group blocks, which share one
 *  owner (the node movers). No Port block any more (docs/channels-not-ports.md — a port
 *  carries no geometry, so there is nothing to aggregate). */
export interface DecodedNodeFrame {
  tick: number;
  nodeCount: number;
  /** DataView over the node block only; byteLength = nodeCount × NODE_STRIDE. */
  nodeView: DataView;
  /** Interior grid rows (nodeCount × INTERIOR_SLOTS_PER_NODE); row = nodeRow*slots + slot. */
  interiorCount: number;
  /** DataView over the interior block; byteLength = interiorCount × INTERIOR_STRIDE. */
  interiorView: DataView;
  /** Total bytes in the trailing label section (self-sizing via the header labelBytesCount). */
  labelBytesCount: number;
  /** Uint8 view over the label-bytes section: every node's label UTF-8 bytes concatenated in
   *  node-row order. A node's label is labelBytes[LabelOff : LabelOff+LabelLen) — see nodeLabel. */
  labelBytes: Uint8Array;
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

// There is no portName() any more (docs/channels-not-ports.md): a port has no name/row of
// its own on the buffer, so there is nothing to slice out of a port-name-bytes section.

/** Decoded view over ONE node's dedicated per-fd NODE-stream frame (BUF_BLOCK_TAG_NODE_STREAM
 *  — see frame-tags.ts's BUF_NODE_STREAM_FRAME_HEADER_SIZE doc comment for the byte layout):
 *  [tick:u32][labelLen:u32][chainBeadCount:u32] +
 *  this node's own single NODE_STRIDE row (index 0) + its own inline label bytes + its own
 *  chain-bead rows. No Port section any more
 *  (docs/channels-not-ports.md — a port carries no geometry, so there is no port-row/
 *  port-name-bytes section to size or decode). */
export interface DecodedNodeStreamFrame {
  tick: number;
  /** DataView over this node's single Node row; byteLength = NODE_STRIDE. */
  nodeView: DataView;
  /** This node's own label, decoded straight from its inline bytes (LabelOff is always 0
   *  into THIS frame's own bytes — unlike the combined Node block's shared label section). */
  label: string;
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
  if (decoded) {
    // Identity check made possible by the NodeId column (task/row-fd-identity-parity):
    // before this column existed, a node's identity WAS the row it arrived on — a frame
    // could never contradict where it landed, so a permutation would render silently in the
    // wrong place. ROW ID = NODE ID - 1 is enforced at Go's load time
    // (persistence-ownership.md) and packed alongside the row by the SAME mover that owns
    // it (node_mover.go), so a disagreement here means the frame arrived on the wrong
    // dedicated fd — a bridge-plumbing bug, not a topology error. Report loudly rather than
    // silently trusting the row, the way every other row-vs-content mismatch in this file
    // degrades to "".
    const statedId = readNodeNodeId(decoded.nodeView, 0);
    const expectedId = row + 1;
    if (statedId !== expectedId) {
      reportNodeIdMismatch(row, expectedId, statedId);
    }
  }
  lastNodeStreamBufByRow.set(row, buf);
  lastDecodedNodeStreamByRow.set(row, decoded);
  return decoded;
}

function decodeNodeStreamFrameUncached(buf: ArrayBuffer): DecodedNodeStreamFrame | null {
  if (buf.byteLength < BUF_NODE_STREAM_FRAME_HEADER_SIZE) return null;
  const hdr = new DataView(buf, 0, BUF_NODE_STREAM_FRAME_HEADER_SIZE);
  const tick               = hdr.getUint32(0,  true);
  const labelLen           = hdr.getUint32(4,  true);
  const chainBeadCount     = hdr.getUint32(8,  true);

  const chainBeadBytes = chainBeadCount * CHAIN_BEAD_STRIDE;
  const expectedLen = BUF_NODE_STREAM_FRAME_HEADER_SIZE + NODE_STRIDE + labelLen + chainBeadBytes;
  if (buf.byteLength < expectedLen) return null;

  let off = BUF_NODE_STREAM_FRAME_HEADER_SIZE;
  const nodeView = new DataView(buf, off, NODE_STRIDE);
  off += NODE_STRIDE;

  const labelBytes = new Uint8Array(buf, off, labelLen);
  const label = STR_DECODER.decode(labelBytes);
  off += labelLen;

  const chainBeadView = new DataView(buf, off, chainBeadBytes);
  off += chainBeadBytes;

  const { count: eventCount, view: eventView, textView: eventTextView } = decodeTrailingEvents(buf, off);

  return { tick, nodeView, label, chainBeadCount, chainBeadView, eventCount, eventView, eventTextView };
}
