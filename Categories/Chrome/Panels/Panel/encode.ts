import { editUpdate, enumIndex } from "./record-writer";
import { PANEL_FLAG_ORDER, type PanelFlag } from "./flags";

export function encodePanelsToggle(flag: PanelFlag): ArrayBuffer {
  const w = editUpdate("panels", "toggle");
  w.u8(enumIndex(PANEL_FLAG_ORDER, flag));
  return w.toArrayBuffer();
}
