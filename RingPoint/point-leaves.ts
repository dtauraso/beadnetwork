import { makeLeafValues } from "./leaf-values";
import { RING_POINT_VALUE_NAMES, type RingPointValueName } from "./point-values-gen";

const values = makeLeafValues<RingPointValueName>(
  "RingPoint/paths",
  RING_POINT_VALUE_NAMES,
);

export const ringPointBytes = values.bytes;
