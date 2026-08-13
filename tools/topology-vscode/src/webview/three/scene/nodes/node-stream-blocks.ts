import { getLatestNodeStreamFrames, getNodeStreamVersion } from "../../../snapshot-buffer";
import {
  decodeNodeStreamFrame,
  type DecodedNodeStreamFrame,
} from "../../decode/buffer-decode-node";
import {
  readNodeRingAxisPhi,
  readNodeRingAxisTheta,
  readNodeCX, readNodeCY, readNodeCZ,
  readChainBeadOX, readChainBeadOY, readChainBeadOZ, readChainBeadLit, readChainBeadLitValue,
  readNodeKindId,
} from "../../../../schema/buffer-layout/buffer-layout";
import { NODE_KIND_NAMES } from "../../../../schema/node-defs";
import { poleAxis } from "../buffer-scene-shared";

export interface ChainBeadsAgg {

  ringAxis: Float32Array;

  positions: Float32Array;

  count: number;

  lit: Uint8Array;

  litValue: Int32Array;

  // comm marks each bead whose owning node holds a constraint it has to
  // tell its outgoing neighbours about — today, an input node. Flattening
  // the beads loses which node they came from, and the comm-edges overlay
  // is the one consumer that needs it back, so it is carried per bead
  // rather than recovered by re-walking the frames.
  comm: Uint8Array;
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
  const comm = new Uint8Array(total);
  let w = 0;
  let b = 0;
  for (const decoded of decodedByRow) {

    const cx = readNodeCX(decoded.nodeView, 0);
    const cy = readNodeCY(decoded.nodeView, 0);
    const cz = readNodeCZ(decoded.nodeView, 0);

    const poleTheta = readNodeRingAxisPhi(decoded.nodeView, 0);
    const polePhi = readNodeRingAxisTheta(decoded.nodeView, 0);
    const [ax, ay, az] = poleAxis(poleTheta, polePhi);
    const isComm = NODE_KIND_NAMES[readNodeKindId(decoded.nodeView, 0)] === "Input" ? 1 : 0;
    for (let i = 0; i < decoded.chainBeadCount; i++) {
      comm[b] = isComm;
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
  lastChainAgg = { positions, ringAxis, count: total, lit, litValue, comm };
  return lastChainAgg;
}
