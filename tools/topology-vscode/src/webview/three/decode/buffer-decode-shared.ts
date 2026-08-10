// buffer-decode-shared.ts — decoding primitives shared by every per-owner stream-frame
// decoder (buffer-decode-view.ts / buffer-decode-edge.ts / buffer-decode-node.ts /
// buffer-decode-interior.ts): the trailing EVENTS section every one of those frames ends
// with, and the UTF-8 decoder each uses to decode its own label/name bytes.

import { EVENT_STRIDE } from "../../../schema/buffer-layout";

/** Shared UTF-8 decoder for the label / port-name / edge-label sections. */
export const STR_DECODER = new TextDecoder();

/** Decodes a trailing EVENTS section ([count:u32] + count × EVENT_STRIDE rows) appended
 *  after `offset` bytes of already-known content in ANY per-owner frame (NODE/EDGE/
 *  INTERIOR/VIEW — memory/feedback_no_single_writer_bridge.md). The view/scene frame no
 *  longer carries an EVENT block at all — each per-owner stream carries its own instead.
 *  Returns {count:0, view: empty} when the buffer is too short to hold even the count
 *  (never null — callers can always safely iterate 0 times). */
// decodeTrailingEvents decodes [count:u32] + count × EVENT_STRIDE event rows, plus the
// single sanctioned free-form event-text-bytes section that follows immediately after
// (see Buffer.BuildEventsSection — TextOff/TextLen on each event row address into
// textView). textView spans whatever bytes remain to the end of the frame; a frame with
// no breadcrumb events has an empty (but valid) textView.
export function decodeTrailingEvents(buf: ArrayBuffer, offset: number): { count: number; view: DataView; textView: DataView } {
  const empty = { count: 0, view: new DataView(buf, buf.byteLength, 0), textView: new DataView(buf, buf.byteLength, 0) };
  if (buf.byteLength < offset + 4) return empty;
  const count = new DataView(buf, offset, 4).getUint32(0, true);
  const bytes = count * EVENT_STRIDE;
  if (buf.byteLength < offset + 4 + bytes) return empty;
  const textStart = offset + 4 + bytes;
  return {
    count,
    view: new DataView(buf, offset + 4, bytes),
    textView: new DataView(buf, textStart, buf.byteLength - textStart),
  };
}
