import { columnU8 } from "../Buffer/column-values";
import { nodeColumn } from "../Buffer/column-owners";
import { COL_STREAM_NODE_KIND_ID } from "../Buffer/column-streams-gen";
import { NODE_DEFS_ARRAY, NODE_KIND_NAMES } from "../schema/node-defs";
import { UNKNOWN_KIND_ID } from "../Buffer/buffer-layout";

const NODE_DEFAULT_FILL = "#ffffff";
const NODE_DEFAULT_STROKE = "#888888";

export function nodeRowColors(row: number): { fill: string; stroke: string } {
  const kindId = columnU8(nodeColumn(row, COL_STREAM_NODE_KIND_ID), UNKNOWN_KIND_ID);
  const def = kindId === UNKNOWN_KIND_ID ? undefined : NODE_DEFS_ARRAY[kindId];
  return {
    fill: def?.fill ?? NODE_DEFAULT_FILL,
    stroke: def?.stroke ?? NODE_DEFAULT_STROKE,
  };
}

export function nodeKindName(row: number): string {
  const kindId = columnU8(nodeColumn(row, COL_STREAM_NODE_KIND_ID), UNKNOWN_KIND_ID);
  if (kindId === UNKNOWN_KIND_ID) return "";
  return NODE_KIND_NAMES[kindId] ?? "";
}
