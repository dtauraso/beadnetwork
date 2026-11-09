import { canvasFont, roundRect } from "../../canvas-box";
import { drawPill, drawPopoverBox, ROW_PAD_X } from "../pill";
import { decodeAt } from "../../leaf-text";
import * as T from "../../chrome-theme";
import {
  anglePillBytes, anglePillF32, anglePillU8, anglePillF32Run,
  anglePillU32Run, anglePillText,
} from "./pill-leaves";

const ROW_GAP = 2;
const ARROW_PAD = 2;

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
  const shownText = anglePillText("stepShownText");
  return [
    anglePillF32("pillX"), anglePillF32("pillY"),
    anglePillF32("popoverH"), open,
    shownText ? decodeAt(shownText, 0, shownText.length) : "",
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

  drawSteppers(c);
}

function drawSteppers(c: CanvasRenderingContext2D): void {
  const x = anglePillF32Run("stepX");
  const y = anglePillF32Run("stepY");
  const nameText = anglePillText("stepNameText");
  const nameLen = anglePillU32Run("stepNameLen");
  const shownText = anglePillText("stepShownText");
  const shownLen = anglePillU32Run("stepShownLen");
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
  if (!x || !y || !nameText || !nameLen || !shownText || !shownLen) return;
  if (!ux || !uy || !uw || !uh || !dx || !dy || !dw || !dh) return;
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
    c.fillText(shown, x[i]! + ROW_PAD_X, y[i]! + ARROW_PAD * 2 + lineH + ROW_GAP);

    drawArrow(c, ux[i]!, uy[i]!, uw[i]!, uh[i]!, "▲", upOn.getUint8(i) !== 0);
    drawArrow(c, dx[i]!, dy[i]!, dw[i]!, dh[i]!, "▼", downOn.getUint8(i) !== 0);
  }
}
