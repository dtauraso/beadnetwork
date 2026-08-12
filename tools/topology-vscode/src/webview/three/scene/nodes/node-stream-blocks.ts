import { getLatestNodeStreamFrames, getNodeStreamVersion } from "../../../snapshot-buffer";
import {
  decodeNodeStreamFrame,
  type DecodedNodeStreamFrame,
} from "../../decode/buffer-decode-node";
import {
  readNodeRingAxisTheta,
  readNodeRingAxisPhi,
  readNodeCX, readNodeCY, readNodeCZ,
  readChainBeadOX, readChainBeadOY, readChainBeadOZ, readChainBeadLit, readChainBeadLitValue,
} from "../../../../schema/buffer-layout";
import { poleAxis } from "../buffer-scene-shared";

export interface ChainBeadsAgg {

  ringAxis: Float32Array;

  positions: Float32Array;

  count: number;

  lit: Uint8Array;

  litValue: Int32Array;
}

let lastChainVersion = -1;
let lastChainAgg: ChainBeadsAgg | null = null;

export function getChainBeads(): ChainBeadsAgg {
  const nodeFrames = getLatestNodeStreamFrames();
  const nv = getNodeStreamVersion();
  if (lastChainAgg !== null && nv === lastChainVersion) {
    return lastChainAgg;
  }

  const decodedByRow: DecodedNodeStreamFrame[] = [];
  let total = 0;
  for (const [row, buf] of nodeFrames) {
    const decoded = decodeNodeStreamFrame(row, buf);
    if (!decoded) continue;
    decodedByRow.push(decoded);
    total += decoded.chainBeadCount;
  }
  const positions = new Float32Array(total * 3);

  const ringAxis = new Float32Array(total * 3);
  const lit = new Uint8Array(total);
  const litValue = new Int32Array(total);
  let w = 0;
  let b = 0;
  for (const decoded of decodedByRow) {

    const cx = readNodeCX(decoded.nodeView, 0);
    const cy = readNodeCY(decoded.nodeView, 0);
    const cz = readNodeCZ(decoded.nodeView, 0);

    const poleTheta = readNodeRingAxisTheta(decoded.nodeView, 0);
    const polePhi = readNodeRingAxisPhi(decoded.nodeView, 0);
    const [ax, ay, az] = poleAxis(poleTheta, polePhi);
    for (let i = 0; i < decoded.chainBeadCount; i++) {
      ringAxis[w] = ax;
      ringAxis[w + 1] = ay;
      ringAxis[w + 2] = az;
      positions[w++] = cx + readChainBeadOX(decoded.chainBeadView, i);
      positions[w++] = cy + readChainBeadOY(decoded.chainBeadView, i);
      positions[w++] = cz + readChainBeadOZ(decoded.chainBeadView, i);
      litValue[b] = readChainBeadLitValue(decoded.chainBeadView, i);
      lit[b++] = readChainBeadLit(decoded.chainBeadView, i);
    }
  }
  lastChainVersion = nv;
  lastChainAgg = { positions, ringAxis, count: total, lit, litValue };
  return lastChainAgg;
}
