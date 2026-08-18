import { columnBytes, columnF32 } from "../Buffer/column-values";
import { drawPanelBox, panelFont } from "../PanelOverlay/panel-box";
import {
  COL_STREAM_SPEED_PANEL_BOX_X, COL_STREAM_SPEED_PANEL_BOX_Y,
  COL_STREAM_SPEED_PANEL_BOX_W, COL_STREAM_SPEED_PANEL_BOX_H,
  COL_STREAM_SPEED_PANEL_RECT_X, COL_STREAM_SPEED_PANEL_RECT_Y,
  COL_STREAM_SPEED_PANEL_RECT_W, COL_STREAM_SPEED_PANEL_RECT_H,
  COL_STREAM_SPEED_PANEL_SELECTED,
  COL_STREAM_SPEED_PANEL_NUM_TEXT, COL_STREAM_SPEED_PANEL_NUM_LEN,
  COL_STREAM_SPEED_PANEL_DEN_TEXT, COL_STREAM_SPEED_PANEL_DEN_LEN,
  COL_STREAM_SPEED_PANEL_TRACK_X, COL_STREAM_SPEED_PANEL_TRACK_Y,
  COL_STREAM_SPEED_PANEL_TRACK_W, COL_STREAM_SPEED_PANEL_TRACK_H,
} from "../Buffer/column-streams-gen";
import { readF32Run, readU32Run, readText, decodeAt } from "../PanelOverlay/panel-columns";

const TICK_FONT_PX = 11;
const FRAC_SCALE = 0.62;
const FRAC_GAP = 1;
const INK = "#222";
const TRACK_FILL = "#c8c8c8";
const THUMB_FILL = "#fff";
const THUMB_EDGE = "#999";
const THUMB_R = 6;

function selectedIndex(): number {
  const sel = columnBytes(COL_STREAM_SPEED_PANEL_SELECTED);
  if (!sel) return -1;
  for (let i = 0; i < sel.byteLength; i++) if (sel.getUint8(i) !== 0) return i;
  return -1;
}

export function speedPanelKey(): string {
  const x = readF32Run(COL_STREAM_SPEED_PANEL_RECT_X);
  return [
    columnF32(COL_STREAM_SPEED_PANEL_BOX_X), columnF32(COL_STREAM_SPEED_PANEL_BOX_Y),
    columnF32(COL_STREAM_SPEED_PANEL_BOX_W), columnF32(COL_STREAM_SPEED_PANEL_BOX_H),
    x ? x.length : 0, selectedIndex(),
  ].join(",");
}

export function drawSpeedPanel(c: CanvasRenderingContext2D): void {
  const x = readF32Run(COL_STREAM_SPEED_PANEL_RECT_X);
  const y = readF32Run(COL_STREAM_SPEED_PANEL_RECT_Y);
  const w = readF32Run(COL_STREAM_SPEED_PANEL_RECT_W);
  const h = readF32Run(COL_STREAM_SPEED_PANEL_RECT_H);
  const sel = columnBytes(COL_STREAM_SPEED_PANEL_SELECTED);
  const numText = readText(COL_STREAM_SPEED_PANEL_NUM_TEXT);
  const numLen = readU32Run(COL_STREAM_SPEED_PANEL_NUM_LEN);
  const denText = readText(COL_STREAM_SPEED_PANEL_DEN_TEXT);
  const denLen = readU32Run(COL_STREAM_SPEED_PANEL_DEN_LEN);
  if (!x || !y || !w || !h || !sel || !numText || !numLen || !denText || !denLen) return;

  drawPanelBox(
    c,
    columnF32(COL_STREAM_SPEED_PANEL_BOX_X),
    columnF32(COL_STREAM_SPEED_PANEL_BOX_Y),
    columnF32(COL_STREAM_SPEED_PANEL_BOX_W),
    columnF32(COL_STREAM_SPEED_PANEL_BOX_H),
  );

  const trackX = columnF32(COL_STREAM_SPEED_PANEL_TRACK_X);
  const trackY = columnF32(COL_STREAM_SPEED_PANEL_TRACK_Y);
  const trackW = columnF32(COL_STREAM_SPEED_PANEL_TRACK_W);
  const trackH = columnF32(COL_STREAM_SPEED_PANEL_TRACK_H);

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
      c.font = panelFont(TICK_FONT_PX, on ? "bold" : undefined);
      c.fillText(num, cx, y[i]!);
    } else {
      const fs = TICK_FONT_PX * FRAC_SCALE;
      c.font = panelFont(fs, on ? "bold" : undefined);
      c.fillText(num, cx, y[i]!);
      const barY = y[i]! + fs + FRAC_GAP;
      c.fillRect(cx - fs * 0.6, barY, fs * 1.2, 1);
      c.fillText(den, cx, barY + 1 + FRAC_GAP);
    }
  }
}
