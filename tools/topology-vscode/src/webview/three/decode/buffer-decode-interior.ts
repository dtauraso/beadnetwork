import { BUF_INTERIOR_STREAM_FRAME_HEADER_SIZE } from "../../../../Buffer/frame-tags";
import { decodeTrailingEvents } from "./buffer-decode-shared";
export { INTERIOR_SLOTS_PER_NODE } from "../../../../Buffer/buffer-layout";

export interface DecodedInteriorStreamFrame {
  tick: number;

  eventCount: number;
  eventView: DataView;
  eventTextView: DataView;
}

const lastInteriorStreamBufByRow = new Map<number, ArrayBuffer>();
const lastDecodedInteriorStreamByRow = new Map<number, DecodedInteriorStreamFrame | null>();

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
  if (buf.byteLength < BUF_INTERIOR_STREAM_FRAME_HEADER_SIZE) return null;
  const hdr = new DataView(buf, 0, BUF_INTERIOR_STREAM_FRAME_HEADER_SIZE);
  const tick = hdr.getUint32(0, true);

  const { count: eventCount, view: eventView, textView: eventTextView } = decodeTrailingEvents(
    buf, BUF_INTERIOR_STREAM_FRAME_HEADER_SIZE);

  return { tick, eventCount, eventView, eventTextView };
}
