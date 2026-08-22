import { makeRowLeafValues } from "./row-leaf-values";
import { INTERIOR_VALUE_NAMES, type InteriorValueName } from "./interior-values-gen";

const values = makeRowLeafValues<InteriorValueName>(
  "Categories/Node/Interior/paths",
  INTERIOR_VALUE_NAMES,
);

export const interiorBytes = values.bytes;
