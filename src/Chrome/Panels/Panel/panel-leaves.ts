import { PANEL_FLAG_ORDER, type PanelFlag } from "../../../Input/messages";

export type PanelFlagVals = Record<PanelFlag, boolean>;

const vals = Object.fromEntries(PANEL_FLAG_ORDER.map((f) => [f, false])) as PanelFlagVals;
export function panelFlagVals(): PanelFlagVals { return vals; }

let seq = 0;
async function readBytes(url: string): Promise<ArrayBuffer | undefined> {
  return readUrl(`${url}?r=${++seq}`, "no-store");
}

async function readGenerated(url: string): Promise<ArrayBuffer | undefined> {
  return readUrl(url, "default");
}

async function readUrl(url: string, cache: RequestCache): Promise<ArrayBuffer | undefined> {
  try {
    const res = await fetch(url, { cache });
    return res.ok ? await res.arrayBuffer() : undefined;
  } catch {
    return undefined;
  }
}

async function loadPaths(src: string): Promise<Map<PanelFlag, string> | undefined> {
  const bufs = await Promise.all(
    PANEL_FLAG_ORDER.map((flag) => readGenerated(`${src}/Chrome/Panels/Panel/paths/${flag}.bin`)),
  );
  const out = new Map<PanelFlag, string>();
  for (const [i, flag] of PANEL_FLAG_ORDER.entries()) {
    const buf = bufs[i];
    if (buf === undefined) return undefined;
    out.set(flag, new TextDecoder().decode(buf));
  }
  return out;
}

const READ_INTERVAL_MS = 100;

let started = false;
export function startPanelReads(): void {
  if (started || typeof window === "undefined") return;
  started = true;
  let paths: Map<PanelFlag, string> | undefined;
  const pump = async () => {
    for (;;) {
      const scene = window.WIREFOLD_SCENE_BASE;
      const src = window.WIREFOLD_SRC_BASE;
      if (scene && src) {
        paths ??= await loadPaths(src);
        if (paths) {
          const bufs = await Promise.all(
            PANEL_FLAG_ORDER.map((f) => readBytes(`${scene}/${paths?.get(f) ?? ""}`)),
          );
          for (const [i, flag] of PANEL_FLAG_ORDER.entries()) {
            const b = bufs[i];
            if (b?.byteLength === 1) vals[flag] = new DataView(b).getUint8(0) !== 0;
          }
        }
      }
      await new Promise((r) => setTimeout(r, READ_INTERVAL_MS));
    }
  };
  void pump();
}
