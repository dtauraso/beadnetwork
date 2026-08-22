import { makeRowLeafValues } from "../../valuefile/row-leaf-values";
import { BEAD_VALUE_NAMES, type BeadValueName } from "./bead-values-gen";

const values = makeRowLeafValues<BeadValueName>(
  "Ring/Bead/paths",
  BEAD_VALUE_NAMES,
);

export const beadBytes = values.bytes;

export const BEAD_RING_NAMES = BEAD_VALUE_NAMES.slice(4) as readonly BeadValueName[];
