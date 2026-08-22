import { nodeI32 } from "../../../Node/node-leaves";
import { drawBox, canvasFont, roundRect } from "../../canvas-box";
import { decodeAt } from "../../leaf-text";
import {
  tiltF32, tiltF32Run, tiltI32Run, tiltU32Run, tiltText,
} from "./panel-leaves";

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
  const rows = tiltI32Run("colNodeRow");
  const values: number[] = [];
  if (rows) {
    for (const row of rows) {
      values.push(
        nodeI32(row, "roundsToParallel"),
        nodeI32(row, "msgsToParallel"),
      );
    }
  }
  return [
    tiltF32("boxX"), tiltF32("boxY"),
    tiltF32("boxW"), tiltF32("boxH"),
    rows ? rows.join(".") : "", values.join("."),
  ].join(",");
}

export function drawTiltPanel(c: CanvasRenderingContext2D): void {
  const rows = tiltI32Run("colNodeRow");
  if (!rows || rows.length === 0) return;

  drawBox(
    c,
    tiltF32("boxX"),
    tiltF32("boxY"),
    tiltF32("boxW"),
    tiltF32("boxH"),
  );

  const startText = tiltText("startText");
  const resetText = tiltText("resetText");
  if (startText && resetText) {
    drawButton(
      c,
      tiltF32("startX"), tiltF32("startY"),
      tiltF32("startW"), tiltF32("startH"),
      decodeAt(startText, 0, startText.length),
    );
    drawButton(
      c,
      tiltF32("resetX"), tiltF32("resetY"),
      tiltF32("resetW"), tiltF32("resetH"),
      decodeAt(resetText, 0, resetText.length),
    );
  }

  const labelText = tiltText("colLabelText");
  const labelLen = tiltU32Run("colLabelLen");
  const headX = tiltF32Run("headX");
  const headY = tiltF32Run("headY");
  const headW = tiltF32Run("headW");
  const headH = tiltF32Run("headH");
  const roundsX = tiltF32Run("roundsX");
  const roundsY = tiltF32Run("roundsY");
  const roundsW = tiltF32Run("roundsW");
  const roundsH = tiltF32Run("roundsH");
  const msgsX = tiltF32Run("msgsX");
  const msgsY = tiltF32Run("msgsY");
  const msgsW = tiltF32Run("msgsW");
  const msgsH = tiltF32Run("msgsH");
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
      "rounds", nodeI32(row, "roundsToParallel"),
    );
    drawCell(
      c, msgsX[i]!, msgsY[i]!, msgsW[i]!, msgsH[i]!,
      "msgs", nodeI32(row, "msgsToParallel"),
    );
  }
}

