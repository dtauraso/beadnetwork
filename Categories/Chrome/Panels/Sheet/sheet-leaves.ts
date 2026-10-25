import { makeLeafValues } from "./leaf-values";
import { SHEET_VALUE_NAMES, type SheetValueName } from "./sheet-values-gen";

const values = makeLeafValues<SheetValueName>(
  "Categories/Chrome/Panels/Sheet/paths",
  SHEET_VALUE_NAMES,
);

export const sheetBytes = values.bytes;
export const sheetF32Run = values.f32Run;
export const sheetU32Run = values.u32Run;
export const sheetText = values.text;
