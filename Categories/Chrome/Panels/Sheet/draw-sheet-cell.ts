import { canvasFont, roundRect } from "../../canvas-box";

const TEXT_INK = "#222";
const MUTED_INK = "#555";
const FONT_PX = 13;
const HEAD_FONT_PX = 11;
const WEIGHT = 500;
const STRONG_WEIGHT = 700;

const BOX_FILL = "#fff";
const ON_FILL = "#e3eefc";
const BOX_EDGE = "#c8ccd2";
const EDIT_EDGE = "#1565c0";
const BOX_RADIUS = 3;
const CELL_PAD_X = 6;

const CARD_FILL = "#f4f5f7";
const GUIDE_INK = "#dde1e6";

const CELL_TEXT = 0;
const CELL_FIELD = 1;

const BAND_SELECTED = 1;
const BAND_GUIDE = 2;

const INK_STRONG = 1;
const INK_HEAD = 2;

export type R = { x: number; y: number; w: number; h: number };

export type CellLook = { kind: number; key: string; value: string; editing: boolean; outlined: boolean; boxed: boolean; lit: boolean; dim: boolean; bare: boolean; ink: number };

export function drawBand(c: CanvasRenderingContext2D, r: R, band: number): void {
  if (band === BAND_GUIDE) {
    c.fillStyle = GUIDE_INK;
    c.fillRect(Math.round(r.x), r.y, 1, r.h);
    return;
  }
  roundRect(c, r.x + 0.5, r.y + 0.5, r.w - 1, r.h - 1, 4);
  c.fillStyle = band === BAND_SELECTED ? ON_FILL : CARD_FILL;
  c.fill();
}

function box(c: CanvasRenderingContext2D, r: R, on: boolean, editing: boolean): void {
  roundRect(c, r.x + 0.5, r.y + 0.5, r.w - 1, r.h - 1, BOX_RADIUS);
  c.fillStyle = on || editing ? ON_FILL : BOX_FILL;
  c.fill();
  c.strokeStyle = editing ? EDIT_EDGE : BOX_EDGE;
  c.lineWidth = 1;
  c.stroke();
}

function label(c: CanvasRenderingContext2D, text: string, x: number, y: number, align: CanvasTextAlign, ink: string, px = FONT_PX, weight = WEIGHT): void {
  c.fillStyle = ink;
  c.font = canvasFont(px, weight);
  c.textAlign = align;
  c.textBaseline = "middle";
  c.fillText(text, x, y);
}

function drawField(c: CanvasRenderingContext2D, r: R, key: string, value: string, editing: boolean): void {
  box(c, r, false, editing);
  label(c, key, r.x + CELL_PAD_X, r.y + r.h / 2, "left", MUTED_INK);
  label(c, value, r.x + r.w - CELL_PAD_X, r.y + r.h / 2, "right", TEXT_INK);
}

function drawBoxed(c: CanvasRenderingContext2D, r: R, text: string, on: boolean, editing: boolean, dim: boolean): void {
  box(c, r, on, editing);
  label(c, text, r.x + r.w / 2, r.y + r.h / 2, "center", dim ? MUTED_INK : TEXT_INK);
}

function drawText(c: CanvasRenderingContext2D, r: R, text: string, dim: boolean, ink: number): void {
  if (ink === INK_HEAD) return label(c, text, r.x + 1, r.y + r.h / 2, "left", MUTED_INK, HEAD_FONT_PX);
  label(c, text, r.x + 1, r.y + r.h / 2, "left", dim ? MUTED_INK : TEXT_INK, FONT_PX, ink === INK_STRONG ? STRONG_WEIGHT : WEIGHT);
}

export function drawCell(c: CanvasRenderingContext2D, r: R, l: CellLook): void {
  if (l.kind === CELL_FIELD && l.bare && !l.editing) drawText(c, r, l.value || l.key, l.dim, l.ink);
  else if (l.kind === CELL_FIELD && l.boxed) drawBoxed(c, r, l.value || l.key, false, l.editing, l.dim || !l.value);
  else if (l.kind === CELL_FIELD) drawField(c, r, l.key, l.value, l.editing);
  else if (l.kind === CELL_TEXT && l.boxed) drawBoxed(c, r, l.key, l.outlined || l.lit, false, l.dim);
  else if (l.kind === CELL_TEXT) drawText(c, r, l.key, l.dim, l.lit && !l.ink ? INK_STRONG : l.ink);
  else drawBoxed(c, r, l.key, l.lit, false, l.dim);
}

export function drawArrow(c: CanvasRenderingContext2D, r: R, up: boolean, lit: boolean, dim: boolean): void {
  if (r.w <= 0 || r.h <= 0) return;
  box(c, { x: r.x, y: r.y + 0.5, w: r.w, h: r.h - 1 }, lit && !dim, false);
  const cx = r.x + r.w / 2, cy = r.y + r.h / 2, half = Math.min(r.w, r.h) * 0.28;
  c.beginPath();
  c.moveTo(cx - half, cy + (up ? 1 : -1) * half * 0.6);
  c.lineTo(cx + half, cy + (up ? 1 : -1) * half * 0.6);
  c.lineTo(cx, cy - (up ? 1 : -1) * half * 0.6);
  c.closePath();
  c.fillStyle = dim ? MUTED_INK : TEXT_INK;
  c.fill();
}

export function slid(c: CanvasRenderingContext2D, b: R, x: number, w: number, on: boolean, draw: () => void): void {
  if (!on || w <= 0) return draw();
  c.save();
  c.beginPath();
  c.rect(x, b.y, w, b.h);
  c.clip();
  try { draw(); } finally { c.restore(); }
}
