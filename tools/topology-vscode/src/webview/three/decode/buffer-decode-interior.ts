// buffer-decode-interior.ts — decodeInteriorStreamFrame: one node's dedicated per-fd
// INTERIOR_STREAM frame (that node's own interior-bead grid).

import { INTERIOR_STRIDE, INTERIOR_SLOTS_PER_NODE } from "../../../schema/buffer-layout";
import { BUF_INTERIOR_STREAM_FRAME_HEADER_SIZE } from "../../../schema/frame-tags";
import { decodeTrailingEvents } from "./buffer-decode-shared";

// Generated (part of BUF_LAYOUT_FINGERPRINT) — re-exported here so existing consumers
// (buffer-scene.tsx, InteriorBeadInstances.tsx, buffer-log.ts) keep importing it from a
// decode module rather than reaching into schema/buffer-layout directly.
export { INTERIOR_SLOTS_PER_NODE } from "../../../schema/buffer-layout";

/** Decoded view over ONE node's dedicated per-fd INTERIOR-stream frame
 *  (BUF_BLOCK_TAG_INTERIOR_STREAM): [tick:u32] followed by a FIXED
 *  INTERIOR_SLOTS_PER_NODE × INTERIOR_STRIDE bytes (that node's own interior-bead grid). */
export interface DecodedInteriorStreamFrame {
  tick: number;
  /** DataView over this node's own INTERIOR_SLOTS_PER_NODE interior rows. */
  interiorView: DataView;
  /** This goroutine's own trailing EVENTS section (.probe log only; see decodeTrailingEvents). */
  eventCount: number;
  eventView: DataView;
  eventTextView: DataView;
}

const lastInteriorStreamBufByRow = new Map<number, ArrayBuffer>();
const lastDecodedInteriorStreamByRow = new Map<number, DecodedInteriorStreamFrame | null>();

/**
 * Decode ONE node row's BUF_BLOCK_TAG_INTERIOR_STREAM frame ArrayBuffer into a typed view.
 * Returns null if the buffer is too small to be a valid frame. Pure, per-row memoized.
 */
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
