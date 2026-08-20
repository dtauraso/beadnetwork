import { columnF32, columnU8, columnU32, columnBytes } from "../../../Buffer/column-values";
import { canvasFont, roundRect } from "../../../webview/canvas-box";
import { drawPill, drawPopoverBox, ROW_PAD_X } from "../pill";
import { readF32Run, readU32Run, readText, decodeAt } from "../../../Buffer/column-reads";
import * as T from "../../../webview/canvas-theme";
import {
  COL_STREAM_NODES_PILL_PILL_X, COL_STREAM_NODES_PILL_PILL_Y, COL_STREAM_NODES_PILL_PILL_W,
  COL_STREAM_NODES_PILL_PILL_H, COL_STREAM_NODES_PILL_OPEN, COL_STREAM_NODES_PILL_POPOVER_X,
  COL_STREAM_NODES_PILL_POPOVER_Y, COL_STREAM_NODES_PILL_POPOVER_W,
  COL_STREAM_NODES_PILL_POPOVER_H, COL_STREAM_NODES_PILL_LABEL_TEXT,
  COL_STREAM_NODES_PILL_ROW_X, COL_STREAM_NODES_PILL_ROW_Y, COL_STREAM_NODES_PILL_ROW_W,
  COL_STREAM_NODES_PILL_ROW_H, COL_STREAM_NODES_PILL_ROW_OPEN,
  COL_STREAM_NODES_PILL_ROW_KIND_TEXT, COL_STREAM_NODES_PILL_ROW_KIND_LEN,
  COL_STREAM_NODES_PILL_ROW_FILL_TEXT, COL_STREAM_NODES_PILL_ROW_FILL_LEN,
  COL_STREAM_NODES_PILL_ROW_STROKE_TEXT, COL_STREAM_NODES_PILL_ROW_STROKE_LEN,
  COL_STREAM_NODES_PILL_SWATCH_X, COL_STREAM_NODES_PILL_SWATCH_Y,
  COL_STREAM_NODES_PILL_SWATCH_W, COL_STREAM_NODES_PILL_SWATCH_H,
  COL_STREAM_NODES_PILL_ROW_DESC_TEXT, COL_STREAM_NODES_PILL_ROW_DESC_LEN,
  COL_STREAM_NODES_PILL_DESC_X, COL_STREAM_NODES_PILL_DESC_Y, COL_STREAM_NODES_PILL_DESC_W,
  COL_STREAM_NODES_PILL_DESC_H, COL_STREAM_NODES_PILL_REFUSED_COUNT,
  COL_STREAM_NODES_PILL_REFUSED_X, COL_STREAM_NODES_PILL_REFUSED_Y,
  COL_STREAM_NODES_PILL_REFUSED_W, COL_STREAM_NODES_PILL_REFUSED_H,
  COL_STREAM_NODES_PILL_REFUSED_TEXT,
} from "./columns-gen";

const SWATCH_GAP = 7;
const DESC_LINE_H = 1.35;
const NOTICE_MS = 4000;

let noticeSeen = -1;
let noticeUntil = 0;

function noticeShowing(): boolean {
  const count = columnU32(COL_STREAM_NODES_PILL_REFUSED_COUNT);
  if (count !== noticeSeen) {
    if (noticeSeen >= 0) noticeUntil = performance.now() + NOTICE_MS;
    noticeSeen = count;
  }
  return performance.now() < noticeUntil;
}

export function nodesPillKey(): string {
  const rowOpen = columnBytes(COL_STREAM_NODES_PILL_ROW_OPEN);
  return [
    columnF32(COL_STREAM_NODES_PILL_PILL_X), columnF32(COL_STREAM_NODES_PILL_PILL_Y),
    columnU8(COL_STREAM_NODES_PILL_OPEN), columnF32(COL_STREAM_NODES_PILL_POPOVER_H),
    rowOpen ? new Uint8Array(rowOpen.buffer, rowOpen.byteOffset, rowOpen.byteLength).join(".") : "",
    noticeShowing() ? `n${noticeUntil}` : "",
  ].join(",");
}

export function drawNodesPill(c: CanvasRenderingContext2D): void {
  const labelText = readText(COL_STREAM_NODES_PILL_LABEL_TEXT);
  const pw = columnF32(COL_STREAM_NODES_PILL_PILL_W);
  const ph = columnF32(COL_STREAM_NODES_PILL_PILL_H);
  if (!labelText || pw <= 0 || ph <= 0) return;
  const px = columnF32(COL_STREAM_NODES_PILL_PILL_X);
  const py = columnF32(COL_STREAM_NODES_PILL_PILL_Y);
  const open = columnU8(COL_STREAM_NODES_PILL_OPEN) !== 0;

  drawPill(c, px, py, pw, ph, decodeAt(labelText, 0, labelText.length), open);

  if (open) {
    drawPopoverBox(
      c,
      columnF32(COL_STREAM_NODES_PILL_POPOVER_X), columnF32(COL_STREAM_NODES_PILL_POPOVER_Y),
      columnF32(COL_STREAM_NODES_PILL_POPOVER_W), columnF32(COL_STREAM_NODES_PILL_POPOVER_H),
    );
    drawRows(c);
  }

  if (noticeShowing()) drawNotice(c);
}

function drawNotice(c: CanvasRenderingContext2D): void {
  const text = readText(COL_STREAM_NODES_PILL_REFUSED_TEXT);
  const w = columnF32(COL_STREAM_NODES_PILL_REFUSED_W);
  const h = columnF32(COL_STREAM_NODES_PILL_REFUSED_H);
  if (!text || w <= 0 || h <= 0) return;
  const x = columnF32(COL_STREAM_NODES_PILL_REFUSED_X);
  const y = columnF32(COL_STREAM_NODES_PILL_REFUSED_Y);

  roundRect(c, x + 0.5, y + 0.5, w - 1, h - 1, T.RADIUS_CHIP);
  c.fillStyle = T.CHIP;
  c.fill();
  c.strokeStyle = T.BORDER;
  c.lineWidth = 1;
  c.stroke();

  c.fillStyle = T.TEXT;
  c.font = canvasFont(T.FONT_SIZE);
  c.textAlign = "center";
  c.textBaseline = "middle";
  c.fillText(decodeAt(text, 0, text.length), x + w / 2, y + h / 2);
}

function wrap(c: CanvasRenderingContext2D, text: string, w: number): string[] {
  const out: string[] = [];
  let line = "";
  for (const word of text.split(" ")) {
    const next = line ? `${line} ${word}` : word;
    if (line && c.measureText(next).width > w) {
      out.push(line);
      line = word;
    } else {
      line = next;
    }
  }
  if (line) out.push(line);
  return out;
}

function drawRows(c: CanvasRenderingContext2D): void {
  const x = readF32Run(COL_STREAM_NODES_PILL_ROW_X);
  const y = readF32Run(COL_STREAM_NODES_PILL_ROW_Y);
  const h = readF32Run(COL_STREAM_NODES_PILL_ROW_H);
  const openRun = columnBytes(COL_STREAM_NODES_PILL_ROW_OPEN);
  const kindText = readText(COL_STREAM_NODES_PILL_ROW_KIND_TEXT);
  const kindLen = readU32Run(COL_STREAM_NODES_PILL_ROW_KIND_LEN);
  const fillText = readText(COL_STREAM_NODES_PILL_ROW_FILL_TEXT);
  const fillLen = readU32Run(COL_STREAM_NODES_PILL_ROW_FILL_LEN);
  const strokeText = readText(COL_STREAM_NODES_PILL_ROW_STROKE_TEXT);
  const strokeLen = readU32Run(COL_STREAM_NODES_PILL_ROW_STROKE_LEN);
  const sx = readF32Run(COL_STREAM_NODES_PILL_SWATCH_X);
  const sy = readF32Run(COL_STREAM_NODES_PILL_SWATCH_Y);
  const sw = readF32Run(COL_STREAM_NODES_PILL_SWATCH_W);
  const sh = readF32Run(COL_STREAM_NODES_PILL_SWATCH_H);
  const descText = readText(COL_STREAM_NODES_PILL_ROW_DESC_TEXT);
  const descLen = readU32Run(COL_STREAM_NODES_PILL_ROW_DESC_LEN);
  const dx = readF32Run(COL_STREAM_NODES_PILL_DESC_X);
  const dy = readF32Run(COL_STREAM_NODES_PILL_DESC_Y);
  const dw = readF32Run(COL_STREAM_NODES_PILL_DESC_W);
  if (!x || !y || !h || !openRun || !kindText || !kindLen) return;
  if (!fillText || !fillLen || !strokeText || !strokeLen) return;
  if (!sx || !sy || !sw || !sh || !descText || !descLen || !dx || !dy || !dw) return;

  let kindOff = 0, fillOff = 0, strokeOff = 0, descOff = 0;
  for (let i = 0; i < x.length; i++) {
    const kind = decodeAt(kindText, kindOff, kindLen[i]!);
    kindOff += kindLen[i]!;
    const fill = decodeAt(fillText, fillOff, fillLen[i]!);
    fillOff += fillLen[i]!;
    const stroke = decodeAt(strokeText, strokeOff, strokeLen[i]!);
    strokeOff += strokeLen[i]!;
    const desc = decodeAt(descText, descOff, descLen[i]!);
    descOff += descLen[i]!;

    const mid = y[i]! + h[i]! / 2;
    c.fillStyle = T.TEXT;
    c.textBaseline = "middle";
    c.textAlign = "left";
    c.font = canvasFont(8);
    c.fillText(openRun.getUint8(i) !== 0 ? "▼" : "▶", x[i]! + ROW_PAD_X, mid);

    roundRect(c, sx[i]!, sy[i]!, sw[i]!, sh[i]!, T.RADIUS_ITEM);
    c.fillStyle = fill || "#888";
    c.fill();
    c.strokeStyle = stroke || "#888";
    c.lineWidth = 1;
    c.stroke();

    c.fillStyle = T.TEXT;
    c.font = canvasFont(T.FONT_SIZE);
    c.fillText(kind, sx[i]! + sw[i]! + SWATCH_GAP, mid);

    if (desc === "") continue;
    c.globalAlpha = 0.8;
    c.textBaseline = "top";
    const lineH = T.FONT_SIZE * DESC_LINE_H;
    wrap(c, desc, dw[i]!).forEach((line, n) => {
      c.fillText(line, dx[i]!, dy[i]! + n * lineH);
    });
    c.globalAlpha = 1;
  }
}
