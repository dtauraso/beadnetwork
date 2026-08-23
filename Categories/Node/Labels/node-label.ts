import { nodeBytes } from "../node-leaves";

const STR_DECODER = new TextDecoder();

export function nodeLabel(row: number): string {
  if (row < 0) return "";
  const bytes = nodeBytes(row, "label");
  if (!bytes || bytes.byteLength === 0) return "";
  return STR_DECODER.decode(new Uint8Array(bytes.buffer, bytes.byteOffset, bytes.byteLength));
}
