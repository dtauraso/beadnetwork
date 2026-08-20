import { FOCAL_PIXELS } from "./camera-consts";

const PRIMITIVES = [
  "pivot-x", "pivot-y", "pivot-z", "r",
  "pos-phi", "pos-theta", "up-phi", "up-theta",
] as const;

export type CameraPrimitive = (typeof PRIMITIVES)[number];

export type CameraPose = Record<CameraPrimitive, number> & { focalPx: number };

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

async function readFloat64(url: string): Promise<number | undefined> {
  try {
    const res = await fetch(url, { cache: "no-store" });
    if (!res.ok) return undefined;
    const buf = await res.arrayBuffer();
    if (buf.byteLength !== 8) return undefined;
    return new DataView(buf).getFloat64(0, true);
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
  const values = await Promise.all(rels.map((rel) => readFloat64(`${b.scene}/${rel ?? ""}`)));
  const pose = { focalPx: FOCAL_PIXELS } as CameraPose;
  for (const [i, name] of PRIMITIVES.entries()) {
    const v = values[i];
    if (v === undefined) return undefined;
    pose[name] = v;
  }
  return pose;
}
