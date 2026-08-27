import { bootPhase, bootFail } from "../Start/extension/webview/log/boot-trace";
import { vscode } from "../Start/extension/webview/vscode-api";
import { postLog } from "../Start/extension/webview/log/post";

declare global {
  interface Window { BEADNETWORK_BOOTED?: boolean }
}
window.BEADNETWORK_BOOTED = true;

bootPhase("bundle-eval");
bootPhase("bases", `scene=${window.BEADNETWORK_SCENE_BASE ? "set" : "MISSING"} src=${window.BEADNETWORK_SRC_BASE ? "set" : "MISSING"}`);
postLog("lifecycle", { phase: "bundle-eval" });

import { createRoot } from "react-dom/client";
import { startSceneBaseReads } from "../Categories/Scene/scene-base";
import { ThreeView } from "../Start/extension/webview/scene/ThreeView";
import { ErrorBoundary } from "../Start/extension/webview/log/ErrorBoundary";
import { CrashListeners } from "../Start/extension/webview/log/CrashListeners";


function Root() {
  return <ThreeView />;
}

try {
  startSceneBaseReads();
  bootPhase("scene-base reads started");
} catch (err) {
  bootFail("startSceneBaseReads", err);
}

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
bootPhase("before-render");
const app = document.getElementById("app");
bootPhase("#app", app ? `${app.clientWidth}x${app.clientHeight}` : "MISSING");
try {
  createRoot(app!).render(
    <ErrorBoundary>
      <CrashListeners />
      <Root />
    </ErrorBoundary>,
  );
  bootPhase("render called");
} catch (err) {
  bootFail("createRoot/render", err);
}

vscode.postMessage({ type: "ready" });
bootPhase("ready sent");
postLog("lifecycle", { phase: "ready-sent" });
