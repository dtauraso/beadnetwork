
const phases: string[] = [];
let box: HTMLElement | null = null;
let failed = false;

function paint(): void {
  try {
    const marker = document.getElementById("html-marker");
    if (marker) marker.remove();
    if (!box) {
      box = document.createElement("div");
      box.id = "boot-trace";
      box.setAttribute(
        "style",
        "position:fixed;left:0;top:0;z-index:2147483647;max-width:100vw;max-height:100vh;" +
          "overflow:auto;margin:0;padding:8px;font:11px ui-monospace,monospace;" +
          "white-space:pre-wrap;color:#9cdcfe;background:rgba(0,0,0,.85)",
      );
      document.body.appendChild(box);
    }
    box.textContent = phases.join("\n");
  } catch {
    box = null;
  }
}

export function bootPhase(name: string, detail?: string): void {
  phases.push(detail === undefined ? name : `${name}  ${detail}`);
  paint();
  if (name === "ThreeView body" && !failed) {
    setTimeout(() => {
      if (failed) return;
      box?.remove();
      box = null;
    }, 2000);
  }
}

export function bootFail(what: string, err: unknown): void {
  failed = true;
  const e = err as { message?: string; stack?: string } | undefined;
  phases.push(`!! ${what}: ${e?.message ?? String(err)}`);
  if (e?.stack) phases.push(e.stack.split("\n").slice(0, 6).join("\n"));
  paint();
}

window.addEventListener("error", (e: ErrorEvent) => {
  bootFail(`window error ${e.filename ?? ""}:${e.lineno ?? 0}`, e.error ?? e.message);
});

window.addEventListener("unhandledrejection", (e: PromiseRejectionEvent) => {
  bootFail("unhandled rejection", e.reason);
});

bootPhase("boot-trace installed");
