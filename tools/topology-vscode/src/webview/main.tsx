import { vscode } from "./vscode-api";
import { postLog } from "./log/post";
postLog("lifecycle", { phase: "bundle-eval" });

import { createRoot } from "react-dom/client";
import { ThreeView } from "./three/scene/ThreeView";
import { parseHostToWebview } from "../messages";
import { ErrorBoundary } from "./log/ErrorBoundary";
import { CrashListeners } from "./log/CrashListeners";
import { setLatestViewFrame, setLatestEdgeStreamFrame, setLatestNodeStreamFrame, setLatestInteriorStreamFrame, setLatestBeadStreamFrame } from "./snapshot-buffer";
import { BUF_BLOCK_TAG_VIEW, BUF_BLOCK_TAG_EDGE_STREAM, BUF_BLOCK_TAG_NODE_STREAM, BUF_BLOCK_TAG_INTERIOR_STREAM, BUF_BLOCK_TAG_BEAD_STREAM, BUF_BLOCK_TAG_COLUMN } from "../Buffer/frame-tags";
import { setColumnValue, columnDiagnostics, columnI32, columnF32, columnU8 } from "../Buffer/column-values";
import { nodeColumn, edgeColumn, ownerCounts } from "../Buffer/column-owners";
import {
  COL_STREAM_NODE_INDEX_R, COL_STREAM_NODE_INDEX_PHI, COL_STREAM_NODE_INDEX_THETA,
  COL_STREAM_NODE_HAS_POS, COL_STREAM_EDGE_SX,
} from "../Buffer/column-streams-gen";

function Root() {
  return <ThreeView />;
}

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
          columnU8(nodeColumn(row, COL_STREAM_NODE_HAS_POS)),
          columnI32(nodeColumn(row, COL_STREAM_NODE_INDEX_R)),
          columnI32(nodeColumn(row, COL_STREAM_NODE_INDEX_PHI)),
          columnI32(nodeColumn(row, COL_STREAM_NODE_INDEX_THETA)),
        ].join("/")),
        edgeStart: Array.from({ length: counts.edges }, (_unused, row) =>
          columnF32(edgeColumn(row, COL_STREAM_EDGE_SX)).toFixed(1)),
      });
      bufSnapLogAt = now;
      bufSnapCount = 0;
    }
  }
});

vscode.postMessage({ type: "ready" });
postLog("lifecycle", { phase: "ready-sent" });
