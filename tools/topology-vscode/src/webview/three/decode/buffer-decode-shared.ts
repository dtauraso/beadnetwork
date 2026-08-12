import { EVENT_STRIDE } from "../../../schema/buffer-layout";

export const STR_DECODER = new TextDecoder();

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
