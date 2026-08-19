import { columnBytes } from "../Buffer/column-values";

const TEXT = new TextDecoder();

export function readF32Run(col: number): Float32Array | null {
  const v = columnBytes(col);
  if (!v || v.byteLength === 0) return null;
  const out = new Float32Array(v.byteLength / 4);
  for (let i = 0; i < out.length; i++) out[i] = v.getFloat32(i * 4, true);
  return out;
}

export function readU32Run(col: number): Uint32Array | null {
  const v = columnBytes(col);
  if (!v || v.byteLength === 0) return null;
  const out = new Uint32Array(v.byteLength / 4);
  for (let i = 0; i < out.length; i++) out[i] = v.getUint32(i * 4, true);
  return out;
}

export function readI32Run(col: number): Int32Array | null {
  const v = columnBytes(col);
  if (!v || v.byteLength === 0) return null;
  const out = new Int32Array(v.byteLength / 4);
  for (let i = 0; i < out.length; i++) out[i] = v.getInt32(i * 4, true);
  return out;
}

export function readText(col: number): Uint8Array | null {
  const v = columnBytes(col);
  if (!v) return null;
  return new Uint8Array(v.buffer, v.byteOffset, v.byteLength);
}

export function decodeAt(bytes: Uint8Array, off: number, len: number): string {
  return TEXT.decode(bytes.subarray(off, off + len));
}
