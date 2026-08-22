import { makeRowLeafValues } from "../valuefile/row-leaf-values";
import { NODE_VALUE_NAMES, type NodeValueName } from "./node-values-gen";

const values = makeRowLeafValues<NodeValueName>(
  "Node/paths",
  NODE_VALUE_NAMES,
);

export const nodeBytes = values.bytes;

export function nodeHas(row: number, name: NodeValueName): boolean {
  return nodeBytes(row, name) !== undefined;
}

export function nodeF32(row: number, name: NodeValueName, fallback = 0): number {
  const v = nodeBytes(row, name);
  return v && v.byteLength >= 4 ? v.getFloat32(0, true) : fallback;
}

export function nodeI32(row: number, name: NodeValueName, fallback = 0): number {
  const v = nodeBytes(row, name);
  return v && v.byteLength >= 4 ? v.getInt32(0, true) : fallback;
}

export function nodeU8(row: number, name: NodeValueName, fallback = 0): number {
  const v = nodeBytes(row, name);
  return v && v.byteLength >= 1 ? v.getUint8(0) : fallback;
}

export const NODE_RING_NAMES = NODE_VALUE_NAMES.slice(15, 31) as readonly NodeValueName[];
