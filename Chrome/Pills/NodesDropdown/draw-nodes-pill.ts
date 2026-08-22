import { canvasFont, roundRect } from "../../canvas-box";
import { drawPill, drawPopoverBox, ROW_PAD_X } from "../pill";
import { decodeAt } from "../../leaf-text";
import * as T from "../../chrome-theme";
import {
  pillBytes, pillF32, pillU8, pillU32, pillU32Run, pillF32Run, pillText,
} from "./pill-leaves";

const SWATCH_GAP = 7;
const DESC_LINE_H = 1.35;
const NOTICE_MS = 4000;

let noticeSeen = -1;
let noticeUntil = 0;

function noticeShowing(): boolean {
  const count = pillU32("refusedCount");
  if (count !== noticeSeen) {
    if (noticeSeen >= 0) noticeUntil = performance.now() + NOTICE_MS;
    noticeSeen = count;
  }
  return performance.now() < noticeUntil;
}

export function nodesPillKey(): string {
  const rowOpen = pillBytes("rowOpen");
  return [
    pillF32("pillX"), pillF32("pillY"),
    pillU8("open"), pillF32("popoverH"),
    rowOpen ? new Uint8Array(rowOpen.buffer, rowOpen.byteOffset, rowOpen.byteLength).join(".") : "",
    noticeShowing() ? `n${noticeUntil}` : "",
  ].join(",");
}

export function drawNodesPill(c: CanvasRenderingContext2D): void {
  const labelText = pillText("labelText");
  const pw = pillF32("pillW");
  const ph = pillF32("pillH");
  if (!labelText || pw <= 0 || ph <= 0) return;
  const px = pillF32("pillX");
  const py = pillF32("pillY");
  const open = pillU8("open") !== 0;

  drawPill(c, px, py, pw, ph, decodeAt(labelText, 0, labelText.length), open);

  if (open) {
    drawPopoverBox(
      c,
      pillF32("popoverX"), pillF32("popoverY"),
      pillF32("popoverW"), pillF32("popoverH"),
    );
    drawRows(c);
  }

  if (noticeShowing()) drawNotice(c);
}

function drawNotice(c: CanvasRenderingContext2D): void {
  const text = pillText("refusedText");
  const w = pillF32("refusedW");
  const h = pillF32("refusedH");
  if (!text || w <= 0 || h <= 0) return;
  const x = pillF32("refusedX");
  const y = pillF32("refusedY");

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
  const x = pillF32Run("rowX");
  const y = pillF32Run("rowY");
  const h = pillF32Run("rowH");
  const openRun = pillBytes("rowOpen");
  const kindText = pillText("rowKindText");
  const kindLen = pillU32Run("rowKindLen");
  const fillText = pillText("rowFillText");
  const fillLen = pillU32Run("rowFillLen");
  const strokeText = pillText("rowStrokeText");
  const strokeLen = pillU32Run("rowStrokeLen");
  const sx = pillF32Run("swatchX");
  const sy = pillF32Run("swatchY");
  const sw = pillF32Run("swatchW");
  const sh = pillF32Run("swatchH");
  const descText = pillText("rowDescText");
  const descLen = pillU32Run("rowDescLen");
  const dx = pillF32Run("descX");
  const dy = pillF32Run("descY");
  const dw = pillF32Run("descW");
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
