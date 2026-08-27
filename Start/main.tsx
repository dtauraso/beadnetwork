import { vscode } from "../Start/extension/webview/vscode-api";
import { postLog } from "../Start/extension/webview/log/post";

declare global {
  interface Window { BEADNETWORK_BOOTED?: boolean }
}
window.BEADNETWORK_BOOTED = true;

postLog("lifecycle", { phase: "bundle-eval" });

import { createRoot } from "react-dom/client";
import { startSceneBaseReads } from "../Categories/Scene/scene-base";
import { ThreeView } from "../Start/extension/webview/scene/ThreeView";
import { ErrorBoundary } from "../Start/extension/webview/log/ErrorBoundary";
import { CrashListeners } from "../Start/extension/webview/log/CrashListeners";
import { armBlankReport } from "../Start/extension/webview/log/BlankReport";
import { ownerCounts } from "../Categories/Scene/owner-counts";


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


vscode.postMessage({ type: "ready" });
postLog("lifecycle", { phase: "ready-sent" });

armBlankReport(() => {
  const { nodes, edges } = ownerCounts();
  return nodes > 0 || edges > 0;
});
