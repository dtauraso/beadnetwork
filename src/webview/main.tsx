import { vscode } from "./vscode-api";
import { postLog } from "./log/post";

declare global {
  interface Window { WIREFOLD_BOOTED?: boolean }
}
window.WIREFOLD_BOOTED = true;

postLog("lifecycle", { phase: "bundle-eval" });

import { createRoot } from "react-dom/client";
import { startSceneBaseReads } from "./scene-base";
import { noteSpawnGen } from "./spawn-gen";
import { ThreeView } from "./scene/ThreeView";
import { parseHostToWebview } from "../Input/messages";
import { ErrorBoundary } from "./log/ErrorBoundary";
import { CrashListeners } from "./log/CrashListeners";
import { setLatestViewFrame, setLatestEdgeStreamFrame, setLatestNodeStreamFrame, setLatestInteriorStreamFrame, setLatestBeadStreamFrame } from "./snapshot-buffer";
import { BUF_BLOCK_TAG_VIEW, BUF_BLOCK_TAG_EDGE_STREAM, BUF_BLOCK_TAG_NODE_STREAM, BUF_BLOCK_TAG_INTERIOR_STREAM, BUF_BLOCK_TAG_BEAD_STREAM, BUF_BLOCK_TAG_COLUMN } from "../Buffer/frame-tags";
import { setColumnValue, columnDiagnostics, columnI32, columnF32, columnU8 } from "../Buffer/column-values";
import { nodeColumn, edgeColumn, ownerCounts } from "../Buffer/column-owners";
import { edgeF32 } from "../Node/Edge/edge-leaves";
import { nodeI32, nodeU8 } from "../Node/node-leaves";

function Root() {
  return <ThreeView />;
}

startSceneBaseReads();

window.addEventListener("pointerdown", (e) => {
  const app = document.getElementById("app");
  const under = document.elementFromPoint(e.clientX, e.clientY);
  const r = app?.getBoundingClientRect();
  postLog("doc-pointer-down", {
    xy: `${Math.round(e.clientX)},${Math.round(e.clientY)}`,
    under: under ? `${under.tagName}${under.id ? "#" + under.id : ""}` : "none",
    appRect: r ? `${Math.round(r.width)}x${Math.round(r.height)}` : "none",
    appClient: app ? `${app.clientWidth}x${app.clientHeight}` : "none",
    docClient: `${document.documentElement.clientWidth}x${document.documentElement.clientHeight}`,
    win: `${window.innerWidth}x${window.innerHeight}`,
  });
}, true);

postLog("lifecycle", { phase: "before-render" });
const app = document.getElementById("app")!;
createRoot(app).render(
  <ErrorBoundary>
    <CrashListeners />
    <Root />
  </ErrorBoundary>,
);

let bufSnapLogAt = 0;
let bufSnapCount = 0;

window.addEventListener("message", (e) => {
  const msg = parseHostToWebview(e.data);
  if (!msg) return;

  if (msg.type === "buffer-snapshot") {
    noteSpawnGen(msg.gen);

    if (msg.tag === BUF_BLOCK_TAG_VIEW) {
      setLatestViewFrame(msg.buffer, msg.gen);
    } else if (msg.tag === BUF_BLOCK_TAG_EDGE_STREAM) {

      if (typeof msg.row === "number") {
        setLatestEdgeStreamFrame(msg.row, msg.buffer, msg.gen);
      }
    } else if (msg.tag === BUF_BLOCK_TAG_NODE_STREAM) {

      if (typeof msg.row === "number") {
        setLatestNodeStreamFrame(msg.row, msg.buffer, msg.gen);
      }
    } else if (msg.tag === BUF_BLOCK_TAG_INTERIOR_STREAM) {

      if (typeof msg.row === "number") {
        setLatestInteriorStreamFrame(msg.row, msg.buffer, msg.gen);
      }
    } else if (msg.tag === BUF_BLOCK_TAG_BEAD_STREAM) {

      if (typeof msg.row === "number") {
        setLatestBeadStreamFrame(msg.row, msg.buffer, msg.gen);
      }
    } else if (msg.tag === BUF_BLOCK_TAG_COLUMN) {

      if (typeof msg.row === "number") {
        setColumnValue(msg.row, msg.buffer);
      }
    }
    bufSnapCount += 1;
    const now = Date.now();
    if (now - bufSnapLogAt >= 1000) {
      const cols = columnDiagnostics();
      const counts = ownerCounts();
      postLog("buf-snapshot", {
        byteLength: msg.buffer.byteLength, sinceLast: bufSnapCount, windowMs: now - bufSnapLogAt,

        colsReceived: cols.received, colVersion: cols.version,
        colLowest: cols.lowest, colHighest: cols.highest,
        ownerNodes: counts.nodes, ownerEdges: counts.edges,
        nodeIndex: Array.from({ length: counts.nodes }, (_unused, row) => [
          nodeU8(row, "hasPos"),
          nodeI32(row, "indexR"),
          nodeI32(row, "indexPhi"),
          nodeI32(row, "indexTheta"),
        ].join("/")),
        edgeStart: Array.from({ length: counts.edges }, (_unused, row) =>
          edgeF32(row, "sx").toFixed(1)),
      });
      bufSnapLogAt = now;
      bufSnapCount = 0;
    }
  }
});

vscode.postMessage({ type: "ready" });
postLog("lifecycle", { phase: "ready-sent" });
