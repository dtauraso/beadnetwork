import { makeLeafValues } from "./leaf-values";
import { PREVIEW_VALUE_NAMES, type PreviewValueName } from "./preview-values-gen";

const values = makeLeafValues<PreviewValueName>(
  "Categories/Scene/Preview/paths",
  PREVIEW_VALUE_NAMES,
  "frame",
);

export const previewI32 = values.i32;
export const previewU8 = values.u8;
export const previewI32Run = values.i32Run;
export const previewBytes = values.bytes;
export const previewF32Run = values.f32Run;
