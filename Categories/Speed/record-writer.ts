import { IN_UPDATE_ATTRS } from "./update-attrs";
import { IN_KIND_EDIT_UPDATE, IN_UPDATE_KINDS } from "../Scene/Drag/input-defs";

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

  toArrayBuffer(): ArrayBuffer {
    return this.buf.buffer.slice(0, this.pos);
  }
}

export function enumIndex(list: readonly string[], s: string): number {
  const i = list.indexOf(s);
  if (i < 0) {
    throw new Error(
      `enumIndex: "${s}" is not in the wire enum [${list.join(",")}] — every caller passes a name ` +
        `that must be there, so this list is missing it. Returning 0 here would silently encode the ` +
        `FIRST enum member instead, sending a correct-looking edit addressed to the wrong one.`,
    );
  }
  return i;
}

export function attrIndex(attr: string): number {
  const i = (IN_UPDATE_ATTRS as readonly string[]).indexOf(attr);
  if (i < 0) {
    throw new Error(
      `attrIndex: no wire byte exists for update attribute "${attr}" — Speed/update_attrs.go does ` +
        `not carry it, so an edit naming it could never be encoded. Add it there and regenerate, in ` +
        `the same commit as the decoder that reads it.`,
    );
  }
  return i;
}

export function editUpdate(entity: string, attr: string): ByteWriter {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, entity));
  w.u8(attrIndex(attr));
  return w;
}
