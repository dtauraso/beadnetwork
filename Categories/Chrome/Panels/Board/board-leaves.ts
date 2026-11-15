import { makeLeafValues } from "./leaf-values";
import { BOARD_VALUE_NAMES, type BoardValueName } from "./board-values-gen";

const values = makeLeafValues<BoardValueName>(
  "Categories/Chrome/Panels/Board/paths",
  BOARD_VALUE_NAMES,
);

export const boardBytes = values.bytes;
export const boardF32Run = values.f32Run;
export const boardU32Run = values.u32Run;
export const boardText = values.text;
