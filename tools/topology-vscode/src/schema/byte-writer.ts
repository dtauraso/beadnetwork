// byte-writer.ts — the little-endian growable record builder input-encode.ts writes
// TS→Go binary records through, plus the shared enum-name→wire-index helper.

/** ByteWriter — a little-endian growable record builder. */
export class ByteWriter {
  private buf = new Uint8Array(64);
  private view = new DataView(this.buf.buffer);
  private pos = 0;

  private ensure(n: number): void {
    if (this.pos + n <= this.buf.length) return;
    let cap = this.buf.length * 2;
    while (cap < this.pos + n) cap *= 2;
    const next = new Uint8Array(cap);
    next.set(this.buf);
    this.buf = next;
    this.view = new DataView(this.buf.buffer);
  }

  u8(v: number): void {
    this.ensure(1);
    this.view.setUint8(this.pos, v);
    this.pos += 1;
  }
  u16(v: number): void {
    this.ensure(2);
    this.view.setUint16(this.pos, v, true);
    this.pos += 2;
  }
  i32(v: number): void {
    this.ensure(4);
    this.view.setInt32(this.pos, v, true);
    this.pos += 4;
  }
  u32(v: number): void {
    this.ensure(4);
    this.view.setUint32(this.pos, v, true);
    this.pos += 4;
  }
  f32(v: number): void {
    this.ensure(4);
    this.view.setFloat32(this.pos, v, true);
    this.pos += 4;
  }
  f64(v: number): void {
    this.ensure(8);
    this.view.setFloat64(this.pos, v, true);
    this.pos += 8;
  }
  bool(v: boolean): void {
    this.u8(v ? 1 : 0);
  }
  str(s: string): void {
    const bytes = new TextEncoder().encode(s);
    this.u32(bytes.length);
    this.ensure(bytes.length);
    this.buf.set(bytes, this.pos);
    this.pos += bytes.length;
  }
  /** The record bytes (a fresh ArrayBuffer sized to content). */
  toArrayBuffer(): ArrayBuffer {
    return this.buf.buffer.slice(0, this.pos);
  }
}

/** enumIndex — look up an enum string's wire index in a generated ordering array,
 *  defaulting to 0 (the ordering's first member) when not found. */
export function enumIndex(list: readonly string[], s: string): number {
  const i = list.indexOf(s);
  return i < 0 ? 0 : i;
}
