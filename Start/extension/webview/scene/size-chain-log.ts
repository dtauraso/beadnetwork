
import { postLog } from "../log/post";
import { ownerCounts } from "../../../../Categories/Scene/owner-counts";

type Row = { el: string; client: string; rect: string; css: string };

function describe(label: string, el: Element | null): Row {
  if (!el) return { el: label, client: "absent", rect: "absent", css: "absent" };
  const r = el.getBoundingClientRect();
  const s = getComputedStyle(el);
  return {
    el: label,
    client: `${el.clientWidth}x${el.clientHeight}`,
    rect: `${Math.round(r.width)}x${Math.round(r.height)}`,
    css: `${s.position} h=${s.height} inset=${s.top}/${s.bottom} disp=${s.display}`,
  };
}

export function logSceneContent(where: string): void {
  let counts: { nodes: number; edges: number } | string;
  try {
    counts = ownerCounts();
  } catch (err) {
    counts = err instanceof Error ? err.message : String(err);
  }
  const w = window as unknown as Record<string, unknown>;
  postLog("scene-content", {
    where,
    nodes: typeof counts === "string" ? counts : counts.nodes,
    edges: typeof counts === "string" ? "-" : counts.edges,
    sceneBase: typeof w.BEADNETWORK_SCENE_BASE === "string" ? w.BEADNETWORK_SCENE_BASE : "unset",
    srcBase: typeof w.BEADNETWORK_SRC_BASE === "string" ? w.BEADNETWORK_SRC_BASE : "unset",
    canvases: document.querySelectorAll("canvas").length,
  });
}

export async function logProbeFetch(where: string): Promise<void> {
  const w = window as unknown as Record<string, unknown>;
  const src = typeof w.BEADNETWORK_SRC_BASE === "string" ? w.BEADNETWORK_SRC_BASE : undefined;
  const scene = typeof w.BEADNETWORK_SCENE_BASE === "string" ? w.BEADNETWORK_SCENE_BASE : undefined;
  if (!src || !scene) {
    postLog("probe-fetch", { where, note: "bases unset" });
    return;
  }
  for (const url of [
    `${src}/Categories/Scene/Camera/paths/block.bin`,
    `${src}/Categories/Scene/paths/constantR.bin`,
    `${scene}/view/camera.bin`,
    `${scene}/view/owner-counts.bin`,
  ]) {
    try {
      const res = await fetch(url, { cache: "no-store" });
      const body = res.ok ? await res.arrayBuffer() : undefined;
      postLog("probe-fetch", {
        where, url, status: res.status, ok: res.ok, bytes: body ? body.byteLength : 0,
      });
    } catch (err) {
      postLog("probe-fetch", {
        where, url, status: "threw", ok: false,
        err: err instanceof Error ? err.message : String(err),
      });
    }
  }
}

export function logSizeChain(where: string, container: Element | null): void {
  const canvas = document.querySelector("canvas");
  const rows: Row[] = [
    describe("html", document.documentElement),
    describe("body", document.body),
    describe("#app", document.getElementById("app")),
    describe("container", container),
    describe("canvasParent", canvas?.parentElement ?? null),
    describe("canvas", canvas),
  ];
  for (const row of rows) {
    postLog("size-chain", { where, ...row });
  }
  if (canvas instanceof HTMLCanvasElement) {
    postLog("size-chain-canvas-attrs", {
      where,
      attrW: canvas.getAttribute("width") ?? "none",
      attrH: canvas.getAttribute("height") ?? "none",
      bufferW: canvas.width,
      bufferH: canvas.height,
      dpr: window.devicePixelRatio,
    });
  }
}
