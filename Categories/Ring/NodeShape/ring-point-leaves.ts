import { makeLeafValues } from "./leaf-values";
import { RING_POINT_VALUE_NAMES, type RingPointValueName } from "./ring-point-values-gen";

const values = makeLeafValues<RingPointValueName>(
  "Categories/Ring/NodeShape/paths",
  RING_POINT_VALUE_NAMES,
  "once",
);

export const nodeRingPointBytes = values.bytes;
