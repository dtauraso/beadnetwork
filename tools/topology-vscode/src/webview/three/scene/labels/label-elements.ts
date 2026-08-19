import * as T from "../../controls/chrome-theme";
import type { BufferLabelPos } from "../buffer-scene-shared";
import { probeLabelLag } from "./label-lag-probe";

const elements = new Map<number, HTMLDivElement>();

const applied = new Map<number, { px: number; py: number }>();

let layer: HTMLDivElement | null = null;

export function setLabelLayer(el: HTMLDivElement | null): void {
  layer = el;
  if (el) return;
  elements.clear();
  applied.clear();
}

function createLabel(): HTMLDivElement {
  const el = document.createElement("div");
  const s = el.style;
  s.position = "absolute";
  s.left = "0";
  s.top = "0";
  s.fontSize = `${T.FONT_SIZE}px`;
  s.fontFamily = T.FONT_STACK;
  s.fontVariantNumeric = "tabular-nums";
  s.color = T.TEXT;
  s.pointerEvents = "none";
  s.lineHeight = "1.25";
  s.textAlign = "center";
  s.whiteSpace = "nowrap";
  s.zIndex = "10";
  s.background = T.CHIP;
  s.border = `1px solid ${T.BORDER}`;
  s.borderRadius = `${T.RADIUS_ITEM}px`;
  s.padding = T.PAD_CHIP;
  return el;
}

export function applyLabels(positions: BufferLabelPos[]): void {
  if (!layer) return;

  probeLabelLag(positions, applied);

  const live = new Set<number>();
  for (const p of positions) {
    live.add(p.row);
    let el = elements.get(p.row);
    if (!el) {
      el = createLabel();
      elements.set(p.row, el);
      layer.appendChild(el);
    }
    const text = p.label || String(p.row);
    if (el.textContent !== text) el.textContent = text;
    el.style.transform = `translate(${p.px}px, ${p.py - 4}px) translate(-50%, -100%)`;
    applied.set(p.row, { px: p.px, py: p.py });
  }

  for (const [row, el] of elements) {
    if (live.has(row)) continue;
    el.remove();
    elements.delete(row);
    applied.delete(row);
  }
}
