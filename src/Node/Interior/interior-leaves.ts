import { makeRowLeafValues } from "../../webview/row-leaf-values";
import { INTERIOR_VALUE_NAMES, type InteriorValueName } from "./interior-values-gen";

const values = makeRowLeafValues<InteriorValueName>(
  "Node/Interior/paths",
  INTERIOR_VALUE_NAMES,
);

export const interiorBytes = values.bytes;
