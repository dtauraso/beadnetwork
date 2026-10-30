import { drawBox, canvasFont, roundRect } from "../../canvas-box";
import { decodeAt } from "../../leaf-text";
import { drawPill } from "../../Pills/pill";
import { sheetBytes, sheetF32Run, sheetU32Run, sheetText } from "./sheet-leaves";

const TITLE_FONT_PX = 12;
const TITLE_INK = "#222";
const TEXT_INK = "#222";
const TEXT_LIT_INK = "#0d47a1";

const CELL_FILL = "#222";
const CELL_EDIT_FILL = "#0d47a1";
const CELL_RADIUS = 3;
const CELL_PAD_X = 6;
const KEY_FONT_PX = 11;
const KEY_INK = "#e0e0e0";
const VAL_FONT_PX = 13;
const VAL_INK = "#fff";

const BUTTON_FILL = "#f2f2f2";
const BUTTON_EDGE = "#bbb";
const BUTTON_LIT_FILL = "#0d47a1";

const OUTLINE_EDGE = "#0d47a1";

const CELL_TEXT = 0;
const CELL_FIELD = 1;

function bytesOf(name: "cellKind" | "cellLit" | "cellEditing" | "pillOpen"): number[] {
  const v = sheetBytes(name);
  const out: number[] = [];
  if (v) for (let i = 0; i < v.byteLength; i++) out.push(v.getUint8(i));
  return out;
}

function strings(text: "pillText" | "titleText" | "keyText" | "valueText", lens: "pillLen" | "titleLen" | "keyLen" | "valueLen"): string[] {
  const bytes = sheetText(text);
  const len = sheetU32Run(lens);
  const out: string[] = [];
  if (!bytes || !len) return out;
  let off = 0;
  for (let i = 0; i < len.length; i++) {
    out.push(decodeAt(bytes, off, len[i]!));
    off += len[i]!;
  }
  return out;
}

function draftText(): string {
  const d = sheetText("draftText");
  return d ? decodeAt(d, 0, d.length) : "";
}

export function sheetDraftOpen(): boolean {
  return bytesOf("cellEditing").some((b) => b !== 0);
}

export function sheetKey(): string {
  return [
    Array.from(sheetF32Run("pillY") ?? []).join("."),
    Array.from(sheetF32Run("boxH") ?? []).join("."),
    Array.from(sheetF32Run("boxW") ?? []).join("."),
    Array.from(sheetF32Run("cellY") ?? []).join("."),
    Array.from(sheetF32Run("cellX") ?? []).join("."),
    Array.from(sheetF32Run("outlineY") ?? []).join("."),
    bytesOf("pillOpen").join(""),
    bytesOf("cellLit").join(""),
    bytesOf("cellEditing").join(""),
    strings("keyText", "keyLen").join("|"),
    strings("valueText", "valueLen").join("|"),
    draftText(),
  ].join(",");
}

type R = { x: number; y: number; w: number; h: number };

function rects(x: "pillX" | "boxX" | "cellX" | "upX" | "downX" | "outlineX", y: "pillY" | "boxY" | "cellY" | "upY" | "downY" | "outlineY",
  w: "pillW" | "boxW" | "cellW" | "upW" | "downW" | "outlineW", h: "pillH" | "boxH" | "cellH" | "upH" | "downH" | "outlineH"): R[] {
  const xs = sheetF32Run(x), ys = sheetF32Run(y), ws = sheetF32Run(w), hs = sheetF32Run(h);
  const out: R[] = [];
  if (!xs || !ys || !ws || !hs) return out;
  for (let i = 0; i < xs.length; i++) out.push({ x: xs[i]!, y: ys[i]!, w: ws[i]!, h: hs[i]! });
  return out;
}

function drawField(c: CanvasRenderingContext2D, r: R, key: string, value: string, editing: boolean): void {
  roundRect(c, r.x, r.y, r.w, r.h, CELL_RADIUS);
  c.fillStyle = editing ? CELL_EDIT_FILL : CELL_FILL;
  c.fill();
  c.textBaseline = "middle";
  c.fillStyle = KEY_INK;
  c.font = canvasFont(KEY_FONT_PX, 600);
  c.textAlign = "left";
  c.fillText(key, r.x + CELL_PAD_X, r.y + r.h / 2);
  c.fillStyle = VAL_INK;
  c.font = canvasFont(VAL_FONT_PX, 700);
  c.textAlign = "right";
  c.fillText(value, r.x + r.w - CELL_PAD_X, r.y + r.h / 2);
}

function drawButton(c: CanvasRenderingContext2D, r: R, label: string, lit: boolean): void {
  roundRect(c, r.x + 0.5, r.y + 0.5, r.w - 1, r.h - 1, CELL_RADIUS);
  c.fillStyle = lit ? BUTTON_LIT_FILL : BUTTON_FILL;
  c.fill();
  c.strokeStyle = BUTTON_EDGE;
  c.lineWidth = 1;
  c.stroke();
  c.fillStyle = lit ? VAL_INK : TEXT_INK;
  c.font = canvasFont(KEY_FONT_PX, 600);
  c.textAlign = "center";
  c.textBaseline = "middle";
  c.fillText(label, r.x + r.w / 2, r.y + r.h / 2);
}

function drawText(c: CanvasRenderingContext2D, r: R, text: string, lit: boolean): void {
  c.fillStyle = lit ? TEXT_LIT_INK : TEXT_INK;
  c.font = canvasFont(KEY_FONT_PX, lit ? 700 : 500);
  c.textAlign = "left";
  c.textBaseline = "middle";
  c.fillText(text, r.x + 1, r.y + r.h / 2);
}

function drawArrow(c: CanvasRenderingContext2D, r: R, up: boolean): void {
  if (r.w <= 0 || r.h <= 0) return;
  roundRect(c, r.x, r.y + 0.5, r.w, r.h - 1, 2);
  c.fillStyle = CELL_FILL;
  c.fill();
  const cx = r.x + r.w / 2, cy = r.y + r.h / 2, half = Math.min(r.w, r.h) * 0.28;
  c.beginPath();
  c.moveTo(cx - half, cy + (up ? 1 : -1) * half * 0.6);
  c.lineTo(cx + half, cy + (up ? 1 : -1) * half * 0.6);
  c.lineTo(cx, cy - (up ? 1 : -1) * half * 0.6);
  c.closePath();
  c.fillStyle = VAL_INK;
  c.fill();
}

export function drawSheets(c: CanvasRenderingContext2D): void {
  const pills = rects("pillX", "pillY", "pillW", "pillH");
  const labels = strings("pillText", "pillLen");
  const open = bytesOf("pillOpen");
  pills.forEach((p, i) => { if (p.w > 0) drawPill(c, p.x, p.y, p.w, p.h, labels[i] ?? "", open[i] !== 0); });

  const boxes = rects("boxX", "boxY", "boxW", "boxH");
  const headBox = sheetU32Run("headBox"), headX = sheetF32Run("headX"), headY = sheetF32Run("headY"), headH = sheetF32Run("headH");
  const titles = strings("titleText", "titleLen");
  const cellBox = sheetU32Run("cellBox");
  const cells = rects("cellX", "cellY", "cellW", "cellH");
  const kinds = bytesOf("cellKind"), lit = bytesOf("cellLit"), editing = bytesOf("cellEditing");
  const keys = strings("keyText", "keyLen"), vals = strings("valueText", "valueLen");
  const upBox = sheetU32Run("upBox");
  const ups = rects("upX", "upY", "upW", "upH"), downs = rects("downX", "downY", "downW", "downH");
  const draft = draftText();
  const outlineBox = sheetU32Run("outlineBox");
  const outlines = rects("outlineX", "outlineY", "outlineW", "outlineH");

  boxes.forEach((b, bi) => {
    if (b.w <= 0 || b.h <= 0) return;
    drawBox(c, b.x, b.y, b.w, b.h);
    c.save();
    c.beginPath();
    c.rect(b.x, b.y, b.w, b.h);
    c.clip();
    try {
      if (headBox && headX && headY && headH) {
        for (let i = 0; i < headBox.length; i++) {
          if (headBox[i] !== bi) continue;
          c.fillStyle = TITLE_INK;
          c.font = canvasFont(TITLE_FONT_PX, 700);
          c.textAlign = "left";
          c.textBaseline = "middle";
          c.fillText(titles[i] ?? "", headX[i]!, headY[i]! + headH[i]! / 2);
        }
      }
      if (cellBox) {
        cells.forEach((r, i) => {
          if (cellBox[i] !== bi || r.w <= 0) return;
          const isEditing = editing[i] !== 0;
          if (kinds[i] === CELL_FIELD) drawField(c, r, keys[i] ?? "", isEditing ? `${draft}_` : vals[i] ?? "", isEditing);
          else if (kinds[i] === CELL_TEXT) drawText(c, r, keys[i] ?? "", lit[i] !== 0);
          else drawButton(c, r, keys[i] ?? "", lit[i] !== 0);
        });
      }
      if (outlineBox) {
        outlines.forEach((r, i) => {
          if (outlineBox[i] !== bi) return;
          roundRect(c, r.x + 0.5, r.y + 0.5, r.w - 1, r.h - 1, CELL_RADIUS + 1);
          c.strokeStyle = OUTLINE_EDGE;
          c.lineWidth = 1.5;
          c.stroke();
        });
      }
      if (upBox) {
        ups.forEach((r, i) => {
          if (upBox[i] !== bi) return;
          drawArrow(c, r, true);
          const down = downs[i];
          if (down) drawArrow(c, down, false);
        });
      }
    } finally {
      c.restore();
    }
  });
}
