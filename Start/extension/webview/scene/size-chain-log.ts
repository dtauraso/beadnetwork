
import { postLog } from "../log/post";

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
