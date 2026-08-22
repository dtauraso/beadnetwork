import { makeLeafValues } from "../../../valuefile/leaf-values";
import { SLIDER_PANEL_VALUE_NAMES, type SliderPanelValueName } from "./panel-values-gen";

const values = makeLeafValues<SliderPanelValueName>(
  "Chrome/Panels/SliderPanel/paths",
  SLIDER_PANEL_VALUE_NAMES,
  "frame",
);

export const sliderBytes = values.bytes;
export const sliderF32 = values.f32;
export const sliderF32Run = values.f32Run;
export const sliderU32Run = values.u32Run;
export const sliderText = values.text;
