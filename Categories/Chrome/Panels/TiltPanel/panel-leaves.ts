import { makeLeafValues } from "./leaf-values";
import { TILT_PANEL_VALUE_NAMES, type TiltPanelValueName } from "./panel-values-gen";

const values = makeLeafValues<TiltPanelValueName>(
  "Categories/Chrome/Panels/TiltPanel/paths",
  TILT_PANEL_VALUE_NAMES,
);

export const tiltBytes = values.bytes;
export const tiltF32 = values.f32;
export const tiltF32Run = values.f32Run;
export const tiltI32Run = values.i32Run;
export const tiltU32Run = values.u32Run;
export const tiltText = values.text;
