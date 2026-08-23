import { editUpdate } from "./record-writer";

export function encodeSceneSelected(tabIndex: number): ArrayBuffer {
  const w = editUpdate("scene", "selected");
  w.u8(tabIndex);
  return w.toArrayBuffer();
}

export function encodeSceneLatticePoints(points: number): ArrayBuffer {
  const w = editUpdate("scene", "latticePoints");
  w.u8(points);
  return w.toArrayBuffer();
}

export function encodeSceneCreate(kindId: number, ndcX: number, ndcY: number): ArrayBuffer {
  const w = editUpdate("scene", "create");
  w.u8(kindId);
  w.f32(ndcX);
  w.f32(ndcY);
  return w.toArrayBuffer();
}

export function encodeSceneDelete(nodeRow: number): ArrayBuffer {
  const w = editUpdate("scene", "delete");
  w.u8(nodeRow);
  return w.toArrayBuffer();
}

export function encodeSceneViewport(width: number, height: number): ArrayBuffer {
  const w = editUpdate("scene", "viewport");
  w.f32(width);
  w.f32(height);
  return w.toArrayBuffer();
}

export function encodeTiltVectorAdjust(nodeRow: number, dir: "up" | "down"): ArrayBuffer {
  const w = editUpdate("tiltVector", "phi");
  w.u8(nodeRow);
  w.u8(dir === "up" ? 1 : 0);
  return w.toArrayBuffer();
}

export function encodeTiltVectorReset(nodeRow: number): ArrayBuffer {
  const w = editUpdate("tiltVector", "reset");
  w.u8(nodeRow);
  return w.toArrayBuffer();
}

export function encodeTiltVectorStart(nodeRow: number): ArrayBuffer {
  const w = editUpdate("tiltVector", "start");
  w.u8(nodeRow);
  return w.toArrayBuffer();
}
