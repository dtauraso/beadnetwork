import { makeLeafStore } from "../../../webview/leaf-store";
import { NODES_PILL_VALUE_NAMES, type NodesPillValueName } from "./pill-values-gen";

const store = makeLeafStore<NodesPillValueName>(
  "Chrome/Pills/NodesDropdown/paths",
  NODES_PILL_VALUE_NAMES,
);

export const pillBytes = store.bytes;
export const pillF32 = store.f32;
export const pillI32 = store.i32;
export const pillU32 = store.u32;
export const pillU8 = store.u8;
export const pillF32Run = store.f32Run;
export const pillI32Run = store.i32Run;
export const pillU32Run = store.u32Run;
export const pillText = store.text;
