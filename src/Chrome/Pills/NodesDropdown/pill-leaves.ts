import { makeLeafValues } from "../../../valuefile/leaf-values";
import { NODES_PILL_VALUE_NAMES, type NodesPillValueName } from "./pill-values-gen";

const values = makeLeafValues<NodesPillValueName>(
  "Chrome/Pills/NodesDropdown/paths",
  NODES_PILL_VALUE_NAMES,
);

export const pillBytes = values.bytes;
export const pillF32 = values.f32;
export const pillI32 = values.i32;
export const pillU32 = values.u32;
export const pillU8 = values.u8;
export const pillF32Run = values.f32Run;
export const pillI32Run = values.i32Run;
export const pillU32Run = values.u32Run;
export const pillText = values.text;
