import { FOCAL_PIXELS } from "./camera-consts";
import { makeLeafValues } from "./leaf-values";

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
    BEADNETWORK_SCENE_BASE?: string;
    BEADNETWORK_SRC_BASE?: string;
  }
}

const values = makeLeafValues<CameraValueName>(
  "Categories/Scene/Camera/paths",
  VALUE_NAMES,
);

export function readCameraPose(): CameraPose | undefined {
  const pose = { focalPx: FOCAL_PIXELS } as CameraPose;
  for (const name of VALUE_NAMES) {
    const v = values.bytes(name);
    if (!v || v.byteLength < 8) return undefined;
    pose[name] = v.getFloat64(0, true);
  }
  return pose;
}
