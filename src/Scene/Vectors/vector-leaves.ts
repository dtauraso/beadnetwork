import { makeRowLeafValues } from "../../valuefile/row-leaf-values";
import { VECTOR_VALUE_NAMES, type VectorValueName } from "./vector-values-gen";

const values = makeRowLeafValues<VectorValueName>(
  "Scene/Vectors/paths",
  VECTOR_VALUE_NAMES,
);

export const channelVectorBytes = values.bytes;

export const VECTOR_SHAFT_NAMES = VECTOR_VALUE_NAMES.slice(0, 16) as readonly VectorValueName[];
export const VECTOR_HEAD_NAMES = VECTOR_VALUE_NAMES.slice(16) as readonly VectorValueName[];
