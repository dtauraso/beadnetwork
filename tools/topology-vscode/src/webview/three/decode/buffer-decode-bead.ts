import { BUF_BEAD_STREAM_FRAME_HEADER_SIZE } from "../../../../Buffer/frame-tags";
import { decodeTrailingEvents, type DecodedEvents } from "./buffer-decode-shared";

export interface DecodedBeadStreamFrame {
  tick: number;

  nodeRow: number;

  events: DecodedEvents;
}

const lastBeadStreamBufByRow = new Map<number, ArrayBuffer>();
const lastDecodedBeadStreamByRow = new Map<number, DecodedBeadStreamFrame | null>();

export function decodeBeadStreamFrame(row: number, buf: ArrayBuffer): DecodedBeadStreamFrame | null {
  if (lastBeadStreamBufByRow.get(row) === buf) {
    return lastDecodedBeadStreamByRow.get(row) ?? null;
  }
  const decoded = decodeBeadStreamFrameUncached(buf);
  lastBeadStreamBufByRow.set(row, buf);
  lastDecodedBeadStreamByRow.set(row, decoded);
  return decoded;
}

function decodeBeadStreamFrameUncached(buf: ArrayBuffer): DecodedBeadStreamFrame | null {
  if (buf.byteLength < BUF_BEAD_STREAM_FRAME_HEADER_SIZE) return null;
  const hdr = new DataView(buf, 0, BUF_BEAD_STREAM_FRAME_HEADER_SIZE);
  const tick = hdr.getUint32(0, true);
  const nodeRow = hdr.getInt32(4, true);

  const events = decodeTrailingEvents(buf, BUF_BEAD_STREAM_FRAME_HEADER_SIZE);

  return { tick, nodeRow, events };
}
