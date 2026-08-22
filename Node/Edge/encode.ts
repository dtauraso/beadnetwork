import { editUpdate } from "./wire-gen";

export function encodeEdgeDragActiveToggle(edgeRow: number): ArrayBuffer {
  const w = editUpdate("edge", "dragActive");
  w.u8(edgeRow);
  return w.toArrayBuffer();
}
