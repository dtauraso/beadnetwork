import { readNodeKindId, UNKNOWN_KIND_ID } from "../Buffer/buffer-layout";
import { NODE_DEFS_ARRAY, NODE_KIND_NAMES } from "../src/schema/node-defs";

// The one reader of the KindId column. Two consumers wanted the kind for two different
// reasons -- colours for the node body, the name for the rules panel -- and each read the
// column itself, which is a second reader of a value that is fixed at load and identical
// both times. They ask here instead.

const NODE_DEFAULT_FILL = "#ffffff";
const NODE_DEFAULT_STROKE = "#888888";

export function nodeRowColors(nodeView: DataView, row: number): { fill: string; stroke: string } {
  const kindId = readNodeKindId(nodeView, row);
  const def = kindId === UNKNOWN_KIND_ID ? undefined : NODE_DEFS_ARRAY[kindId];
  return {
    fill: def?.fill ?? NODE_DEFAULT_FILL,
    stroke: def?.stroke ?? NODE_DEFAULT_STROKE,
  };
}

export function nodeKindName(nodeView: DataView, row: number): string {
  const kindId = readNodeKindId(nodeView, row);
  if (kindId === UNKNOWN_KIND_ID) return "";
  return NODE_KIND_NAMES[kindId] ?? "";
}
