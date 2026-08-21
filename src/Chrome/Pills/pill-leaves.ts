import { makeLeafValues } from "../../webview/leaf-values";
import { OVERLAYS_PILL_VALUE_NAMES, type OverlaysPillValueName } from "./pill-values-gen";

const values = makeLeafValues<OverlaysPillValueName>(
  "Chrome/Pills/paths",
  OVERLAYS_PILL_VALUE_NAMES,
);

export const overlaysBytes = values.bytes;
export const overlaysF32 = values.f32;
export const overlaysU8 = values.u8;
export const overlaysF32Run = values.f32Run;
export const overlaysU32Run = values.u32Run;
export const overlaysText = values.text;
