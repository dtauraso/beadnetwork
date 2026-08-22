package main

const tsByteWriterSource = "" +
	`export class ByteWriter {
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
      ` + "`" + `enumIndex: "${s}" is not in the wire enum [${list.join(",")}] — every caller passes a name ` + "`" + ` +
        ` + "`" + `that must be there, and the list this file was generated with does not carry it, so the ` + "`" + ` +
        ` + "`" + `input layout in Input/gen was not extended when the flag was added. Returning 0 here ` + "`" + ` +
        ` + "`" + `silently encoded the FIRST enum member instead, sending a correct-looking edit addressed ` + "`" + ` +
        ` + "`" + `to the wrong one.` + "`" + `,
    );
  }
  return i;
}
`

const tsMessagesSource = `import type { RawInputEvent } from "../Input/Drag/raw-input";
import type { OverlayEditMsg } from "../Overlay/edits";
import type { PanelEditMsg } from "../Chrome/Panels/Panel/edits";
import type { ClockEditMsg } from "../Clock/edits";
import type { SceneEditMsg } from "../Scene/edits";
import type { NodeEditMsg } from "../Node/edits";
import type { EdgeEditMsg } from "../Node/Edge/edits";
import type { OverlayFlag } from "../Overlay/flags";
import type { PanelFlag } from "../Chrome/Panels/Panel/flags";

export type { OverlayFlag, PanelFlag };

type EditMsg =
  | OverlayEditMsg
  | PanelEditMsg
  | ClockEditMsg
  | SceneEditMsg
  | NodeEditMsg
  | EdgeEditMsg;

export type WebviewToHostMsg =
  | { type: "ready" }
  | { type: "go-record"; record: ArrayBuffer }
  | { type: "raw-input"; event: RawInputEvent }
  | { type: "save" }
  | { type: "webview-log"; entry: string }
  | EditMsg;

export const WEBVIEW_TO_HOST_TYPES: ReadonlySet<WebviewToHostMsg["type"]> = new Set([
  "ready", "webview-log", "edit", "save", "raw-input", "go-record",
]);

export function parseWebviewToHost(raw: unknown): WebviewToHostMsg | undefined {
  if (!raw || typeof raw !== "object") return undefined;
  const t = (raw as { type?: unknown }).type;
  if (typeof t !== "string" || !WEBVIEW_TO_HOST_TYPES.has(t as WebviewToHostMsg["type"])) {
    return undefined;
  }
  const m = raw as Record<string, unknown>;
  switch (t) {
    case "webview-log":
      return typeof m.entry === "string" ? (m as unknown as WebviewToHostMsg) : undefined;
    case "go-record":
      return m.record instanceof ArrayBuffer ? (m as unknown as WebviewToHostMsg) : undefined;
    default:
      return m as unknown as WebviewToHostMsg;
  }
}
`

const tsAttrIndexSource = "" +
	`export function attrIndex(attr: string): number {
  const i = (IN_UPDATE_ATTRS as readonly string[]).indexOf(attr);
  if (i < 0) {
    throw new Error(
      ` + "`" + `attrIndex: no wire byte exists for update attribute "${attr}" — the attr list this file was ` + "`" + ` +
        ` + "`" + `generated with does not carry it, so an edit naming it could never be encoded. Declare it ` + "`" + ` +
        ` + "`" + `in Input/gen's input layout and regenerate, in the same commit as the decoder that ` + "`" + ` +
        ` + "`" + `reads it.` + "`" + `,
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
`
