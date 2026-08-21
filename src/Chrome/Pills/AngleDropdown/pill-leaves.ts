import { makeLeafStore } from "../../../webview/leaf-store";
import { ANGLE_PILL_VALUE_NAMES, type AnglePillValueName } from "./pill-values-gen";

const store = makeLeafStore<AnglePillValueName>(
  "Chrome/Pills/AngleDropdown/paths",
  ANGLE_PILL_VALUE_NAMES,
);

export const anglePillBytes = store.bytes;
export const anglePillF32 = store.f32;
export const anglePillI32 = store.i32;
export const anglePillU32 = store.u32;
export const anglePillU8 = store.u8;
export const anglePillF32Run = store.f32Run;
export const anglePillI32Run = store.i32Run;
export const anglePillU32Run = store.u32Run;
export const anglePillText = store.text;
