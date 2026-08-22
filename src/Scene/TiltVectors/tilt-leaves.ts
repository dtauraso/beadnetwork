import { makeRowLeafValues } from "./row-leaf-values";
import { TILT_VALUE_NAMES, type TiltValueName } from "./tilt-values-gen";

const values = makeRowLeafValues<TiltValueName>(
  "Scene/TiltVectors/paths",
  TILT_VALUE_NAMES,
);

export const tiltArrowBytes = values.bytes;

export const TILT_SHAFT_NAMES = TILT_VALUE_NAMES.slice(1, 17) as readonly TiltValueName[];
export const TILT_HEAD_NAMES = TILT_VALUE_NAMES.slice(17) as readonly TiltValueName[];
