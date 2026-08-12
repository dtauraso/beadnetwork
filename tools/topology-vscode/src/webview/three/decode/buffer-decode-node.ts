



import {
  NODE_STRIDE,
  CHAIN_BEAD_STRIDE,
  readNodeLabelOff,
  readNodeLabelLen,
  readNodeNodeId,
} from "../../../schema/buffer-layout";
import { BUF_NODE_STREAM_FRAME_HEADER_SIZE } from "../../../schema/frame-tags";
import { STR_DECODER, decodeTrailingEvents } from "./buffer-decode-shared";









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


export interface DecodedNodeFrame {
  tick: number;
  nodeCount: number;

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

  chainBeadCount: number;

  chainBeadView: DataView;

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
