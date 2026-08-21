import { makeLeafStore } from "../../../webview/leaf-store";
import { RULES_VALUE_NAMES, type RulesValueName } from "./rules-values-gen";

const store = makeLeafStore<RulesValueName>(
  "Chrome/Panels/PolarRulesPanel/paths",
  RULES_VALUE_NAMES,
);

export const rulesBytes = store.bytes;
export const rulesF32 = store.f32;
export const rulesI32 = store.i32;
export const rulesU8 = store.u8;
export const rulesF32Run = store.f32Run;
export const rulesI32Run = store.i32Run;
export const rulesU32Run = store.u32Run;
export const rulesText = store.text;
