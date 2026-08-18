import { columnBytes, columnF32, columnI32, columnU8 } from "../Buffer/column-values";
import { nodeColumn } from "../Buffer/column-owners";
import { COL_STREAM_NODE_TOP_TILT_VECTOR_IDX } from "../Buffer/column-streams-gen";
import { panelFont, roundRect } from "../PanelOverlay/panel-box";
import { readF32Run, readI32Run, readU32Run, readText, decodeAt } from "../PanelOverlay/panel-columns";
import * as T from "../src/webview/three/controls/chrome-theme";
import {
  COL_STREAM_ANGLE_PILL_PILL_X, COL_STREAM_ANGLE_PILL_PILL_Y,
  COL_STREAM_ANGLE_PILL_PILL_W, COL_STREAM_ANGLE_PILL_PILL_H,
  COL_STREAM_ANGLE_PILL_OPEN,
  COL_STREAM_ANGLE_PILL_POPOVER_X, COL_STREAM_ANGLE_PILL_POPOVER_Y,
  COL_STREAM_ANGLE_PILL_POPOVER_W, COL_STREAM_ANGLE_PILL_POPOVER_H,
  COL_STREAM_ANGLE_PILL_LABEL_TEXT,
  COL_STREAM_ANGLE_PILL_STEP_X, COL_STREAM_ANGLE_PILL_STEP_Y,
  COL_STREAM_ANGLE_PILL_STEP_W, COL_STREAM_ANGLE_PILL_STEP_H,
  COL_STREAM_ANGLE_PILL_STEP_NAME_TEXT, COL_STREAM_ANGLE_PILL_STEP_NAME_LEN,
  COL_STREAM_ANGLE_PILL_STEP_SHOWN_TEXT, COL_STREAM_ANGLE_PILL_STEP_SHOWN_LEN,
  COL_STREAM_ANGLE_PILL_STEP_VALUE_ROW, COL_STREAM_ANGLE_PILL_STEP_DENOM,
  COL_STREAM_ANGLE_PILL_STEP_UP_X, COL_STREAM_ANGLE_PILL_STEP_UP_Y,
  COL_STREAM_ANGLE_PILL_STEP_UP_W, COL_STREAM_ANGLE_PILL_STEP_UP_H,
  COL_STREAM_ANGLE_PILL_STEP_DOWN_X, COL_STREAM_ANGLE_PILL_STEP_DOWN_Y,
  COL_STREAM_ANGLE_PILL_STEP_DOWN_W, COL_STREAM_ANGLE_PILL_STEP_DOWN_H,
  COL_STREAM_ANGLE_PILL_STEP_UP_ON, COL_STREAM_ANGLE_PILL_STEP_DOWN_ON,
  COL_STREAM_ANGLE_PILL_GROUP_X, COL_STREAM_ANGLE_PILL_GROUP_Y,
  COL_STREAM_ANGLE_PILL_GROUP_W, COL_STREAM_ANGLE_PILL_GROUP_H,
  COL_STREAM_ANGLE_PILL_GROUP_OPEN,
  COL_STREAM_ANGLE_PILL_GROUP_HEAD_TEXT, COL_STREAM_ANGLE_PILL_GROUP_HEAD_LEN,
} from "../Buffer/column-streams-gen";

const CARET_W = 20;
const ROW_PAD_X = 6;
const ROW_GAP = 2;
const ARROW_PAD = 2;

function stepperValue(shown: string, valueRow: number, denom: number): string {
  if (valueRow < 0) return shown;
  const idx = columnI32(nodeColumn(valueRow, COL_STREAM_NODE_TOP_TILT_VECTOR_IDX));
  if (idx === 0) return "0";
  const sign = idx < 0 ? "-" : "";
  return `${sign}${Math.abs(idx)}π/${Math.max(1, denom)}`;
}

function drawChip(c: CanvasRenderingContext2D, x: number, y: number, w: number, h: number): void {
  roundRect(c, x + 0.5, y + 0.5, w - 1, h - 1, T.RADIUS_CHIP);
  c.fillStyle = T.CHIP;
  c.fill();
  c.strokeStyle = T.BORDER;
  c.lineWidth = 1;
  c.stroke();
}

function drawArrow(
  c: CanvasRenderingContext2D,
  x: number, y: number, w: number, h: number,
  glyph: string, enabled: boolean,
): void {
  if (w <= 0 || h <= 0) return;
  c.globalAlpha = enabled ? 1 : T.DISABLED_OPACITY;
  roundRect(c, x, y, w, h, T.RADIUS_ITEM);
  c.fillStyle = T.HOVER_ROW;
  c.fill();
  c.fillStyle = T.TEXT;
  c.font = panelFont(T.FONT_SIZE_GLYPH);
  c.textAlign = "center";
  c.textBaseline = "middle";
  c.fillText(glyph, x + w / 2, y + h / 2);
  c.globalAlpha = 1;
}

export function anglePillKey(): string {
  const open = columnU8(COL_STREAM_ANGLE_PILL_OPEN);
  const rows = readI32Run(COL_STREAM_ANGLE_PILL_STEP_VALUE_ROW);
  const denom = readI32Run(COL_STREAM_ANGLE_PILL_STEP_DENOM);
  const shownText = readText(COL_STREAM_ANGLE_PILL_STEP_SHOWN_TEXT);
  const shownLen = readU32Run(COL_STREAM_ANGLE_PILL_STEP_SHOWN_LEN);
  const values: string[] = [];
  if (rows && denom && shownText && shownLen) {
    let off = 0;
    for (let i = 0; i < rows.length; i++) {
      const shown = decodeAt(shownText, off, shownLen[i]!);
      off += shownLen[i]!;
      values.push(stepperValue(shown, rows[i]!, denom[i]!));
    }
  }
  const groupOpen = columnBytes(COL_STREAM_ANGLE_PILL_GROUP_OPEN);
  return [
    columnF32(COL_STREAM_ANGLE_PILL_PILL_X), columnF32(COL_STREAM_ANGLE_PILL_PILL_Y),
    columnF32(COL_STREAM_ANGLE_PILL_POPOVER_H), open,
    groupOpen ? new Uint8Array(groupOpen.buffer, groupOpen.byteOffset, groupOpen.byteLength).join(".") : "",
    values.join("."),
  ].join(",");
}

export function drawAnglePill(c: CanvasRenderingContext2D): void {
  const px = columnF32(COL_STREAM_ANGLE_PILL_PILL_X);
  const py = columnF32(COL_STREAM_ANGLE_PILL_PILL_Y);
  const pw = columnF32(COL_STREAM_ANGLE_PILL_PILL_W);
  const ph = columnF32(COL_STREAM_ANGLE_PILL_PILL_H);
  const labelText = readText(COL_STREAM_ANGLE_PILL_LABEL_TEXT);
  if (pw <= 0 || ph <= 0 || !labelText) return;
  const open = columnU8(COL_STREAM_ANGLE_PILL_OPEN) !== 0;

  drawChip(c, px, py, pw, ph);
  c.fillStyle = T.TEXT;
  c.font = panelFont(T.FONT_SIZE, T.FONT_WEIGHT_LABEL);
  c.textAlign = "left";
  c.textBaseline = "middle";
  c.fillText(decodeAt(labelText, 0, labelText.length), px + 9, py + ph / 2);
  c.font = panelFont(T.FONT_SIZE_GLYPH);
  c.textAlign = "center";
  c.fillText(open ? "▲" : "▼", px + pw - CARET_W / 2 - 3, py + ph / 2);

  if (!open) return;

  const box = {
    x: columnF32(COL_STREAM_ANGLE_PILL_POPOVER_X),
    y: columnF32(COL_STREAM_ANGLE_PILL_POPOVER_Y),
    w: columnF32(COL_STREAM_ANGLE_PILL_POPOVER_W),
    h: columnF32(COL_STREAM_ANGLE_PILL_POPOVER_H),
  };
  if (box.w > 0 && box.h > 0) {
    roundRect(c, box.x + 0.5, box.y + 0.5, box.w - 1, box.h - 1, T.RADIUS_PANEL);
    c.fillStyle = T.SURFACE;
    c.fill();
    c.strokeStyle = T.BORDER;
    c.lineWidth = 1;
    c.stroke();
  }

  drawGroups(c);
  drawSteppers(c);
}

function drawGroups(c: CanvasRenderingContext2D): void {
  const x = readF32Run(COL_STREAM_ANGLE_PILL_GROUP_X);
  const y = readF32Run(COL_STREAM_ANGLE_PILL_GROUP_Y);
  const h = readF32Run(COL_STREAM_ANGLE_PILL_GROUP_H);
  const openRun = columnBytes(COL_STREAM_ANGLE_PILL_GROUP_OPEN);
  const headText = readText(COL_STREAM_ANGLE_PILL_GROUP_HEAD_TEXT);
  const headLen = readU32Run(COL_STREAM_ANGLE_PILL_GROUP_HEAD_LEN);
  if (!x || !y || !h || !openRun || !headText || !headLen) return;

  let off = 0;
  for (let i = 0; i < x.length; i++) {
    const head = decodeAt(headText, off, headLen[i]!);
    off += headLen[i]!;
    const mid = y[i]! + h[i]! / 2;

    c.fillStyle = T.TEXT;
    c.textBaseline = "middle";
    c.textAlign = "left";
    c.font = panelFont(8);
    c.fillText(openRun.getUint8(i) !== 0 ? "▼" : "▶", x[i]! + ROW_PAD_X, mid);
    c.font = panelFont(T.FONT_SIZE_HEADING);
    c.fillText(head.toUpperCase(), x[i]! + ROW_PAD_X + 13, mid);
  }
}

function drawSteppers(c: CanvasRenderingContext2D): void {
  const x = readF32Run(COL_STREAM_ANGLE_PILL_STEP_X);
  const y = readF32Run(COL_STREAM_ANGLE_PILL_STEP_Y);
  const h = readF32Run(COL_STREAM_ANGLE_PILL_STEP_H);
  const nameText = readText(COL_STREAM_ANGLE_PILL_STEP_NAME_TEXT);
  const nameLen = readU32Run(COL_STREAM_ANGLE_PILL_STEP_NAME_LEN);
  const shownText = readText(COL_STREAM_ANGLE_PILL_STEP_SHOWN_TEXT);
  const shownLen = readU32Run(COL_STREAM_ANGLE_PILL_STEP_SHOWN_LEN);
  const valueRow = readI32Run(COL_STREAM_ANGLE_PILL_STEP_VALUE_ROW);
  const denom = readI32Run(COL_STREAM_ANGLE_PILL_STEP_DENOM);
  const ux = readF32Run(COL_STREAM_ANGLE_PILL_STEP_UP_X);
  const uy = readF32Run(COL_STREAM_ANGLE_PILL_STEP_UP_Y);
  const uw = readF32Run(COL_STREAM_ANGLE_PILL_STEP_UP_W);
  const uh = readF32Run(COL_STREAM_ANGLE_PILL_STEP_UP_H);
  const dx = readF32Run(COL_STREAM_ANGLE_PILL_STEP_DOWN_X);
  const dy = readF32Run(COL_STREAM_ANGLE_PILL_STEP_DOWN_Y);
  const dw = readF32Run(COL_STREAM_ANGLE_PILL_STEP_DOWN_W);
  const dh = readF32Run(COL_STREAM_ANGLE_PILL_STEP_DOWN_H);
  const upOn = columnBytes(COL_STREAM_ANGLE_PILL_STEP_UP_ON);
  const downOn = columnBytes(COL_STREAM_ANGLE_PILL_STEP_DOWN_ON);
  if (!x || !y || !h || !nameText || !nameLen || !shownText || !shownLen) return;
  if (!valueRow || !denom || !ux || !uy || !uw || !uh || !dx || !dy || !dw || !dh) return;
  if (!upOn || !downOn) return;

  let nameOff = 0;
  let shownOff = 0;
  for (let i = 0; i < x.length; i++) {
    const name = decodeAt(nameText, nameOff, nameLen[i]!);
    nameOff += nameLen[i]!;
    const shown = decodeAt(shownText, shownOff, shownLen[i]!);
    shownOff += shownLen[i]!;

    c.fillStyle = T.TEXT;
    c.font = panelFont(T.FONT_SIZE);
    c.textAlign = "left";
    c.textBaseline = "top";
    const lineH = T.FONT_SIZE * 1.2;
    c.fillText(name, x[i]! + ROW_PAD_X, y[i]! + ARROW_PAD * 2);
    c.fillText(
      stepperValue(shown, valueRow[i]!, denom[i]!),
      x[i]! + ROW_PAD_X,
      y[i]! + ARROW_PAD * 2 + lineH + ROW_GAP,
    );

    drawArrow(c, ux[i]!, uy[i]!, uw[i]!, uh[i]!, "▲", upOn.getUint8(i) !== 0);
    drawArrow(c, dx[i]!, dy[i]!, dw[i]!, dh[i]!, "▼", downOn.getUint8(i) !== 0);
  }
}
