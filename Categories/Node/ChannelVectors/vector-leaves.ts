import { nodeBytes } from "../node-leaves";
import { VECTOR_VALUE_NAMES, type VectorValueName } from "./vector-values-gen";

export const channelVectorBytes = (row: number, name: VectorValueName) => nodeBytes(row, name);

export const VECTOR_SHAFT_NAMES = VECTOR_VALUE_NAMES.slice(0, 16) as readonly VectorValueName[];
export const VECTOR_HEAD_NAMES = VECTOR_VALUE_NAMES.slice(16) as readonly VectorValueName[];
