import { EDGE_STRIDE, readEdgeEdgeLabelOff, readEdgeEdgeLabelLen } from "../../../../Buffer/buffer-layout";
import { BUF_EDGE_STREAM_FRAME_HEADER_SIZE } from "../../../../Buffer/frame-tags";
import { STR_DECODER, decodeTrailingEvents } from "./buffer-decode-shared";

export interface DecodedEdgeFrame {
  tick: number;
  edgeCount: number;

  edgeView: DataView;

  edgeLabelBytes: Uint8Array;
}

export interface DecodedEdgeStreamFrame {
  tick: number;

  edgeView: DataView;

  label: string;

  eventCount: number;
  eventView: DataView;
  eventTextView: DataView;
}

const lastEdgeStreamBufByRow = new Map<number, ArrayBuffer>();
const lastDecodedEdgeStreamByRow = new Map<number, DecodedEdgeStreamFrame | null>();

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
  const hdr = new DataView(buf, 0, BUF_EDGE_STREAM_FRAME_HEADER_SIZE);
  const tick = hdr.getUint32(0, true);

  let off = BUF_EDGE_STREAM_FRAME_HEADER_SIZE;
  const edgeView = new DataView(buf, off, EDGE_STRIDE);
  off += EDGE_STRIDE;

  const labelLen = readEdgeEdgeLabelLen(edgeView, 0);
  if (buf.byteLength < off + labelLen) return null;
  const labelBytes = new Uint8Array(buf, off, labelLen);
  const label = STR_DECODER.decode(labelBytes);
  off += labelLen;

  const { count: eventCount, view: eventView, textView: eventTextView } = decodeTrailingEvents(buf, off);

  return { tick, edgeView, label, eventCount, eventView, eventTextView };
}

export function edgeLabel(decoded: DecodedEdgeFrame, row: number): string {

  if (row < 0 || row >= decoded.edgeCount) return "";
  const off = readEdgeEdgeLabelOff(decoded.edgeView, row);
  const len = readEdgeEdgeLabelLen(decoded.edgeView, row);
  if (len === 0) return "";
  if (off < 0 || len < 0 || off + len > decoded.edgeLabelBytes.byteLength) return "";
  return STR_DECODER.decode(decoded.edgeLabelBytes.subarray(off, off + len));
}
