import { makeLeafStore } from "../../webview/leaf-store";
import { TAB_STRIP_VALUE_NAMES, type TabStripValueName } from "./strip-values-gen";

const store = makeLeafStore<TabStripValueName>(
  "Chrome/Tabs/paths",
  TAB_STRIP_VALUE_NAMES,
);

export const stripBytes = store.bytes;
export const stripF32 = store.f32;
export const stripF32Run = store.f32Run;
export const stripU32Run = store.u32Run;
export const stripText = store.text;
