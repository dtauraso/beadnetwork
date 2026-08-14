import {
  getLatestEdgeStreamFrames, getEdgeStreamVersion,
  getLatestNodeStreamFrames, getNodeStreamVersion,
} from "../../../snapshot-buffer";
import { decodeEdgeStreamFrame } from "../../decode/buffer-decode-edge";
import { decodeNodeStreamFrame } from "../../decode/buffer-decode-node";
import {
  readEdgeBeadX, readEdgeBeadY, readEdgeBeadZ, readEdgeBeadValue,
  readEdgeSrcNodeRow,
  readNodeRingAxisPhi, readNodeRingAxisTheta,
} from "../../../../schema/buffer-layout/buffer-layout";
import { poleAxis } from "../buffer-scene-shared";

export interface EdgeBeadsAgg {
  positions: Float32Array;
  value: Int32Array;

  srcNodeRow: Int32Array;

  ringAxis: Float32Array;

  count: number;
}

const RING_AXIS_FALLBACK: [number, number, number] = [0, 1, 0];

function nodeRingAxes(): Map<number, [number, number, number]> {
  const axes = new Map<number, [number, number, number]>();
  for (const [row, buf] of getLatestNodeStreamFrames()) {
    const d = decodeNodeStreamFrame(row, buf);
    if (!d) continue;
    axes.set(row, poleAxis(
      readNodeRingAxisPhi(d.nodeView, 0),
      readNodeRingAxisTheta(d.nodeView, 0),
    ));
  }
  return axes;
}

let lastVersion = -1;
let lastNodeVersion = -1;
let lastAgg: EdgeBeadsAgg | null = null;

export function getEdgeBeads(): EdgeBeadsAgg {
  const ev = getEdgeStreamVersion();
  const nv = getNodeStreamVersion();
  if (lastAgg !== null && ev === lastVersion && nv === lastNodeVersion) return lastAgg;

  const decoded = [];
  let total = 0;
  for (const [row, buf] of getLatestEdgeStreamFrames()) {
    const d = decodeEdgeStreamFrame(row, buf);
    if (!d) continue;
    decoded.push(d);
    total += d.beadCount;
  }

  const srcRingAxes = nodeRingAxes();

  const positions = new Float32Array(total * 3);
  const ringAxis = new Float32Array(total * 3);
  const value = new Int32Array(total);
  const srcNodeRow = new Int32Array(total);
  let w = 0;
  let b = 0;
  for (const d of decoded) {
    const src = readEdgeSrcNodeRow(d.edgeView, 0);
    const [ax, ay, az] = srcRingAxes.get(src) ?? RING_AXIS_FALLBACK;

    for (let i = 0; i < d.beadCount; i++) {
      ringAxis[w] = ax;
      ringAxis[w + 1] = ay;
      ringAxis[w + 2] = az;
      positions[w++] = readEdgeBeadX(d.beadView, i);
      positions[w++] = readEdgeBeadY(d.beadView, i);
      positions[w++] = readEdgeBeadZ(d.beadView, i);
      value[b] = readEdgeBeadValue(d.beadView, i);
      srcNodeRow[b++] = src;
    }
  }

  lastVersion = ev;
  lastNodeVersion = nv;
  lastAgg = { positions, ringAxis, value, srcNodeRow, count: total };
  return lastAgg;
}
