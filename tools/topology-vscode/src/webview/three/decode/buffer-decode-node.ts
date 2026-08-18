import {
  NODE_STRIDE,
  TILT_ARROW_STRIDE,
  readNodeLabelOff,
  readNodeLabelLen,
  readNodeNodeId,
} from "../../../../Buffer/buffer-layout";
import { BUF_NODE_STREAM_FRAME_HEADER_SIZE } from "../../../../Buffer/frame-tags";
import { STR_DECODER, decodeTrailingEvents } from "./buffer-decode-shared";

const reportedNodeIdMismatches = new Set<string>();

function reportNodeIdMismatch(row: number, expectedId: number, statedId: number): void {
  const seenKey = `${row}:${statedId}`;
  if (reportedNodeIdMismatches.has(seenKey)) return;
  reportedNodeIdMismatches.add(seenKey);

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

let reportedShortFrame = false;

function reportShortNodeFrame(got: number, expected: number, labelLen: number): void {
  if (reportedShortFrame) return;
  reportedShortFrame = true;

  const message =
    `node stream frame is ${got} bytes but this webview's layout needs ${expected} ` +
    `(header + NODE_STRIDE ${NODE_STRIDE} + label ${labelLen}). ` +
    `Go and the webview are built ` +
    `against different buffer layouts, so EVERY node frame is being dropped and nothing on a ` +
    `node will update. Run "Developer: Reload Window" — reopening the file reloads only the ` +
    `webview, not the extension host that spawns Go.`;
  if (typeof window === "undefined") {
    // eslint-disable-next-line no-console
    console.error(`[wirefold] node-frame-layout-skew: ${message}`);
    return;
  }
  void import("../../log/post").then(({ postLog }) => {
    postLog("load-error", { reason: "node-frame-layout-skew", message, gotBytes: got, expectedBytes: expected });
  });
}

export interface DecodedNodeFrame {
  tick: number;
  nodeCount: number;

  tiltArrowCount: number;

  tiltArrowView: DataView;

  nodeView: DataView;

  interiorCount: number;

  interiorView: DataView;

  labelBytesCount: number;

  labelBytes: Uint8Array;
}

export function nodeLabel(decoded: DecodedNodeFrame, row: number): string {

  if (row < 0 || row >= decoded.nodeCount) return "";
  const off = readNodeLabelOff(decoded.nodeView, row);
  const len = readNodeLabelLen(decoded.nodeView, row);
  if (len === 0) return "";
  if (off < 0 || len < 0 || off + len > decoded.labelBytes.byteLength) return "";
  return STR_DECODER.decode(decoded.labelBytes.subarray(off, off + len));
}

export interface DecodedNodeStreamFrame {
  tick: number;

  nodeView: DataView;

  label: string;

  tiltArrowCount: number;
  tiltArrowView: DataView;

  eventCount: number;
  eventView: DataView;
  eventTextView: DataView;
}

const lastNodeStreamBufByRow = new Map<number, ArrayBuffer>();
const lastDecodedNodeStreamByRow = new Map<number, DecodedNodeStreamFrame | null>();

export function decodeNodeStreamFrame(row: number, buf: ArrayBuffer): DecodedNodeStreamFrame | null {
  if (lastNodeStreamBufByRow.get(row) === buf) {
    return lastDecodedNodeStreamByRow.get(row) ?? null;
  }
  const decoded = decodeNodeStreamFrameUncached(buf);
  if (decoded) {

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
  const tiltArrowCount     = hdr.getUint32(8,  true);

  const expectedLen = BUF_NODE_STREAM_FRAME_HEADER_SIZE + NODE_STRIDE + labelLen
    + tiltArrowCount * TILT_ARROW_STRIDE;
  if (buf.byteLength < expectedLen) {
    reportShortNodeFrame(buf.byteLength, expectedLen, labelLen);
    return null;
  }

  let off = BUF_NODE_STREAM_FRAME_HEADER_SIZE;
  const nodeView = new DataView(buf, off, NODE_STRIDE);
  off += NODE_STRIDE;

  const labelBytes = new Uint8Array(buf, off, labelLen);
  const label = STR_DECODER.decode(labelBytes);
  off += labelLen;

  const tiltArrowView = new DataView(buf, off, tiltArrowCount * TILT_ARROW_STRIDE);
  off += tiltArrowCount * TILT_ARROW_STRIDE;

  const { count: eventCount, view: eventView, textView: eventTextView } = decodeTrailingEvents(buf, off);

  return { tick, nodeView, label, tiltArrowCount, tiltArrowView, eventCount, eventView, eventTextView };
}
