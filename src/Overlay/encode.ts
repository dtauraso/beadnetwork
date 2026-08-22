import { editUpdate } from "../Input/Codec/attr-index";
import { enumIndex } from "../Input/Codec/byte-writer";
import { OVERLAY_FLAG_ORDER, type OverlayFlag } from "./flags";

export function encodeOverlaysToggle(flag: OverlayFlag): ArrayBuffer {
  const w = editUpdate("overlays", "toggle");
  w.u8(enumIndex(OVERLAY_FLAG_ORDER, flag));
  return w.toArrayBuffer();
}
