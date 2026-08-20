import { columnF32, columnI32 } from "../../../schema/buffer-layout/column-values";
import { nodeColumn } from "../../../schema/buffer-layout/column-owners";
import {
  COL_STREAM_NODE_ROUNDS_TO_PARALLEL, COL_STREAM_NODE_MSGS_TO_PARALLEL,
} from "../../../Node/columns-gen";
import { drawBox, canvasFont, roundRect } from "../../../webview/canvas-box";
import { readF32Run, readI32Run, readU32Run, readText, decodeAt } from "../../../schema/buffer-layout/column-reads";
import {
  COL_STREAM_TILT_PANEL_BOX_X, COL_STREAM_TILT_PANEL_BOX_Y, COL_STREAM_TILT_PANEL_BOX_W,
  COL_STREAM_TILT_PANEL_BOX_H, COL_STREAM_TILT_PANEL_START_X, COL_STREAM_TILT_PANEL_START_Y,
  COL_STREAM_TILT_PANEL_START_W, COL_STREAM_TILT_PANEL_START_H,
  COL_STREAM_TILT_PANEL_RESET_X, COL_STREAM_TILT_PANEL_RESET_Y,
  COL_STREAM_TILT_PANEL_RESET_W, COL_STREAM_TILT_PANEL_RESET_H,
  COL_STREAM_TILT_PANEL_START_TEXT, COL_STREAM_TILT_PANEL_RESET_TEXT,
  COL_STREAM_TILT_PANEL_COL_NODE_ROW, COL_STREAM_TILT_PANEL_COL_LABEL_TEXT,
  COL_STREAM_TILT_PANEL_COL_LABEL_LEN, COL_STREAM_TILT_PANEL_HEAD_X,
  COL_STREAM_TILT_PANEL_HEAD_Y, COL_STREAM_TILT_PANEL_HEAD_W, COL_STREAM_TILT_PANEL_HEAD_H,
  COL_STREAM_TILT_PANEL_ROUNDS_X, COL_STREAM_TILT_PANEL_ROUNDS_Y,
  COL_STREAM_TILT_PANEL_ROUNDS_W, COL_STREAM_TILT_PANEL_ROUNDS_H,
  COL_STREAM_TILT_PANEL_MSGS_X, COL_STREAM_TILT_PANEL_MSGS_Y, COL_STREAM_TILT_PANEL_MSGS_W,
  COL_STREAM_TILT_PANEL_MSGS_H,
} from "./columns-gen";

const BTN_FILL = "#fafafa";
const BTN_EDGE = "#ddd";
const BTN_INK = "#333";
const BTN_RADIUS = 4;
const BTN_FONT_PX = 12;

const HEAD_FONT_PX = 13;
const HEAD_INK = "#222";

const CELL_FILL = "#222";
const CELL_RADIUS = 3;
const CELL_PAD_X = 7;
const KEY_FONT_PX = 11;
const KEY_INK = "#e0e0e0";
const VAL_FONT_PX = 13;
const VAL_INK = "#fff";

function drawButton(
  c: CanvasRenderingContext2D,
  x: number, y: number, w: number, h: number,
  label: string,
): void {
  if (w <= 0 || h <= 0) return;
  roundRect(c, x + 0.5, y + 0.5, w - 1, h - 1, BTN_RADIUS);
  c.fillStyle = BTN_FILL;
  c.fill();
  c.strokeStyle = BTN_EDGE;
  c.lineWidth = 1;
  c.stroke();

  c.fillStyle = BTN_INK;
  c.font = canvasFont(BTN_FONT_PX);
  c.textAlign = "center";
  c.textBaseline = "middle";
  c.fillText(label, x + w / 2, y + h / 2);
}

function drawCell(
  c: CanvasRenderingContext2D,
  x: number, y: number, w: number, h: number,
  key: string, value: number,
): void {
  if (w <= 0 || h <= 0) return;
  roundRect(c, x, y, w, h, CELL_RADIUS);
  c.fillStyle = CELL_FILL;
  c.fill();

  c.textBaseline = "middle";
  c.fillStyle = KEY_INK;
  c.font = canvasFont(KEY_FONT_PX, 600);
  c.textAlign = "left";
  c.fillText(key, x + CELL_PAD_X, y + h / 2);

  c.fillStyle = VAL_INK;
  c.font = canvasFont(VAL_FONT_PX, 700);
  c.textAlign = "right";
  c.fillText(String(value), x + w - CELL_PAD_X, y + h / 2);
}

export function tiltPanelKey(): string {
  const rows = readI32Run(COL_STREAM_TILT_PANEL_COL_NODE_ROW);
  const values: number[] = [];
  if (rows) {
    for (const row of rows) {
      values.push(
        columnI32(nodeColumn(row, COL_STREAM_NODE_ROUNDS_TO_PARALLEL)),
        columnI32(nodeColumn(row, COL_STREAM_NODE_MSGS_TO_PARALLEL)),
      );
    }
  }
  return [
    columnF32(COL_STREAM_TILT_PANEL_BOX_X), columnF32(COL_STREAM_TILT_PANEL_BOX_Y),
    columnF32(COL_STREAM_TILT_PANEL_BOX_W), columnF32(COL_STREAM_TILT_PANEL_BOX_H),
    rows ? rows.join(".") : "", values.join("."),
  ].join(",");
}

export function drawTiltPanel(c: CanvasRenderingContext2D): void {
  const rows = readI32Run(COL_STREAM_TILT_PANEL_COL_NODE_ROW);
  if (!rows || rows.length === 0) return;

  drawBox(
    c,
    columnF32(COL_STREAM_TILT_PANEL_BOX_X),
    columnF32(COL_STREAM_TILT_PANEL_BOX_Y),
    columnF32(COL_STREAM_TILT_PANEL_BOX_W),
    columnF32(COL_STREAM_TILT_PANEL_BOX_H),
  );

  const startText = readText(COL_STREAM_TILT_PANEL_START_TEXT);
  const resetText = readText(COL_STREAM_TILT_PANEL_RESET_TEXT);
  if (startText && resetText) {
    drawButton(
      c,
      columnF32(COL_STREAM_TILT_PANEL_START_X), columnF32(COL_STREAM_TILT_PANEL_START_Y),
      columnF32(COL_STREAM_TILT_PANEL_START_W), columnF32(COL_STREAM_TILT_PANEL_START_H),
      decodeAt(startText, 0, startText.length),
    );
    drawButton(
      c,
      columnF32(COL_STREAM_TILT_PANEL_RESET_X), columnF32(COL_STREAM_TILT_PANEL_RESET_Y),
      columnF32(COL_STREAM_TILT_PANEL_RESET_W), columnF32(COL_STREAM_TILT_PANEL_RESET_H),
      decodeAt(resetText, 0, resetText.length),
    );
  }

  const labelText = readText(COL_STREAM_TILT_PANEL_COL_LABEL_TEXT);
  const labelLen = readU32Run(COL_STREAM_TILT_PANEL_COL_LABEL_LEN);
  const headX = readF32Run(COL_STREAM_TILT_PANEL_HEAD_X);
  const headY = readF32Run(COL_STREAM_TILT_PANEL_HEAD_Y);
  const headW = readF32Run(COL_STREAM_TILT_PANEL_HEAD_W);
  const headH = readF32Run(COL_STREAM_TILT_PANEL_HEAD_H);
  const roundsX = readF32Run(COL_STREAM_TILT_PANEL_ROUNDS_X);
  const roundsY = readF32Run(COL_STREAM_TILT_PANEL_ROUNDS_Y);
  const roundsW = readF32Run(COL_STREAM_TILT_PANEL_ROUNDS_W);
  const roundsH = readF32Run(COL_STREAM_TILT_PANEL_ROUNDS_H);
  const msgsX = readF32Run(COL_STREAM_TILT_PANEL_MSGS_X);
  const msgsY = readF32Run(COL_STREAM_TILT_PANEL_MSGS_Y);
  const msgsW = readF32Run(COL_STREAM_TILT_PANEL_MSGS_W);
  const msgsH = readF32Run(COL_STREAM_TILT_PANEL_MSGS_H);
  if (!labelText || !labelLen || !headX || !headY || !headW || !headH) return;
  if (!roundsX || !roundsY || !roundsW || !roundsH) return;
  if (!msgsX || !msgsY || !msgsW || !msgsH) return;

  let labelOff = 0;
  for (let i = 0; i < rows.length; i++) {
    const label = decodeAt(labelText, labelOff, labelLen[i]!);
    labelOff += labelLen[i]!;
    const row = rows[i]!;

    c.fillStyle = HEAD_INK;
    c.font = canvasFont(HEAD_FONT_PX, 700);
    c.textAlign = "center";
    c.textBaseline = "middle";
    c.fillText(`node ${label}`, headX[i]! + headW[i]! / 2, headY[i]! + headH[i]! / 2);

    drawCell(
      c, roundsX[i]!, roundsY[i]!, roundsW[i]!, roundsH[i]!,
      "rounds", columnI32(nodeColumn(row, COL_STREAM_NODE_ROUNDS_TO_PARALLEL)),
    );
    drawCell(
      c, msgsX[i]!, msgsY[i]!, msgsW[i]!, msgsH[i]!,
      "msgs", columnI32(nodeColumn(row, COL_STREAM_NODE_MSGS_TO_PARALLEL)),
    );
  }
}
