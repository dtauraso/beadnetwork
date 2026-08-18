import { getNodeFrame } from "../nodes/node-frame-aggregate";
import { getEdgeStreamAccessor } from "./edge-stream-blocks";
import { nodeCenterX, nodeCenterY, nodeCenterZ } from "../../../../../Node/node-frame";
import { readSceneConstantR } from "../../../../../Buffer/buffer-layout";
import { getViewBlocks } from "../view-blocks";
import { postLog } from "../../../log/post";


const TOLERANCE = 0.25;

const reported = new Set<number>();

export function checkEdgeLandsOnNode(): void {
  const decoded = getNodeFrame();
  const edges = getEdgeStreamAccessor();
  const blocks = getViewBlocks();
  if (!decoded || !edges) return;
  const { nodeCount, nodeView } = decoded;

  for (let edgeRow = 0; edgeRow < edges.edgeCount; edgeRow++) {
    if (reported.has(edgeRow)) continue;
    const dst = edges.dstNodeRow(edgeRow);
    if (dst < 0 || dst >= nodeCount) continue;

    const seg = edges.segment(edgeRow);
    const ex = seg[3] ?? 0;
    const ey = seg[4] ?? 0;
    const ez = seg[5] ?? 0;
    if (ex === 0 && ey === 0 && ez === 0) continue;

    const cx = nodeCenterX(nodeView, dst);
    const cy = nodeCenterY(nodeView, dst);
    const cz = nodeCenterZ(nodeView, dst);
    if (cx === 0 && cy === 0 && cz === 0) continue;

    const gap = Math.hypot(ex - cx, ey - cy, ez - cz);
    const step = blocks ? readSceneConstantR(blocks.sceneView) : 0;
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
