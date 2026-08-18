import { BUF_EDGE_STREAM_FRAME_HEADER_SIZE } from "../../../../Buffer/frame-tags";
import { STR_DECODER, decodeTrailingEvents, type DecodedEvents } from "./buffer-decode-shared";
import { columnBytes } from "../../../../Buffer/column-values";
import { edgeColumn } from "../../../../Buffer/column-owners";
import { COL_STREAM_EDGE_LABEL } from "../../../../Buffer/column-streams-gen";

export interface DecodedEdgeStreamFrame {
  tick: number;

  events: DecodedEvents;
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
  if (buf.byteLength < BUF_EDGE_STREAM_FRAME_HEADER_SIZE) return null;
  const hdr = new DataView(buf, 0, BUF_EDGE_STREAM_FRAME_HEADER_SIZE);
  const tick = hdr.getUint32(0, true);

  const events = decodeTrailingEvents(
    buf, BUF_EDGE_STREAM_FRAME_HEADER_SIZE);

  return { tick, events };
}

export function edgeLabel(row: number): string {
  if (row < 0) return "";
  const bytes = columnBytes(edgeColumn(row, COL_STREAM_EDGE_LABEL));
  if (!bytes || bytes.byteLength === 0) return "";
  return STR_DECODER.decode(new Uint8Array(bytes.buffer, bytes.byteOffset, bytes.byteLength));
}
