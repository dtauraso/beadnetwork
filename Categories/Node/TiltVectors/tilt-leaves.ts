import { nodeBytes } from "../node-leaves";
import { TILT_VALUE_NAMES, type TiltValueName } from "./tilt-values-gen";

export const tiltArrowBytes = (row: number, name: TiltValueName) => nodeBytes(row, name);

export const TILT_SHAFT_NAMES = TILT_VALUE_NAMES.slice(1, 17) as readonly TiltValueName[];
export const TILT_HEAD_NAMES = TILT_VALUE_NAMES.slice(17) as readonly TiltValueName[];
