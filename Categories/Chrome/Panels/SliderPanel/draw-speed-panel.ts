import { drawBox, canvasFont, roundRect } from "../../canvas-box";
import { decodeAt } from "../../leaf-text";
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
const STEP_FILL = "#f2f2f2";
const STEP_EDGE = "#bbb";
const STEP_RADIUS = 4;

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
    sliderF32("stepX"), sliderF32("stepW"),
    sliderF32("resetX"), sliderF32("resetW"),
    sliderF32("startX"), sliderF32("startW"),
    x ? x.length : 0, selectedIndex(),
  ].join(",");
}

const STEP_BUTTON = { x: "stepX", y: "stepY", w: "stepW", h: "stepH", text: "stepText" } as const;
const RESET_BUTTON = { x: "resetX", y: "resetY", w: "resetW", h: "resetH", text: "resetText" } as const;
const START_BUTTON = { x: "startX", y: "startY", w: "startW", h: "startH", text: "startText" } as const;

function drawButton(
  c: CanvasRenderingContext2D,
  b: typeof STEP_BUTTON | typeof RESET_BUTTON | typeof START_BUTTON,
): void {
  const w = sliderF32(b.w);
  const h = sliderF32(b.h);
  const label = sliderText(b.text);
  if (w <= 0 || h <= 0 || !label) return;
  const x = sliderF32(b.x);
  const y = sliderF32(b.y);

  roundRect(c, x + 0.5, y + 0.5, w - 1, h - 1, STEP_RADIUS);
  c.fillStyle = STEP_FILL;
  c.fill();
  c.strokeStyle = STEP_EDGE;
  c.lineWidth = 1;
  c.stroke();

  c.fillStyle = INK;
  c.font = canvasFont(TICK_FONT_PX);
  c.textAlign = "center";
  c.textBaseline = "middle";
  c.fillText(decodeAt(label, 0, label.length), x + w / 2, y + h / 2);
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

  drawButton(c, STEP_BUTTON);
  drawButton(c, RESET_BUTTON);
  drawButton(c, START_BUTTON);

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
