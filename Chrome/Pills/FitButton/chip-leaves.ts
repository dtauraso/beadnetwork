import { makeLeafValues } from "./leaf-values";
import { FIT_CHIP_VALUE_NAMES, type FitChipValueName } from "./chip-values-gen";

const values = makeLeafValues<FitChipValueName>(
  "Chrome/Pills/FitButton/paths",
  FIT_CHIP_VALUE_NAMES,
);

export const chipF32 = values.f32;
export const chipText = values.text;
