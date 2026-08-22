import { canvasFont, roundRect } from "../canvas-box";
import { drawPill, drawPopoverBox, drawHeadingText, ROW_PAD_X } from "./pill";
import { decodeAt } from "../leaf-text";
import * as T from "../chrome-theme";
import {
  overlaysBytes, overlaysF32, overlaysU8, overlaysF32Run, overlaysU32Run, overlaysText,
} from "./pill-leaves";

const ROW_HEADING = 0;

const INDENT_HEADING = 10;
const INDENT_ROW = 14;
const CHECK_SIZE = 13;
const ICON_W = 11;
const GAP = 7;

export function overlaysPillKey(): string {
  const kinds = overlaysBytes("rowKind");
  const on = overlaysBytes("rowOn");
  const bytes = (v: DataView | undefined) =>
    v ? new Uint8Array(v.buffer, v.byteOffset, v.byteLength).join(".") : "";
  return [
    overlaysF32("pillX"), overlaysF32("pillY"),
    overlaysU8("open"), overlaysU8("active"),
    overlaysF32("popoverH"),
    overlaysF32("scrollY"),
    bytes(kinds), bytes(on),
  ].join(",");
}

export function drawOverlaysPill(c: CanvasRenderingContext2D): void {
  const labelText = overlaysText("labelText");
  const pw = overlaysF32("pillW");
  const ph = overlaysF32("pillH");
  if (!labelText || pw <= 0 || ph <= 0) return;
  const px = overlaysF32("pillX");
  const py = overlaysF32("pillY");
  const open = overlaysU8("open") !== 0;
  const active = overlaysU8("active") !== 0;

  drawPill(c, px, py, pw, ph, decodeAt(labelText, 0, labelText.length), open, active);
  if (!open) return;

  const bx = overlaysF32("popoverX");
  const by = overlaysF32("popoverY");
  const bw = overlaysF32("popoverW");
  const bh = overlaysF32("popoverH");
  drawPopoverBox(c, bx, by, bw, bh);

  c.save();
  roundRect(c, bx, by, bw, bh, T.RADIUS_PANEL);
  c.clip();
  drawRows(c);
  c.restore();
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
  c.font = canvasFont(T.FONT_SIZE_GLYPH, 900);
  c.textAlign = "center";
  c.textBaseline = "middle";
  c.fillText("✓", x + CHECK_SIZE / 2, y + CHECK_SIZE / 2);
}

function drawRows(c: CanvasRenderingContext2D): void {
  const kinds = overlaysBytes("rowKind");
  const depths = overlaysBytes("rowDepth");
  const x = overlaysF32Run("rowX");
  const y = overlaysF32Run("rowY");
  const w = overlaysF32Run("rowW");
  const h = overlaysF32Run("rowH");
  const textData = overlaysText("rowTextData");
  const textLen = overlaysU32Run("rowTextLen");
  const iconData = overlaysText("rowIconData");
  const iconLen = overlaysU32Run("rowIconLen");
  const on = overlaysBytes("rowOn");
  const disabled = overlaysBytes("rowDisabled");
  const countOn = overlaysU32Run("rowCountOn");
  const countAll = overlaysU32Run("rowCountAll");
  const cx = overlaysF32Run("countX");
  const cy = overlaysF32Run("countY");
  const cw = overlaysF32Run("countW");
  const ch = overlaysF32Run("countH");
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
      c.font = canvasFont(8);
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
    c.font = canvasFont(T.FONT_SIZE);
    c.textAlign = "center";
    c.fillText(icon, left + CHECK_SIZE + GAP + ICON_W / 2, mid);
    c.textAlign = "left";
    c.fillText(text, left + CHECK_SIZE + GAP + ICON_W + GAP, mid);
    c.globalAlpha = 1;
  }
}

