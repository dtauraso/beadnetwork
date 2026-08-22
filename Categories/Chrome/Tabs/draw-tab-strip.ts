import { canvasFont, roundRect } from "../canvas-box";
import { decodeAt } from "../leaf-text";
import * as T from "../chrome-theme";
import { stripBytes, stripF32, stripF32Run, stripU32Run, stripText } from "./strip-leaves";

export function tabStripKey(): string {
  const sel = stripBytes("tabSelected");
  return [
    stripF32("stripX"), stripF32("stripW"),
    sel ? new Uint8Array(sel.buffer, sel.byteOffset, sel.byteLength).join(".") : "",
  ].join(",");
}

export function drawTabStrip(c: CanvasRenderingContext2D): void {
  const sw = stripF32("stripW");
  const sh = stripF32("stripH");
  if (sw <= 0 || sh <= 0) return;
  const sx = stripF32("stripX");
  const sy = stripF32("stripY");

  roundRect(c, sx + 0.5, sy + 0.5, sw - 1, sh - 1, T.RADIUS_CHIP);
  c.fillStyle = T.CHIP;
  c.fill();
  c.strokeStyle = T.BORDER;
  c.lineWidth = 1;
  c.stroke();

  const x = stripF32Run("tabX");
  const y = stripF32Run("tabY");
  const w = stripF32Run("tabW");
  const h = stripF32Run("tabH");
  const nameText = stripText("tabNameText");
  const nameLen = stripU32Run("tabNameLen");
  const sel = stripBytes("tabSelected");
  if (!x || !y || !w || !h || !nameText || !nameLen || !sel) return;

  let off = 0;
  for (let i = 0; i < x.length; i++) {
    const name = decodeAt(nameText, off, nameLen[i]!);
    off += nameLen[i]!;
    const on = sel.getUint8(i) !== 0;

    if (on) {
      roundRect(c, x[i]!, y[i]!, w[i]!, h[i]!, T.RADIUS_ITEM);
      c.fillStyle = T.ACCENT;
      c.fill();
    }
    c.fillStyle = on ? T.ACCENT_INK : T.TEXT;
    c.font = canvasFont(T.FONT_SIZE);
    c.textAlign = "center";
    c.textBaseline = "middle";
    c.fillText(name, x[i]! + w[i]! / 2, y[i]! + h[i]! / 2);
  }
}
