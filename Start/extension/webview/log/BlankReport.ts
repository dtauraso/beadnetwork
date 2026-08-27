import { postLog } from "./post";

// WHY THIS EXISTS
//
// A blank panel has had three different causes so far — the scene base pointing
// at a tree nobody writes, the scene directory not being an authorized webview
// resource, and Go still building on spawn — and every one of them looks
// identical from the outside: an empty window, a healthy Go, green guards.
//
// The error boundary covers a render throw and html.ts covers a bundle that
// never ran. Neither covers "everything ran and there is still nothing there",
// which is the case that keeps costing a session. This reports that case in
// PLAIN DOM, so it survives a dead WebGL context and an unmounted tree.

const DELAY_MS = 2500;

function line(text: string): HTMLDivElement {
  const d = document.createElement("div");
  d.textContent = text;
  return d;
}

async function reachable(url: string): Promise<string> {
  try {
    const res = await fetch(`${url}?probe=${Date.now()}`, { cache: "no-store" });
    return res.ok ? `${res.status} ok, ${(await res.arrayBuffer()).byteLength} bytes` : `${res.status}`;
  } catch (e) {
    return `unreachable (${(e as Error).message})`;
  }
}

export function armBlankReport(drewSomething: () => boolean): void {
  window.setTimeout(() => { void report(drewSomething); }, DELAY_MS);
}

async function report(drewSomething: () => boolean): Promise<void> {
  {
    if (drewSomething()) return;

    const base = (window as unknown as { BEADNETWORK_SCENE_BASE?: string }).BEADNETWORK_SCENE_BASE ?? "";
    const anchor = (window as unknown as { BEADNETWORK_ANCHOR_BASE?: string }).BEADNETWORK_ANCHOR_BASE ?? "";

    const counts = await reachable(`${base}/view/owner-counts.bin`);
    const strip = await reachable(`${base}/view/chrome/tab-strip.bin`);
    const selected = await reachable(`${anchor}/view/scene/selected.bin`);

    postLog("blank-panel", { base, counts, strip, selected });

    const box = document.createElement("div");
    box.style.cssText =
      "position:fixed;left:0;top:0;right:0;z-index:9999;padding:12px 16px;" +
      "font:12px ui-monospace,monospace;color:#e7e7ea;background:#2f2f37;" +
      "border-bottom:1px solid #3a3a44;white-space:pre-wrap";
    box.appendChild(line("nothing has been drawn. what the webview can see:"));
    box.appendChild(line(`  scene base      ${base || "(none)"}`));
    box.appendChild(line(`  owner-counts    ${counts}`));
    box.appendChild(line(`  tab-strip       ${strip}`));
    box.appendChild(line(`  selected.bin    ${selected}`));
    box.appendChild(line("a 401/404 on these is the scene directory not being readable;"));
    box.appendChild(line("'ok, 0 bytes' or a stale read is Go not writing this tree."));
    document.body.appendChild(box);
  }
}
