
const latest = new Map<number, DataView>();
let version = 0;
const subs = new Set<() => void>();

export function setColumnValue(col: number, buf: ArrayBuffer): void {
  latest.set(col, new DataView(buf));
  version++;
  for (const fn of subs) fn();
}

export function columnVersion(): number {
  return version;
}

export function subscribeColumns(fn: () => void): () => void {
  subs.add(fn);
  return () => { subs.delete(fn); };
}

export function hasColumn(col: number): boolean {
  return latest.has(col);
}

export function columnF32(col: number, fallback = 0): number {
  const v = latest.get(col);
  return v && v.byteLength >= 4 ? v.getFloat32(0, true) : fallback;
}

export function columnI32(col: number, fallback = 0): number {
  const v = latest.get(col);
  return v && v.byteLength >= 4 ? v.getInt32(0, true) : fallback;
}

export function columnU32(col: number, fallback = 0): number {
  const v = latest.get(col);
  return v && v.byteLength >= 4 ? v.getUint32(0, true) : fallback;
}

export function columnU8(col: number, fallback = 0): number {
  const v = latest.get(col);
  return v && v.byteLength >= 1 ? v.getUint8(0) : fallback;
}

export function columnBytes(col: number): DataView | undefined {
  return latest.get(col);
}
