import { noteSpawnGen } from "./spawn-gen";
import { SCENE_VALUES } from "./scene-values-gen";

const vals = new Map<string, number>(SCENE_VALUES.map((v) => [v.name, 0]));
export function sceneValue(name: string): number { return vals.get(name) ?? 0; }

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

async function loadPaths(src: string): Promise<Map<string, string> | undefined> {
  const bufs = await Promise.all(
    SCENE_VALUES.map((v) => readGenerated(`${src}/Categories/Scene/paths/${v.name}.bin`)),
  );
  const out = new Map<string, string>();
  for (const [i, v] of SCENE_VALUES.entries()) {
    const buf = bufs[i];
    if (buf === undefined) return undefined;
    out.set(v.name, new TextDecoder().decode(buf));
  }
  return out;
}

const READ_INTERVAL_MS = 100;

let started = false;
export function startSceneReads(): void {
  if (started || typeof window === "undefined") return;
  started = true;
  let paths: Map<string, string> | undefined;
  const pump = async () => {
    for (;;) {
      const scene = window.BEADNETWORK_SCENE_BASE;
      const src = window.BEADNETWORK_SRC_BASE;
      if (scene && src) {
        paths ??= await loadPaths(src);
        if (paths) {
          const bufs = await Promise.all(
            SCENE_VALUES.map((v) => readBytes(`${scene}/${paths?.get(v.name) ?? ""}`)),
          );
          for (const [i, v] of SCENE_VALUES.entries()) {
            const b = bufs[i];
            if (b?.byteLength !== 8) continue;
            const dv = new DataView(b);
            vals.set(v.name, v.kind === "f64" ? dv.getFloat64(0, true) : Number(dv.getBigInt64(0, true)));
          }
          noteSpawnGen(vals.get("spawn") ?? 0);
        }
      }
      await new Promise((r) => setTimeout(r, READ_INTERVAL_MS));
    }
  };
  void pump();
}
