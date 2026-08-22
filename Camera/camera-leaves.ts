import { FOCAL_PIXELS } from "./camera-consts";

const VALUE_NAMES = [
  "pivotX", "pivotY", "pivotZ",
  "r",
  "posPhi", "posTheta",
  "upPhi", "upTheta",
] as const;

type CameraValueName = (typeof VALUE_NAMES)[number];

export type CameraPose = Record<CameraValueName, number> & { focalPx: number };

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

export async function loadCameraBlockPath(): Promise<string | undefined> {
  const b = bases();
  if (!b) return undefined;
  try {
    const res = await fetch(`${b.src}/Camera/paths/block.bin`, { cache: "default" });
    if (!res.ok) return undefined;
    return await res.text();
  } catch {
    return undefined;
  }
}

export async function readCameraPose(blockPath: string): Promise<CameraPose | undefined> {
  const b = bases();
  if (!b) return undefined;
  let buf: ArrayBuffer;
  try {
    const res = await fetch(`${b.scene}/${blockPath}`, { cache: "no-store" });
    if (!res.ok) return undefined;
    buf = await res.arrayBuffer();
  } catch {
    return undefined;
  }

  const dv = new DataView(buf);
  const pose = { focalPx: FOCAL_PIXELS } as CameraPose;
  let off = 0;
  for (const name of VALUE_NAMES) {
    if (off + 4 > buf.byteLength) return undefined;
    const len = dv.getUint32(off, true);
    off += 4;
    if (len < 8 || off + len > buf.byteLength) return undefined;
    pose[name] = dv.getFloat64(off, true);
    off += len;
  }
  return pose;
}
