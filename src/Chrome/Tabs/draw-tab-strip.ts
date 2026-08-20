import { columnBytes, columnF32 } from "../../schema/buffer-layout/column-values";
import { panelFont, roundRect } from "../PanelOverlay/panel-box";
import { readF32Run, readU32Run, readText, decodeAt } from "../PanelOverlay/panel-columns";
import * as T from "../chrome-theme";
import {
  COL_STREAM_TAB_STRIP_STRIP_X, COL_STREAM_TAB_STRIP_STRIP_Y, COL_STREAM_TAB_STRIP_STRIP_W,
  COL_STREAM_TAB_STRIP_STRIP_H, COL_STREAM_TAB_STRIP_TAB_X, COL_STREAM_TAB_STRIP_TAB_Y,
  COL_STREAM_TAB_STRIP_TAB_W, COL_STREAM_TAB_STRIP_TAB_H, COL_STREAM_TAB_STRIP_TAB_NAME_TEXT,
  COL_STREAM_TAB_STRIP_TAB_NAME_LEN, COL_STREAM_TAB_STRIP_TAB_SELECTED,
} from "./columns-gen";

export function tabStripKey(): string {
  const sel = columnBytes(COL_STREAM_TAB_STRIP_TAB_SELECTED);
  return [
    columnF32(COL_STREAM_TAB_STRIP_STRIP_X), columnF32(COL_STREAM_TAB_STRIP_STRIP_W),
    sel ? new Uint8Array(sel.buffer, sel.byteOffset, sel.byteLength).join(".") : "",
  ].join(",");
}

export function drawTabStrip(c: CanvasRenderingContext2D): void {
  const sw = columnF32(COL_STREAM_TAB_STRIP_STRIP_W);
  const sh = columnF32(COL_STREAM_TAB_STRIP_STRIP_H);
  if (sw <= 0 || sh <= 0) return;
  const sx = columnF32(COL_STREAM_TAB_STRIP_STRIP_X);
  const sy = columnF32(COL_STREAM_TAB_STRIP_STRIP_Y);

  roundRect(c, sx + 0.5, sy + 0.5, sw - 1, sh - 1, T.RADIUS_CHIP);
  c.fillStyle = T.CHIP;
  c.fill();
  c.strokeStyle = T.BORDER;
  c.lineWidth = 1;
  c.stroke();

  const x = readF32Run(COL_STREAM_TAB_STRIP_TAB_X);
  const y = readF32Run(COL_STREAM_TAB_STRIP_TAB_Y);
  const w = readF32Run(COL_STREAM_TAB_STRIP_TAB_W);
  const h = readF32Run(COL_STREAM_TAB_STRIP_TAB_H);
  const nameText = readText(COL_STREAM_TAB_STRIP_TAB_NAME_TEXT);
  const nameLen = readU32Run(COL_STREAM_TAB_STRIP_TAB_NAME_LEN);
  const sel = columnBytes(COL_STREAM_TAB_STRIP_TAB_SELECTED);
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
    c.font = panelFont(T.FONT_SIZE);
    c.textAlign = "center";
    c.textBaseline = "middle";
    c.fillText(name, x[i]! + w[i]! / 2, y[i]! + h[i]! / 2);
  }
}
