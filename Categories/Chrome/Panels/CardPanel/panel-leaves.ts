import { makeLeafValues } from "./leaf-values";
import { CARD_PANEL_VALUE_NAMES, type CardPanelValueName } from "./panel-values-gen";

const values = makeLeafValues<CardPanelValueName>(
  "Categories/Chrome/Panels/CardPanel/paths",
  CARD_PANEL_VALUE_NAMES,
);

export const cardBytes = values.bytes;
export const cardF32 = values.f32;
export const cardU8 = values.u8;
export const cardF32Run = values.f32Run;
export const cardI32Run = values.i32Run;
export const cardU32Run = values.u32Run;
export const cardText = values.text;
