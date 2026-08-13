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

  // srcNodeRow per bead — the node whose edge it is travelling. It is what
  // lets a renderer colour a bead by the kind of node that sent it without
  // walking the frames a second time.
  srcNodeRow: Int32Array;

  // ringAxis is what the bead's torus faces along: the ring axis of the node
  // the bead came FROM, streamed on that node's own frame. It is not the
  // direction of travel — a bead is threaded on its source node's ring, and
  // aiming it down the edge instead turns every torus the wrong way.
  ringAxis: Float32Array;

  count: number;
}

const RING_AXIS_FALLBACK: [number, number, number] = [0, 1, 0];

// nodeRingAxes is each node row's own ring axis, off that node's own frame.
// A bead's torus is aimed by the node it came from, so the aggregate has to
// read the node stream as well as the edge stream — the two angles are
// streamed, and poleAxis only turns them into the vector they already name.
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

// getEdgeBeads is every in-flight bead in the scene, flattened across edges.
// The positions are WORLD positions: an edge places its own beads along its
// own segment, so there is no centre to add them to.
export function getEdgeBeads(): EdgeBeadsAgg {
  // Two streams feed this now: the beads from the edges, their aim from the
  // nodes. A node turning without a bead moving still re-aims every torus,
  // so the cache has to watch both versions.
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
