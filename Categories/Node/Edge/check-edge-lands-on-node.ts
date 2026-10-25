import { ownerCounts } from "../../Scene/owner-counts";
import { nodeF32 } from "../node-leaves";
import { getEdgeStreamAccessor } from "./edge-stream-blocks";
import { SHADING_PARAM_NODE_RING_TUBE_RATIO } from "../shading-params";
import { postLog } from "../../../Start/extension/webview/log/post";

const TOLERANCE = 0.05;

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

    const cx = nodeF32(dst, "poleAnchorX");
    const cy = nodeF32(dst, "poleAnchorY");
    const cz = nodeF32(dst, "poleAnchorZ");
    if (cx === 0 && cy === 0 && cz === 0) continue;

    const gap = Math.hypot(ex - cx, ey - cy, ez - cz);
    const torusR = nodeF32(dst, "radius") * (1 + SHADING_PARAM_NODE_RING_TUBE_RATIO);
    if (!(torusR > 0)) continue;
    if (Math.abs(gap - torusR) <= TOLERANCE) continue;

    reported.add(edgeRow);
    postLog("edge-misses-node", {
      edgeRow,
      dstNodeRow: dst,
      gap: gap.toFixed(3),
      torusR: torusR.toFixed(3),
      end: `${ex.toFixed(2)},${ey.toFixed(2)},${ez.toFixed(2)}`,
      centre: `${cx.toFixed(2)},${cy.toFixed(2)},${cz.toFixed(2)}`,
      why:
        "an edge end is pulled back from its target's centre by exactly that node's torus " +
        "radius; any other gap means the edge was aimed somewhere the node is not — a node " +
        "on a float place of the scene sphere is not on a whole grid step, so the gap is " +
        "measured against the torus radius, not against the scene's radial step",
    });
  }
}
