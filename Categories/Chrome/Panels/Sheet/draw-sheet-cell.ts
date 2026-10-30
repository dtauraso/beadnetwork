import { canvasFont, roundRect } from "../../canvas-box";

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

export type R = { x: number; y: number; w: number; h: number };

export type CellLook = { kind: number; key: string; value: string; editing: boolean; outlined: boolean; boxed: boolean; lit: boolean };

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

function drawBoxed(c: CanvasRenderingContext2D, r: R, text: string, outlined: boolean, lit: boolean): void {
  if (outlined) {
    roundRect(c, r.x + 0.75, r.y + 0.75, r.w - 1.5, r.h - 1.5, CELL_RADIUS);
    c.strokeStyle = OUTLINE_EDGE;
    c.lineWidth = 1.5;
    c.stroke();
  }
  c.fillStyle = lit ? TEXT_LIT_INK : TEXT_INK;
  c.font = canvasFont(KEY_FONT_PX, lit ? 700 : 500);
  c.textAlign = "center";
  c.textBaseline = "middle";
  c.fillText(text, r.x + r.w / 2, r.y + r.h / 2);
}

export function drawCell(c: CanvasRenderingContext2D, r: R, l: CellLook): void {
  if (l.kind === CELL_FIELD) {
    drawField(c, r, l.key, l.value, l.editing);
    if (l.outlined) {
      roundRect(c, r.x - 1.5, r.y - 1.5, r.w + 3, r.h + 3, CELL_RADIUS + 1);
      c.strokeStyle = OUTLINE_EDGE;
      c.lineWidth = 2;
      c.stroke();
    }
  }
  else if (l.kind === CELL_TEXT && l.boxed) drawBoxed(c, r, l.key, l.outlined, l.lit);
  else if (l.kind === CELL_TEXT) drawText(c, r, l.key, l.lit);
  else drawButton(c, r, l.key, l.lit);
}

export function drawArrow(c: CanvasRenderingContext2D, r: R, up: boolean): void {
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

export function slid(c: CanvasRenderingContext2D, b: R, x: number, w: number, on: boolean, draw: () => void): void {
  if (!on || w <= 0) return draw();
  c.save();
  c.beginPath();
  c.rect(x, b.y, w, b.h);
  c.clip();
  try { draw(); } finally { c.restore(); }
}
