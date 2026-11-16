import { canvasFont, drawCaret, roundRect } from "../../canvas-box";
import { decodeAt } from "../../leaf-text";
import { BOARD_VALUE_NAMES, type BoardValueName } from "./board-values-gen";
import { boardBytes, boardF32Run, boardText, boardU32Run } from "./board-leaves";
import { drawDial } from "./draw-dial";

const PAINT = ["", "#222", "#555", "#f4f5f7", "#dde1e6", "#e3eefc", "#c8ccd2", "#1565c0", "#FF9F1C", "#3DDC84", "#fff", "#b26a00", "#fff4e0", "#9aa0a8", "#fce8e8", "#b0b4ba", "#fafafa", "#f0f1f3", "#c4c7cc", "#c62828"];
const K_BOX = 0, K_TEXT = 1, K_DIAL = 2, K_LINE = 3, K_CIRCLE = 4;

type R = { x: number; y: number; w: number; h: number };

function u8s(name: BoardValueName): Uint8Array {
  const v = boardBytes(name);
  return v ? new Uint8Array(v.buffer, v.byteOffset, v.byteLength) : new Uint8Array(0);
}

function f32s(name: BoardValueName): Float32Array {
  return boardF32Run(name) ?? new Float32Array(0);
}

export function boardDraftOpen(): boolean {
  return u8s("editing").some((b) => b !== 0);
}

export function boardKey(): string {
  let h = 0;
  for (const name of BOARD_VALUE_NAMES) {
    const b = u8s(name);
    h = (h * 31 + b.length) | 0;
    for (let i = 0; i < b.length; i++) h = (h * 31 + b[i]!) | 0;
  }
  return String(h);
}

function texts(): string[] {
  const bytes = boardText("markText"), lens = boardU32Run("markLen");
  const out: string[] = [];
  if (!bytes || !lens) return out;
  let off = 0;
  for (let i = 0; i < lens.length; i++) {
    out.push(decodeAt(bytes, off, lens[i]!));
    off += lens[i]!;
  }
  return out;
}

function corners(c: CanvasRenderingContext2D, x: number, y: number, w: number, h: number, r: number, which: number): void {
  const rr = Math.max(0, Math.min(r, w / 2, h / 2)), top = which === 1 ? 0 : rr, bot = which === 2 ? 0 : rr;
  c.beginPath();
  c.moveTo(x + top, y);
  c.arcTo(x + w, y, x + w, y + h, top);
  c.arcTo(x + w, y + h, x, y + h, bot);
  c.arcTo(x, y + h, x, y, bot);
  c.arcTo(x, y, x + w, y, top);
  c.closePath();
}

export function drawBoard(c: CanvasRenderingContext2D): void {
  const kind = u8s("markKind");
  if (kind.length === 0) return;
  const mx = f32s("markX"), my = f32s("markY"), mw = f32s("markW"), mh = f32s("markH");
  const clip = u8s("markClip"), fill = u8s("markFill"), edge = u8s("markEdge"), ink = u8s("markInk");
  const radius = f32s("markRadius"), stroke = f32s("markStroke"), px = f32s("markPx"), alpha = f32s("markAlpha");
  const bold = u8s("markBold"), dash = u8s("markDash"), corner = u8s("markCorner"), labels = texts();
  const dAxis = u8s("dialAxis"), dStart = f32s("dialStart"), dTurn = f32s("dialTurn"), dInc = f32s("dialInc"), dLabels = u8s("dialLabels");
  const cx = f32s("clipX"), cy = f32s("clipY"), cw = f32s("clipW"), ch = f32s("clipH");
  const n = kind.length, caretMark = boardU32Run("caretMark")?.[0] ?? 0, caretAt = boardU32Run("caretAt")?.[0] ?? 0;
  if ([mx, my, mw, mh, clip, fill, edge, radius, alpha].some((a) => a.length !== n)) return;
  for (let i = 0; i < kind.length; i++) {
    const r = { x: mx[i]!, y: my[i]!, w: mw[i]!, h: mh[i]! }, k = clip[i] ?? 0;
    c.save();
    c.beginPath();
    c.rect(cx[k] ?? 0, cy[k] ?? 0, cw[k] ?? 0, ch[k] ?? 0);
    c.clip();
    c.globalAlpha = alpha[i] ?? 1;
    c.setLineDash(dash[i] ? [4, 3] : []);
    if (kind[i] === K_BOX) {
      if (r.w < 1 || r.h < 1) { c.restore(); continue; }
      if (corner[i]) corners(c, r.x + 0.5, r.y + 0.5, r.w - 1, r.h - 1, radius[i] ?? 0, corner[i]!);
      else roundRect(c, r.x + 0.5, r.y + 0.5, r.w - 1, r.h - 1, radius[i] ?? 0);
      if (fill[i]) { c.fillStyle = PAINT[fill[i]!]!; c.fill(); }
      if (edge[i]) { c.strokeStyle = PAINT[edge[i]!]!; c.lineWidth = stroke[i] || 1; c.stroke(); }
    } else if (kind[i] === K_TEXT) {
      c.fillStyle = PAINT[ink[i]!] || PAINT[1]!;
      const b = bold[i] ?? 0;
      c.font = canvasFont(px[i] || 13, b === 0 ? 400 : b === 1 ? 700 : b * 100);
      c.textAlign = "left";
      c.textBaseline = "middle";
      c.fillText(labels[i] ?? "", r.x, r.y + r.h / 2);
      if (caretMark === i + 1) drawCaret(c, labels[i] ?? "", r.x, r.y, r.h, caretAt, PAINT[ink[i]!] || PAINT[1]!);
    } else if (kind[i] === K_LINE) {
      c.strokeStyle = PAINT[ink[i]!] || PAINT[1]!;
      c.lineWidth = stroke[i] || 1;
      c.lineCap = "round";
      c.beginPath();
      c.moveTo(r.x, r.y);
      c.lineTo(r.x + r.w, r.y + r.h);
      c.stroke();
    } else if (kind[i] === K_CIRCLE) {
      c.beginPath();
      c.arc(r.x + r.w / 2, r.y + r.h / 2, Math.max(0, Math.min(r.w, r.h) / 2), 0, 2 * Math.PI);
      if (fill[i]) { c.fillStyle = PAINT[fill[i]!]!; c.fill(); }
      if (edge[i]) { c.strokeStyle = PAINT[edge[i]!]!; c.lineWidth = stroke[i] || 1; c.stroke(); }
    } else if (kind[i] === K_DIAL) {
      drawDial(c, r, { axis: dAxis[i]!, start: dStart[i]!, turn: dTurn[i]!, inc: dInc[i]!, labels: !!dLabels[i] }, PAINT[dAxis[i]!] || PAINT[8]!);
    }
    c.restore();
  }
}
