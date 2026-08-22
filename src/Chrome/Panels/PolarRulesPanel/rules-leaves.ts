import { makeLeafValues } from "./leaf-values";
import { RULES_VALUE_NAMES, type RulesValueName } from "./rules-values-gen";

const values = makeLeafValues<RulesValueName>(
  "Chrome/Panels/PolarRulesPanel/paths",
  RULES_VALUE_NAMES,
);

export const rulesBytes = values.bytes;
export const rulesF32 = values.f32;
export const rulesI32 = values.i32;
export const rulesU8 = values.u8;
export const rulesF32Run = values.f32Run;
export const rulesI32Run = values.i32Run;
export const rulesU32Run = values.u32Run;
export const rulesText = values.text;
