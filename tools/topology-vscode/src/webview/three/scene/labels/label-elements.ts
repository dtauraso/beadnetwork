import type { BufferLabelPos } from "../buffer-scene-shared";

const elements = new Map<number, HTMLDivElement>();

const applied = new Map<number, { px: number; py: number }>();

export function registerLabelElement(row: number, el: HTMLDivElement | null): void {
  if (el) {
    elements.set(row, el);
  } else {
    elements.delete(row);
    applied.delete(row);
  }
}

export function applyLabelPositions(positions: BufferLabelPos[]): void {
  for (const p of positions) {
    const el = elements.get(p.row);
    if (!el) continue;
    el.style.transform = `translate(${p.px}px, ${p.py - 4}px) translate(-50%, -100%)`;
    applied.set(p.row, { px: p.px, py: p.py });
  }
}

export function appliedPositions(): Map<number, { px: number; py: number }> {
  return applied;
}
