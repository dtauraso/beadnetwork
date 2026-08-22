import { nodeU8 } from "./node-leaves";
import { NODE_DEFS_ARRAY, NODE_KIND_NAMES } from "../NodeKinds/node-defs";
import { UNKNOWN_KIND_ID } from "../NodeKinds/node-defs";

const NODE_DEFAULT_FILL = "#ffffff";
const NODE_DEFAULT_STROKE = "#888888";

export function nodeRowColors(row: number): { fill: string; stroke: string } {
  const kindId = nodeU8(row, "kindId", UNKNOWN_KIND_ID);
  const def = kindId === UNKNOWN_KIND_ID ? undefined : NODE_DEFS_ARRAY[kindId];
  return {
    fill: def?.fill ?? NODE_DEFAULT_FILL,
    stroke: def?.stroke ?? NODE_DEFAULT_STROKE,
  };
}

export function nodeKindName(row: number): string {
  const kindId = nodeU8(row, "kindId", UNKNOWN_KIND_ID);
  if (kindId === UNKNOWN_KIND_ID) return "";
  return NODE_KIND_NAMES[kindId] ?? "";
}
