import { NODE_KIND_NAMES } from "../../../schema/node-defs";
import type { DecodedNodeFrame } from "./buffer-decode-node";
import {
  readNodeRadius,
  readNodeKindId,
  UNKNOWN_KIND_ID,
} from "../../../../Buffer/buffer-layout";
import type { Line } from "./decode-event-line";
import { nodeCenterX, nodeCenterY, nodeCenterZ } from "../../../../Node/node-frame";

export function nodeGeometryLine(dn: DecodedNodeFrame, nodeRow: number, node: string): Line {

  if (nodeRow < 0 || nodeRow >= dn.nodeCount) return { kind: "node-geometry", node };
  const n = dn.nodeView;
  const cx = nodeCenterX(n, nodeRow), cy = nodeCenterY(n, nodeRow), cz = nodeCenterZ(n, nodeRow);
  const radius = readNodeRadius(n, nodeRow);
  const kindId = readNodeKindId(n, nodeRow);

  const l: Line = { kind: "node-geometry", node };
  if (node) l.label = node;
  if (kindId !== UNKNOWN_KIND_ID && NODE_KIND_NAMES[kindId] !== undefined) l.nodeKind = NODE_KIND_NAMES[kindId];
  l.nx = cx; l.ny = cy; l.nz = cz; l.radius = radius;
  return l;
}
