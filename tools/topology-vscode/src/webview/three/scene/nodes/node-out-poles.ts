import { getLatestNodeStreamFrames, getNodeStreamVersion } from "../../../snapshot-buffer";
import { decodeNodeStreamFrame } from "../../decode/buffer-decode-node";
import {
  readNodeCX, readNodeCY, readNodeCZ, readNodeRadius,
  readOutPoleDX, readOutPoleDY, readOutPoleDZ,
} from "../../../../schema/buffer-layout/buffer-layout";

// One frame per OUTGOING neighbour of one node: the node's centre and the pole
// direction Go streams for that edge. Go owns the direction — this only groups
// the rows it sent by node.
export interface NodeOutPoles {
  row: number;
  cx: number; cy: number; cz: number;
  radius: number;
  poles: { x: number; y: number; z: number }[];
}

let lastVersion = -1;
let lastAgg: NodeOutPoles[] = [];

export function getNodeOutPoles(): NodeOutPoles[] {
  const nv = getNodeStreamVersion();
  if (nv === lastVersion) return lastAgg;

  const out: NodeOutPoles[] = [];
  for (const [row, buf] of getLatestNodeStreamFrames()) {
    const decoded = decodeNodeStreamFrame(row, buf);
    if (!decoded || decoded.outPoleCount === 0) continue;
    const poles: { x: number; y: number; z: number }[] = [];
    for (let i = 0; i < decoded.outPoleCount; i++) {
      poles.push({
        x: readOutPoleDX(decoded.outPoleView, i),
        y: readOutPoleDY(decoded.outPoleView, i),
        z: readOutPoleDZ(decoded.outPoleView, i),
      });
    }
    out.push({
      row,
      cx: readNodeCX(decoded.nodeView, 0),
      cy: readNodeCY(decoded.nodeView, 0),
      cz: readNodeCZ(decoded.nodeView, 0),
      radius: readNodeRadius(decoded.nodeView, 0),
      poles,
    });
  }

  lastVersion = nv;
  lastAgg = out;
  return out;
}
