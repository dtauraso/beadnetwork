import { columnBytes, columnF32, columnU8 } from "../Buffer/column-values";
import { panelFont, roundRect } from "../PanelOverlay/panel-box";
import { drawPill, drawPopoverBox, drawHeadingText, ROW_PAD_X } from "../PanelOverlay/pill-chrome";
import { readF32Run, readU32Run, readText, decodeAt } from "../PanelOverlay/panel-columns";
import * as T from "../src/webview/three/controls/chrome-theme";
import {
  COL_STREAM_OVERLAYS_PILL_PILL_X, COL_STREAM_OVERLAYS_PILL_PILL_Y,
  COL_STREAM_OVERLAYS_PILL_PILL_W, COL_STREAM_OVERLAYS_PILL_PILL_H,
  COL_STREAM_OVERLAYS_PILL_OPEN, COL_STREAM_OVERLAYS_PILL_ACTIVE,
  COL_STREAM_OVERLAYS_PILL_POPOVER_X, COL_STREAM_OVERLAYS_PILL_POPOVER_Y,
  COL_STREAM_OVERLAYS_PILL_POPOVER_W, COL_STREAM_OVERLAYS_PILL_POPOVER_H,
  COL_STREAM_OVERLAYS_PILL_LABEL_TEXT,
  COL_STREAM_OVERLAYS_PILL_ROW_KIND, COL_STREAM_OVERLAYS_PILL_ROW_DEPTH,
  COL_STREAM_OVERLAYS_PILL_ROW_X, COL_STREAM_OVERLAYS_PILL_ROW_Y,
  COL_STREAM_OVERLAYS_PILL_ROW_W, COL_STREAM_OVERLAYS_PILL_ROW_H,
  COL_STREAM_OVERLAYS_PILL_ROW_TEXT_DATA, COL_STREAM_OVERLAYS_PILL_ROW_TEXT_LEN,
  COL_STREAM_OVERLAYS_PILL_ROW_ICON_DATA, COL_STREAM_OVERLAYS_PILL_ROW_ICON_LEN,
  COL_STREAM_OVERLAYS_PILL_ROW_ON, COL_STREAM_OVERLAYS_PILL_ROW_DISABLED,
  COL_STREAM_OVERLAYS_PILL_ROW_COUNT_ON, COL_STREAM_OVERLAYS_PILL_ROW_COUNT_ALL,
  COL_STREAM_OVERLAYS_PILL_COUNT_X, COL_STREAM_OVERLAYS_PILL_COUNT_Y,
  COL_STREAM_OVERLAYS_PILL_COUNT_W, COL_STREAM_OVERLAYS_PILL_COUNT_H,
} from "../Buffer/column-streams-gen";

const ROW_HEADING = 0;

const INDENT_HEADING = 10;
const INDENT_ROW = 14;
const CHECK_SIZE = 13;
const ICON_W = 11;
const GAP = 7;

export function overlaysPillKey(): string {
  const kinds = columnBytes(COL_STREAM_OVERLAYS_PILL_ROW_KIND);
  const on = columnBytes(COL_STREAM_OVERLAYS_PILL_ROW_ON);
  const bytes = (v: DataView | undefined) =>
    v ? new Uint8Array(v.buffer, v.byteOffset, v.byteLength).join(".") : "";
  return [
    columnF32(COL_STREAM_OVERLAYS_PILL_PILL_X), columnF32(COL_STREAM_OVERLAYS_PILL_PILL_Y),
    columnU8(COL_STREAM_OVERLAYS_PILL_OPEN), columnU8(COL_STREAM_OVERLAYS_PILL_ACTIVE),
    columnF32(COL_STREAM_OVERLAYS_PILL_POPOVER_H),
    bytes(kinds), bytes(on),
  ].join(",");
}

export function drawOverlaysPill(c: CanvasRenderingContext2D): void {
  const labelText = readText(COL_STREAM_OVERLAYS_PILL_LABEL_TEXT);
  const pw = columnF32(COL_STREAM_OVERLAYS_PILL_PILL_W);
  const ph = columnF32(COL_STREAM_OVERLAYS_PILL_PILL_H);
  if (!labelText || pw <= 0 || ph <= 0) return;
  const px = columnF32(COL_STREAM_OVERLAYS_PILL_PILL_X);
  const py = columnF32(COL_STREAM_OVERLAYS_PILL_PILL_Y);
  const open = columnU8(COL_STREAM_OVERLAYS_PILL_OPEN) !== 0;
  const active = columnU8(COL_STREAM_OVERLAYS_PILL_ACTIVE) !== 0;

  drawPill(c, px, py, pw, ph, decodeAt(labelText, 0, labelText.length), open, active);
  if (!open) return;

  drawPopoverBox(
    c,
    columnF32(COL_STREAM_OVERLAYS_PILL_POPOVER_X), columnF32(COL_STREAM_OVERLAYS_PILL_POPOVER_Y),
    columnF32(COL_STREAM_OVERLAYS_PILL_POPOVER_W), columnF32(COL_STREAM_OVERLAYS_PILL_POPOVER_H),
  );
  drawRows(c);
}

function drawCheckbox(c: CanvasRenderingContext2D, x: number, y: number, on: boolean): void {
  roundRect(c, x, y, CHECK_SIZE, CHECK_SIZE, T.RADIUS_ITEM);
  c.fillStyle = on ? T.ACCENT : "transparent";
  if (on) c.fill();
  c.strokeStyle = on ? T.ACCENT : T.TEXT;
  c.lineWidth = 1;
  c.stroke();
  if (!on) return;
  c.fillStyle = T.ACCENT_INK;
  c.font = panelFont(T.FONT_SIZE_GLYPH, 900);
  c.textAlign = "center";
  c.textBaseline = "middle";
  c.fillText("✓", x + CHECK_SIZE / 2, y + CHECK_SIZE / 2);
}

function drawRows(c: CanvasRenderingContext2D): void {
  const kinds = columnBytes(COL_STREAM_OVERLAYS_PILL_ROW_KIND);
  const depths = columnBytes(COL_STREAM_OVERLAYS_PILL_ROW_DEPTH);
  const x = readF32Run(COL_STREAM_OVERLAYS_PILL_ROW_X);
  const y = readF32Run(COL_STREAM_OVERLAYS_PILL_ROW_Y);
  const w = readF32Run(COL_STREAM_OVERLAYS_PILL_ROW_W);
  const h = readF32Run(COL_STREAM_OVERLAYS_PILL_ROW_H);
  const textData = readText(COL_STREAM_OVERLAYS_PILL_ROW_TEXT_DATA);
  const textLen = readU32Run(COL_STREAM_OVERLAYS_PILL_ROW_TEXT_LEN);
  const iconData = readText(COL_STREAM_OVERLAYS_PILL_ROW_ICON_DATA);
  const iconLen = readU32Run(COL_STREAM_OVERLAYS_PILL_ROW_ICON_LEN);
  const on = columnBytes(COL_STREAM_OVERLAYS_PILL_ROW_ON);
  const disabled = columnBytes(COL_STREAM_OVERLAYS_PILL_ROW_DISABLED);
  const countOn = readU32Run(COL_STREAM_OVERLAYS_PILL_ROW_COUNT_ON);
  const countAll = readU32Run(COL_STREAM_OVERLAYS_PILL_ROW_COUNT_ALL);
  const cx = readF32Run(COL_STREAM_OVERLAYS_PILL_COUNT_X);
  const cy = readF32Run(COL_STREAM_OVERLAYS_PILL_COUNT_Y);
  const cw = readF32Run(COL_STREAM_OVERLAYS_PILL_COUNT_W);
  const ch = readF32Run(COL_STREAM_OVERLAYS_PILL_COUNT_H);
  if (!kinds || !depths || !x || !y || !w || !h) return;
  if (!textData || !textLen || !iconData || !iconLen || !on || !disabled) return;
  if (!countOn || !countAll || !cx || !cy || !cw || !ch) return;

  let textOff = 0, iconOff = 0;
  for (let i = 0; i < x.length; i++) {
    const text = decodeAt(textData, textOff, textLen[i]!);
    textOff += textLen[i]!;
    const icon = decodeAt(iconData, iconOff, iconLen[i]!);
    iconOff += iconLen[i]!;

    const isHeading = kinds.getUint8(i) === ROW_HEADING;
    const depth = depths.getUint8(i);
    const mid = y[i]! + h[i]! / 2;
    c.globalAlpha = disabled.getUint8(i) !== 0 ? T.DISABLED_OPACITY : 1;
    c.textBaseline = "middle";
    c.textAlign = "left";

    if (isHeading) {
      const left = x[i]! + 3 + depth * INDENT_HEADING;
      c.fillStyle = T.TEXT;
      c.font = panelFont(8);
      c.fillText(icon, left, mid);
      drawHeadingText(c, text, left + GAP + 5, mid);

      const n = countOn[i]!;
      c.fillStyle = n > 0 ? T.ACCENT : T.TEXT;
      c.textAlign = "right";
      c.fillText(`${n}/${countAll[i]!}`, cx[i]! + cw[i]! - 5, cy[i]! + ch[i]! / 2);
      c.globalAlpha = 1;
      continue;
    }

    const left = x[i]! + ROW_PAD_X + depth * INDENT_ROW;
    drawCheckbox(c, left, mid - CHECK_SIZE / 2, on.getUint8(i) !== 0);
    c.fillStyle = T.TEXT;
    c.font = panelFont(T.FONT_SIZE);
    c.textAlign = "center";
    c.fillText(icon, left + CHECK_SIZE + GAP + ICON_W / 2, mid);
    c.textAlign = "left";
    c.fillText(text, left + CHECK_SIZE + GAP + ICON_W + GAP, mid);
    c.globalAlpha = 1;
  }
}
