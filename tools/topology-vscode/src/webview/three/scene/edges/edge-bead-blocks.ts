import { getLatestEdgeStreamFrames, getEdgeStreamVersion } from "../../../snapshot-buffer";
import { decodeEdgeStreamFrame } from "../../decode/buffer-decode-edge";
import {
  readEdgeBeadX, readEdgeBeadY, readEdgeBeadZ, readEdgeBeadValue,
  readEdgeSrcNodeRow,
} from "../../../../schema/buffer-layout/buffer-layout";

export interface EdgeBeadsAgg {
  positions: Float32Array;
  value: Int32Array;

  // srcNodeRow per bead — the node whose edge it is travelling. It is what
  // lets a renderer colour a bead by the kind of node that sent it without
  // walking the frames a second time.
  srcNodeRow: Int32Array;

  count: number;
}

let lastVersion = -1;
let lastAgg: EdgeBeadsAgg | null = null;

// getEdgeBeads is every in-flight bead in the scene, flattened across edges.
// The positions are WORLD positions: an edge places its own beads along its
// own segment, so there is no centre to add them to.
export function getEdgeBeads(): EdgeBeadsAgg {
  const ev = getEdgeStreamVersion();
  if (lastAgg !== null && ev === lastVersion) return lastAgg;

  const decoded = [];
  let total = 0;
  for (const [row, buf] of getLatestEdgeStreamFrames()) {
    const d = decodeEdgeStreamFrame(row, buf);
    if (!d) continue;
    decoded.push(d);
    total += d.beadCount;
  }

  const positions = new Float32Array(total * 3);
  const value = new Int32Array(total);
  const srcNodeRow = new Int32Array(total);
  let w = 0;
  let b = 0;
  for (const d of decoded) {
    const src = readEdgeSrcNodeRow(d.edgeView, 0);
    for (let i = 0; i < d.beadCount; i++) {
      positions[w++] = readEdgeBeadX(d.beadView, i);
      positions[w++] = readEdgeBeadY(d.beadView, i);
      positions[w++] = readEdgeBeadZ(d.beadView, i);
      value[b] = readEdgeBeadValue(d.beadView, i);
      srcNodeRow[b++] = src;
    }
  }

  lastVersion = ev;
  lastAgg = { positions, value, srcNodeRow, count: total };
  return lastAgg;
}
