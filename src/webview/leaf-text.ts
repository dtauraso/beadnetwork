const TEXT = new TextDecoder();

export function decodeAt(bytes: Uint8Array, off: number, len: number): string {
  return TEXT.decode(bytes.subarray(off, off + len));
}
