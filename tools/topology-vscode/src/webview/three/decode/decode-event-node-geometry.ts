import { NODE_KIND_NAMES } from "../../../schema/node-defs";
import { UNKNOWN_KIND_ID } from "../../../../Buffer/buffer-layout";
import { columnF32, columnI32, columnU8 } from "../../../../Buffer/column-values";
import { nodeColumn } from "../../../../Buffer/column-owners";
import {
  COL_STREAM_NODE_INDEX_R, COL_STREAM_NODE_INDEX_PHI, COL_STREAM_NODE_INDEX_THETA,
  COL_STREAM_NODE_RADIUS, COL_STREAM_NODE_KIND_ID,
} from "../../../../Buffer/column-streams-gen";
import type { Line } from "./decode-event-line";

export function nodeGeometryLine(row: number, node: string): Line {
  const l: Line = { kind: "node-geometry", node };
  if (row < 0) return l;

  const kindId = columnU8(nodeColumn(row, COL_STREAM_NODE_KIND_ID), UNKNOWN_KIND_ID);
  if (kindId !== UNKNOWN_KIND_ID && NODE_KIND_NAMES[kindId] !== undefined) {
    l.nodeKind = NODE_KIND_NAMES[kindId];
  }
  if (node) l.label = node;
  l.indexR = columnI32(nodeColumn(row, COL_STREAM_NODE_INDEX_R));
  l.indexPhi = columnI32(nodeColumn(row, COL_STREAM_NODE_INDEX_PHI));
  l.indexTheta = columnI32(nodeColumn(row, COL_STREAM_NODE_INDEX_THETA));
  l.radius = columnF32(nodeColumn(row, COL_STREAM_NODE_RADIUS));
  return l;
}
