import { drawBox, canvasFont } from "../../../webview/canvas-box";
import { decodeAt } from "../../../Buffer/column-reads";
import {
  sliderBytes, sliderF32, sliderF32Run, sliderU32Run, sliderText,
} from "./panel-leaves";

const TICK_FONT_PX = 11;
const FRAC_SCALE = 0.62;
const FRAC_GAP = 1;
const INK = "#222";
const TRACK_FILL = "#c8c8c8";
const THUMB_FILL = "#fff";
const THUMB_EDGE = "#999";
const THUMB_R = 6;

function selectedIndex(): number {
  const sel = sliderBytes("selected");
  if (!sel) return -1;
  for (let i = 0; i < sel.byteLength; i++) if (sel.getUint8(i) !== 0) return i;
  return -1;
}

export function speedPanelKey(): string {
  const x = sliderF32Run("rectX");
  return [
    sliderF32("boxX"), sliderF32("boxY"),
    sliderF32("boxW"), sliderF32("boxH"),
    x ? x.length : 0, selectedIndex(),
  ].join(",");
}

export function drawSpeedPanel(c: CanvasRenderingContext2D): void {
  const x = sliderF32Run("rectX");
  const y = sliderF32Run("rectY");
  const w = sliderF32Run("rectW");
  const h = sliderF32Run("rectH");
  const sel = sliderBytes("selected");
  const numText = sliderText("numText");
  const numLen = sliderU32Run("numLen");
  const denText = sliderText("denText");
  const denLen = sliderU32Run("denLen");
  if (!x || !y || !w || !h || !sel || !numText || !numLen || !denText || !denLen) return;

  drawBox(
    c,
    sliderF32("boxX"),
    sliderF32("boxY"),
    sliderF32("boxW"),
    sliderF32("boxH"),
  );

  const trackX = sliderF32("trackX");
  const trackY = sliderF32("trackY");
  const trackW = sliderF32("trackW");
  const trackH = sliderF32("trackH");

  c.fillStyle = TRACK_FILL;
  c.fillRect(trackX, trackY, trackW, trackH);

  let numOff = 0;
  let denOff = 0;
  for (let i = 0; i < x.length; i++) {
    const num = decodeAt(numText, numOff, numLen[i]!);
    numOff += numLen[i]!;
    const den = decodeAt(denText, denOff, denLen[i]!);
    denOff += denLen[i]!;
    const on = sel.getUint8(i) !== 0;
    const cx = x[i]! + w[i]! / 2;

    if (on) {
      c.fillStyle = THUMB_FILL;
      c.strokeStyle = THUMB_EDGE;
      c.lineWidth = 1;
      c.beginPath();
      c.arc(cx, trackY + trackH / 2, THUMB_R, 0, Math.PI * 2);
      c.fill();
      c.stroke();
    }

    c.fillStyle = INK;
    c.textAlign = "center";
    c.textBaseline = "top";
    if (den === "") {
      c.font = canvasFont(TICK_FONT_PX, on ? "bold" : undefined);
      c.fillText(num, cx, y[i]!);
    } else {
      const fs = TICK_FONT_PX * FRAC_SCALE;
      c.font = canvasFont(fs, on ? "bold" : undefined);
      c.fillText(num, cx, y[i]!);
      const barY = y[i]! + fs + FRAC_GAP;
      c.fillRect(cx - fs * 0.6, barY, fs * 1.2, 1);
      c.fillText(den, cx, barY + 1 + FRAC_GAP);
    }
  }
}
