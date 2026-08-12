// decode-event-node-geometry.ts — the node-geometry EVENT-row decode, split out of
// decode-event-line.ts. Pure per-row decode: reads one node's geometry columns off the
// cached per-node stream frame into the JSON-shaped Line the .probe logger writes.

import { NODE_KIND_NAMES } from "../../../schema/node-defs";
import type { DecodedNodeFrame } from "./buffer-decode-node";
import {
  readNodeCX, readNodeCY, readNodeCZ, readNodeRadius, readNodeSphereR,
  readNodeVRX, readNodeVRY, readNodeVRZ, readNodeFRX, readNodeFRY, readNodeFRZ,
  readNodeKindId,
  UNKNOWN_KIND_ID,
} from "../../../schema/buffer-layout";
import type { Line } from "./decode-event-line";

export function nodeGeometryLine(dn: DecodedNodeFrame, nodeRow: number, node: string): Line {
  // A node-geometry event riding the VIEW bucket resolves its node columns against the
  // last cached per-node stream frame, which can be a STALE generation with fewer rows than the
  // topology — reading nodeRow past nodeView would throw. Degrade to the label-only line
  // (same graceful-empty contract as nodeLabel), never crash the .probe logger.
  if (nodeRow < 0 || nodeRow >= dn.nodeCount) return { kind: "node-geometry", node };
  const n = dn.nodeView;
  const cx = readNodeCX(n, nodeRow), cy = readNodeCY(n, nodeRow), cz = readNodeCZ(n, nodeRow);
  const radius = readNodeRadius(n, nodeRow);
  const sphereR = readNodeSphereR(n, nodeRow);
  const kindId = readNodeKindId(n, nodeRow);
  // No `ports` array any more (docs/bead-model/channels-not-ports.md): a port carries no geometry,
  // so there is nothing to report per node beyond its own fields below.
  const l: Line = { kind: "node-geometry", node };
  if (node) l.label = node;
  if (kindId !== UNKNOWN_KIND_ID && NODE_KIND_NAMES[kindId] !== undefined) l.nodeKind = NODE_KIND_NAMES[kindId];
  l.nx = cx; l.ny = cy; l.nz = cz; l.radius = radius;
  if (sphereR !== 0) l.sphereR = sphereR;
  l.vrx = readNodeVRX(n, nodeRow); l.vry = readNodeVRY(n, nodeRow); l.vrz = readNodeVRZ(n, nodeRow);
  l.frx = readNodeFRX(n, nodeRow); l.fry = readNodeFRY(n, nodeRow); l.frz = readNodeFRZ(n, nodeRow);
  return l;
}
