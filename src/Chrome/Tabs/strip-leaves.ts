import { makeLeafValues } from "../../valuefile/leaf-values";
import { TAB_STRIP_VALUE_NAMES, type TabStripValueName } from "./strip-values-gen";

const values = makeLeafValues<TabStripValueName>(
  "Chrome/Tabs/paths",
  TAB_STRIP_VALUE_NAMES,
);

export const stripBytes = values.bytes;
export const stripF32 = values.f32;
export const stripF32Run = values.f32Run;
export const stripU32Run = values.u32Run;
export const stripText = values.text;
