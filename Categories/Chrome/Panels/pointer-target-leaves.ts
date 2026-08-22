import { makeLeafValues } from "./leaf-values";
import { POINTER_TARGET_VALUE_NAMES, type PointerTargetValueName } from "./pointer-target-values-gen";

const values = makeLeafValues<PointerTargetValueName>(
  "Categories/Chrome/Panels/paths",
  POINTER_TARGET_VALUE_NAMES,
  "frame",
);

export const pointerF32 = values.f32;
export const pointerU8 = values.u8;
export const pointerText = values.text;
