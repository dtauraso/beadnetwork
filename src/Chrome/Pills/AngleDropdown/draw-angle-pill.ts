import { columnI32 } from "../../../Buffer/column-values";
import { nodeColumn } from "../../../Buffer/column-owners";
import {
  COL_STREAM_NODE_TOP_TILT_VECTOR_IDX,
} from "../../../Node/columns-gen";
import { canvasFont, roundRect } from "../../../webview/canvas-box";
import { drawPill, drawPopoverBox, drawHeadingText, ROW_PAD_X } from "../pill";
import { decodeAt } from "../../../Buffer/column-reads";
import * as T from "../../../webview/canvas-theme";
import {
  anglePillBytes, anglePillF32, anglePillU8, anglePillF32Run, anglePillI32Run,
  anglePillU32Run, anglePillText,
} from "./pill-leaves";

const ROW_GAP = 2;
const ARROW_PAD = 2;

function stepperValue(shown: string, valueRow: number, denom: number): string {
  if (valueRow < 0) return shown;
  const idx = columnI32(nodeColumn(valueRow, COL_STREAM_NODE_TOP_TILT_VECTOR_IDX));
  if (idx === 0) return "0";
  const sign = idx < 0 ? "-" : "";
  return `${sign}${Math.abs(idx)}π/${Math.max(1, denom)}`;
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
  c.font = canvasFont(T.FONT_SIZE_GLYPH);
  c.textAlign = "center";
  c.textBaseline = "middle";
  c.fillText(glyph, x + w / 2, y + h / 2);
  c.globalAlpha = 1;
}

export function anglePillKey(): string {
  const open = anglePillU8("open");
  const rows = anglePillI32Run("stepValueRow");
  const denom = anglePillI32Run("stepDenom");
  const shownText = anglePillText("stepShownText");
  const shownLen = anglePillU32Run("stepShownLen");
  const values: string[] = [];
  if (rows && denom && shownText && shownLen) {
    let off = 0;
    for (let i = 0; i < rows.length; i++) {
      const shown = decodeAt(shownText, off, shownLen[i]!);
      off += shownLen[i]!;
      values.push(stepperValue(shown, rows[i]!, denom[i]!));
    }
  }
  const groupOpen = anglePillBytes("groupOpen");
  return [
    anglePillF32("pillX"), anglePillF32("pillY"),
    anglePillF32("popoverH"), open,
    groupOpen ? new Uint8Array(groupOpen.buffer, groupOpen.byteOffset, groupOpen.byteLength).join(".") : "",
    values.join("."),
  ].join(",");
}

export function drawAnglePill(c: CanvasRenderingContext2D): void {
  const px = anglePillF32("pillX");
  const py = anglePillF32("pillY");
  const pw = anglePillF32("pillW");
  const ph = anglePillF32("pillH");
  const labelText = anglePillText("labelText");
  if (pw <= 0 || ph <= 0 || !labelText) return;
  const open = anglePillU8("open") !== 0;

  drawPill(c, px, py, pw, ph, decodeAt(labelText, 0, labelText.length), open);

  if (!open) return;

  drawPopoverBox(
    c,
    anglePillF32("popoverX"), anglePillF32("popoverY"),
    anglePillF32("popoverW"), anglePillF32("popoverH"),
  );

  drawGroups(c);
  drawSteppers(c);
}

function drawGroups(c: CanvasRenderingContext2D): void {
  const x = anglePillF32Run("groupX");
  const y = anglePillF32Run("groupY");
  const h = anglePillF32Run("groupH");
  const openRun = anglePillBytes("groupOpen");
  const headText = anglePillText("groupHeadText");
  const headLen = anglePillU32Run("groupHeadLen");
  if (!x || !y || !h || !openRun || !headText || !headLen) return;

  let off = 0;
  for (let i = 0; i < x.length; i++) {
    const head = decodeAt(headText, off, headLen[i]!);
    off += headLen[i]!;
    const mid = y[i]! + h[i]! / 2;

    c.fillStyle = T.TEXT;
    c.textBaseline = "middle";
    c.textAlign = "left";
    c.font = canvasFont(8);
    c.fillText(openRun.getUint8(i) !== 0 ? "▼" : "▶", x[i]! + ROW_PAD_X, mid);
    drawHeadingText(c, head, x[i]! + ROW_PAD_X + 13, mid);
  }
}

function drawSteppers(c: CanvasRenderingContext2D): void {
  const x = anglePillF32Run("stepX");
  const y = anglePillF32Run("stepY");
  const h = anglePillF32Run("stepH");
  const nameText = anglePillText("stepNameText");
  const nameLen = anglePillU32Run("stepNameLen");
  const shownText = anglePillText("stepShownText");
  const shownLen = anglePillU32Run("stepShownLen");
  const valueRow = anglePillI32Run("stepValueRow");
  const denom = anglePillI32Run("stepDenom");
  const ux = anglePillF32Run("stepUpX");
  const uy = anglePillF32Run("stepUpY");
  const uw = anglePillF32Run("stepUpW");
  const uh = anglePillF32Run("stepUpH");
  const dx = anglePillF32Run("stepDownX");
  const dy = anglePillF32Run("stepDownY");
  const dw = anglePillF32Run("stepDownW");
  const dh = anglePillF32Run("stepDownH");
  const upOn = anglePillBytes("stepUpOn");
  const downOn = anglePillBytes("stepDownOn");
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
    c.font = canvasFont(T.FONT_SIZE);
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
