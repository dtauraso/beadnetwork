import { ByteWriter, enumIndex } from "./byte-writer";
import {
  IN_KIND_RAW_INPUT,
  IN_KIND_EDIT_UPDATE,
  IN_EVENT_KINDS,
  IN_HIT_KINDS,
  IN_UPDATE_KINDS,
} from "./input-layout-gen";
import {
  IN_TILT_VECTOR_ATTR_PHI,
  IN_TILT_VECTOR_ATTR_RESET,
  IN_TILT_VECTOR_ATTR_START,
  IN_SCENE_ATTR_LATTICE_POINTS,
  IN_SCENE_ATTR_CREATE,
  IN_SCENE_ATTR_DELETE,
} from "./input-attrs";
import type { RawInputEvent } from "../../messages";

export function encodeSceneCreate(kindId: number, ndcX: number, ndcY: number): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, "scene"));
  w.u8(IN_SCENE_ATTR_CREATE);
  w.u8(kindId);
  w.f32(ndcX);
  w.f32(ndcY);
  return w.toArrayBuffer();
}

export function encodeSceneDelete(nodeRow: number): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, "scene"));
  w.u8(IN_SCENE_ATTR_DELETE);
  w.u8(nodeRow);
  return w.toArrayBuffer();
}

export function encodeSceneLatticePoints(points: number): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, "scene"));
  w.u8(IN_SCENE_ATTR_LATTICE_POINTS);
  w.u8(points);
  return w.toArrayBuffer();
}

export function encodeTiltVectorAdjust(nodeRow: number, dir: "up" | "down"): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, "tiltVector"));
  w.u8(IN_TILT_VECTOR_ATTR_PHI);
  w.u8(nodeRow);
  w.u8(dir === "up" ? 1 : 0);
  return w.toArrayBuffer();
}

export function encodeTiltVectorReset(nodeRow: number): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, "tiltVector"));
  w.u8(IN_TILT_VECTOR_ATTR_RESET);
  w.u8(nodeRow);
  return w.toArrayBuffer();
}

export function encodeTiltVectorStart(nodeRow: number): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_EDIT_UPDATE);
  w.u8(enumIndex(IN_UPDATE_KINDS, "tiltVector"));
  w.u8(IN_TILT_VECTOR_ATTR_START);
  w.u8(nodeRow);
  return w.toArrayBuffer();
}

export function encodeRawInput(ev: RawInputEvent): ArrayBuffer {
  const w = new ByteWriter();
  w.u8(IN_KIND_RAW_INPUT);
  w.u8(enumIndex(IN_EVENT_KINDS, ev.kind));
  w.f64(ev.x);
  w.f64(ev.y);
  w.f64(ev.rectLeft);
  w.f64(ev.rectTop);
  w.f64(ev.rectWidth);
  w.f64(ev.rectHeight);
  w.i32(ev.button);
  w.bool(ev.ctrl);
  w.bool(ev.shift);
  w.bool(ev.alt);
  w.bool(ev.meta);
  w.f64(ev.deltaX);
  w.f64(ev.deltaY);
  w.f64(ev.fov);
  w.u8(enumIndex(IN_HIT_KINDS, ev.hit.kind));
  w.bool(ev.hit.isInput);
  w.i32(ev.hit.nodeRow);
  w.i32(ev.hit.portRow);
  w.i32(ev.hit.edgeRow);
  return w.toArrayBuffer();
}

export function frameRecord(record: ArrayBuffer): Uint8Array {
  const rec = new Uint8Array(record);
  const out = new Uint8Array(4 + rec.length);
  new DataView(out.buffer).setUint32(0, rec.length, true);
  out.set(rec, 4);
  return out;
}
