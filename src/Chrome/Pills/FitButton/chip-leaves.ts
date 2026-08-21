import { makeLeafStore } from "../../../webview/leaf-store";
import { FIT_CHIP_VALUE_NAMES, type FitChipValueName } from "./chip-values-gen";

const store = makeLeafStore<FitChipValueName>(
  "Chrome/Pills/FitButton/paths",
  FIT_CHIP_VALUE_NAMES,
);

export const chipF32 = store.f32;
export const chipText = store.text;
