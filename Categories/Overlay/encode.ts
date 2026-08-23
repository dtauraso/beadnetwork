import { editUpdate, enumIndex } from "./record-writer";
import { OVERLAY_FLAG_ORDER, type OverlayFlag } from "./flags";

export function encodeOverlaysToggle(flag: OverlayFlag): ArrayBuffer {
  const w = editUpdate("overlays", "toggle");
  w.u8(enumIndex(OVERLAY_FLAG_ORDER, flag));
  return w.toArrayBuffer();
}
