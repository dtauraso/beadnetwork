import { INTERIOR_STRIDE, INTERIOR_SLOTS_PER_NODE } from "../../../../Buffer/buffer-layout";
import { BUF_INTERIOR_STREAM_FRAME_HEADER_SIZE } from "../../../../Buffer/frame-tags";
import { decodeTrailingEvents } from "./buffer-decode-shared";

export { INTERIOR_SLOTS_PER_NODE } from "../../../../Buffer/buffer-layout";

export interface DecodedInteriorStreamFrame {
  tick: number;

  interiorView: DataView;

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
  const interiorBytes = INTERIOR_SLOTS_PER_NODE * INTERIOR_STRIDE;
  const expectedLen = BUF_INTERIOR_STREAM_FRAME_HEADER_SIZE + interiorBytes;
  if (buf.byteLength < expectedLen) return null;
  const tick = new DataView(buf, 0, BUF_INTERIOR_STREAM_FRAME_HEADER_SIZE).getUint32(0, true);
  const interiorView = new DataView(buf, BUF_INTERIOR_STREAM_FRAME_HEADER_SIZE, interiorBytes);
  const { count: eventCount, view: eventView, textView: eventTextView } = decodeTrailingEvents(buf, expectedLen);
  return { tick, interiorView, eventCount, eventView, eventTextView };
}
