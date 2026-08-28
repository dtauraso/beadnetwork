import { noteSpawnGen } from "./spawn-gen";
import { SCENE_VALUES } from "./scene-values-gen";

const vals = new Map<string, number>(SCENE_VALUES.map((v) => [v.name, 0]));
export function sceneValue(name: string): number { return vals.get(name) ?? 0; }

interface SceneBlockMsg {
  pathsDir: string;
  pathFile?: string;
  b64: string;
}

interface SceneBlocks {
  on: (fn: (m: SceneBlockMsg) => void) => void;
  wantFile: (pathsDir: string, pathFile: string, cadenceMs: number) => void;
  bytes: (b64: string) => ArrayBuffer;
}

function sceneBlocks(): SceneBlocks | undefined {
  const w = window as unknown as { BEADNETWORK_BLOCKS?: Partial<SceneBlocks> };
  const b = w.BEADNETWORK_BLOCKS;
  return b?.on && b.wantFile && b.bytes ? (b as SceneBlocks) : undefined;
}

const PATHS_DIR = "Categories/Scene/paths";
const READ_INTERVAL_MS = 100;

const kindOf = new Map(SCENE_VALUES.map((v) => [`${v.name}.bin`, v.kind]));
const nameOf = new Map(SCENE_VALUES.map((v) => [`${v.name}.bin`, v.name]));

let started = false;
export function startSceneReads(): void {
  if (started || typeof window === "undefined") return;
  const blocks = sceneBlocks();
  if (!blocks) return;
  started = true;

  blocks.on((m) => {
    if (m.pathsDir !== PATHS_DIR) return;
    const name = nameOf.get(m.pathFile ?? "");
    if (name === undefined) return;
    const buf = blocks.bytes(m.b64);
    if (buf.byteLength !== 8) return;
    const dv = new DataView(buf);
    vals.set(name, kindOf.get(`${name}.bin`) === "f64"
      ? dv.getFloat64(0, true)
      : Number(dv.getBigInt64(0, true)));
    if (name === "spawn") noteSpawnGen(vals.get("spawn") ?? 0);
  });

  for (const v of SCENE_VALUES) blocks.wantFile(PATHS_DIR, `${v.name}.bin`, READ_INTERVAL_MS);
}
