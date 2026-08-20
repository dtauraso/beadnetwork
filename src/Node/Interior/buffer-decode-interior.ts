import { BUF_INTERIOR_STREAM_FRAME_HEADER_SIZE } from "../../schema/buffer-layout/frame-tags";
import { decodeTrailingEvents, type DecodedEvents } from "../../webview/decode/buffer-decode-shared";
export { INTERIOR_SLOTS_PER_NODE } from "../../schema/buffer-layout/buffer-layout";

export interface DecodedInteriorStreamFrame {
  tick: number;

  events: DecodedEvents;
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

  const events = decodeTrailingEvents(
    buf, BUF_INTERIOR_STREAM_FRAME_HEADER_SIZE);

  return { tick, events };
}
