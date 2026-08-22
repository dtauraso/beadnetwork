import { editUpdate } from "../../Input/Codec/attr-index";

export function encodeEdgeDragActiveToggle(edgeRow: number): ArrayBuffer {
  const w = editUpdate("edge", "dragActive");
  w.u8(edgeRow);
  return w.toArrayBuffer();
}
