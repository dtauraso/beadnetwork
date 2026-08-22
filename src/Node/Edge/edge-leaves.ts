import { makeRowLeafValues } from "../../valuefile/row-leaf-values";
import { EDGE_VALUE_NAMES, type EdgeValueName } from "./edge-values-gen";

const values = makeRowLeafValues<EdgeValueName>(
  "Node/Edge/paths",
  EDGE_VALUE_NAMES,
);

export const edgeBytes = values.bytes;

export function edgeF32(row: number, name: EdgeValueName, fallback = 0): number {
  const v = edgeBytes(row, name);
  return v && v.byteLength >= 4 ? v.getFloat32(0, true) : fallback;
}

export function edgeI32(row: number, name: EdgeValueName, fallback = 0): number {
  const v = edgeBytes(row, name);
  return v && v.byteLength >= 4 ? v.getInt32(0, true) : fallback;
}

export function edgeU8(row: number, name: EdgeValueName, fallback = 0): number {
  const v = edgeBytes(row, name);
  return v && v.byteLength >= 1 ? v.getUint8(0) : fallback;
}
