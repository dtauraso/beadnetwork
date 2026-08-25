import { makeRowLeafValues } from "./row-leaf-values";
import { TOP_VECTOR_VALUE_NAMES, type TopVectorValueName } from "./top-vector-values-gen";

const values = makeRowLeafValues<TopVectorValueName>(
  "Categories/Node/TopVector/paths",
  TOP_VECTOR_VALUE_NAMES,
);

export const topVectorBytes = values.bytes;

export const TOP_VECTOR_SHAFT_NAMES = TOP_VECTOR_VALUE_NAMES.slice(1, 17) as readonly TopVectorValueName[];
export const TOP_VECTOR_HEAD_NAMES = TOP_VECTOR_VALUE_NAMES.slice(17) as readonly TopVectorValueName[];
