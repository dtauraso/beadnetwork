

import { vscode } from "./vscode-api";
import { postLog } from "./log/post";
postLog("lifecycle", { phase: "bundle-eval" });

import { createRoot } from "react-dom/client";
import "./webview.css";
import { ThreeView } from "./three/scene/ThreeView";
import { SpeedSlider } from "./three/controls/panels/SpeedSlider";
import { TiltVectorButtons } from "./three/controls/panels/TiltVectorButtons";
import { parseHostToWebview } from "../messages";
import { ErrorBoundary } from "./log/ErrorBoundary";
import { CrashListeners } from "./log/CrashListeners";
import { setLatestViewFrame, setLatestEdgeStreamFrame, setLatestNodeStreamFrame, setLatestInteriorStreamFrame } from "./snapshot-buffer";
import { BUF_BLOCK_TAG_VIEW, BUF_BLOCK_TAG_EDGE_STREAM, BUF_BLOCK_TAG_NODE_STREAM, BUF_BLOCK_TAG_INTERIOR_STREAM } from "../schema/frame-tags";

function Root() {
  return (
    <>
      <ThreeView />
      <SpeedSlider />
      <TiltVectorButtons />
    </>
  );
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
    }
    bufSnapCount += 1;
    const now = Date.now();
    if (now - bufSnapLogAt >= 1000) {
      postLog("buf-snapshot", { byteLength: msg.buffer.byteLength, sinceLast: bufSnapCount, windowMs: now - bufSnapLogAt });
      bufSnapLogAt = now;
      bufSnapCount = 0;
    }
  }
});


vscode.postMessage({ type: "ready" });
postLog("lifecycle", { phase: "ready-sent" });
