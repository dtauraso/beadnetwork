import { FOCAL_PIXELS } from "./camera-consts";

const VEC_LEN: Record<string, number> = { pivot: 3 };

const PRIMITIVES = [
  "pivot", "r",
  "pos-phi", "pos-theta", "up-phi", "up-theta",
] as const;

export type CameraPrimitive = (typeof PRIMITIVES)[number];

export type CameraPose = Record<Exclude<CameraPrimitive, "pivot">, number> & {
  pivotX: number; pivotY: number; pivotZ: number;
  focalPx: number;
};

declare global {
  interface Window {
    WIREFOLD_SCENE_BASE?: string;
    WIREFOLD_SRC_BASE?: string;
  }
}

function bases(): { scene: string; src: string } | undefined {
  const scene = typeof window === "undefined" ? undefined : window.WIREFOLD_SCENE_BASE;
  const src = typeof window === "undefined" ? undefined : window.WIREFOLD_SRC_BASE;
  if (!scene || !src) return undefined;
  return { scene, src };
}

async function readText(url: string): Promise<string | undefined> {
  try {
    const res = await fetch(url, { cache: "no-store" });
    if (!res.ok) return undefined;
    return await res.text();
  } catch {
    return undefined;
  }
}

async function readFloats(url: string, n: number): Promise<number[] | undefined> {
  try {
    const res = await fetch(url, { cache: "no-store" });
    if (!res.ok) return undefined;
    const buf = await res.arrayBuffer();
    if (buf.byteLength !== n * 8) return undefined;
    const dv = new DataView(buf);
    const out: number[] = [];
    for (let i = 0; i < n; i++) out.push(dv.getFloat64(i * 8, true));
    return out;
  } catch {
    return undefined;
  }
}

export async function loadCameraPaths(): Promise<Map<CameraPrimitive, string> | undefined> {
  const b = bases();
  if (!b) return undefined;
  const out = new Map<CameraPrimitive, string>();
  for (const name of PRIMITIVES) {
    const rel = await readText(`${b.src}/Camera/paths/${name}.bin`);
    if (rel === undefined) return undefined;
    out.set(name, rel);
  }
  return out;
}

export async function readCameraPose(paths: Map<CameraPrimitive, string>): Promise<CameraPose | undefined> {
  const b = bases();
  if (!b) return undefined;
  const rels = PRIMITIVES.map((name) => paths.get(name));
  if (rels.some((rel) => rel === undefined)) return undefined;
  const values = await Promise.all(
    PRIMITIVES.map((name, i) => readFloats(`${b.scene}/${rels[i] ?? ""}`, VEC_LEN[name] ?? 1)),
  );
  const pose = { focalPx: FOCAL_PIXELS } as CameraPose;
  for (const [i, name] of PRIMITIVES.entries()) {
    const v = values[i];
    if (v === undefined) return undefined;
    if (name === "pivot") {
      [pose.pivotX, pose.pivotY, pose.pivotZ] = [v[0] ?? 0, v[1] ?? 0, v[2] ?? 0];
    } else {
      pose[name] = v[0] ?? 0;
    }
  }
  return pose;
}
