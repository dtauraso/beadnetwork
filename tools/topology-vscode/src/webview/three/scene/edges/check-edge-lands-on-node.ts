import { columnF32 } from "../../../../../Buffer/column-values";
import { nodeColumn, ownerCounts } from "../../../../../Buffer/column-owners";
import {
  COL_STREAM_NODE_POLE_ANCHOR_X, COL_STREAM_NODE_POLE_ANCHOR_Y, COL_STREAM_NODE_POLE_ANCHOR_Z,
} from "../../../../../Buffer/column-streams-gen";
import { getEdgeStreamAccessor } from "./edge-stream-blocks";
import { sceneSteps } from "../../../../../Scene/scene-frame";
import { postLog } from "../../../log/post";

const TOLERANCE = 0.25;

const reported = new Set<number>();

export function checkEdgeLandsOnNode(): void {
  const edges = getEdgeStreamAccessor();
  if (!edges) return;
  const { nodes: nodeCount } = ownerCounts();

  for (let edgeRow = 0; edgeRow < edges.edgeCount; edgeRow++) {
    if (reported.has(edgeRow)) continue;
    const dst = edges.dstNodeRow(edgeRow);
    if (dst < 0 || dst >= nodeCount) continue;

    const seg = edges.segment(edgeRow);
    const ex = seg[3] ?? 0;
    const ey = seg[4] ?? 0;
    const ez = seg[5] ?? 0;
    if (ex === 0 && ey === 0 && ez === 0) continue;

    const cx = columnF32(nodeColumn(dst, COL_STREAM_NODE_POLE_ANCHOR_X));
    const cy = columnF32(nodeColumn(dst, COL_STREAM_NODE_POLE_ANCHOR_Y));
    const cz = columnF32(nodeColumn(dst, COL_STREAM_NODE_POLE_ANCHOR_Z));
    if (cx === 0 && cy === 0 && cz === 0) continue;

    const gap = Math.hypot(ex - cx, ey - cy, ez - cz);
    const step = sceneSteps().constantR;
    if (!(step > 0)) continue;
    const steps = gap / step;
    if (Math.abs(steps - Math.round(steps)) <= TOLERANCE) continue;

    reported.add(edgeRow);
    postLog("edge-misses-node", {
      edgeRow,
      dstNodeRow: dst,
      gap: gap.toFixed(3),
      steps: steps.toFixed(3),
      end: `${ex.toFixed(2)},${ey.toFixed(2)},${ez.toFixed(2)}`,
      centre: `${cx.toFixed(2)},${cy.toFixed(2)},${cz.toFixed(2)}`,
      why:
        "an edge end is pulled back from its target's centre by that node's torus radius, " +
        "which is a whole number of lattice steps; a fractional gap means the edge was " +
        "aimed somewhere the node is not",
    });
  }
}
