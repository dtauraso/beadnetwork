import { editUpdate } from "../Input/Codec/attr-index";

export const SLIDER_NUM_SCALE = 4;

export function encodeClockSpeed(speed: number): ArrayBuffer {
  const w = editUpdate("clock", "speed");
  w.u8(Math.round(speed * SLIDER_NUM_SCALE));
  return w.toArrayBuffer();
}
