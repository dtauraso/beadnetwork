import { makeLeafValues } from "./leaf-values";
import { ANGLE_PILL_VALUE_NAMES, type AnglePillValueName } from "./pill-values-gen";

const values = makeLeafValues<AnglePillValueName>(
  "Chrome/Pills/AngleDropdown/paths",
  ANGLE_PILL_VALUE_NAMES,
);

export const anglePillBytes = values.bytes;
export const anglePillF32 = values.f32;
export const anglePillI32 = values.i32;
export const anglePillU32 = values.u32;
export const anglePillU8 = values.u8;
export const anglePillF32Run = values.f32Run;
export const anglePillI32Run = values.i32Run;
export const anglePillU32Run = values.u32Run;
export const anglePillText = values.text;
