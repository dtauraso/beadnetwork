import { drawBox, canvasFont } from "../../canvas-box";
import { decodeAt } from "../../leaf-text";
import { drawPill } from "../../Pills/pill";
import { drawArrow, drawCell, slid, type R } from "./draw-sheet-cell";
import { sheetBytes, sheetF32Run, sheetU32Run, sheetText } from "./sheet-leaves";

const TITLE_FONT_PX = 12;
const TITLE_INK = "#222";

function bytesOf(name: "cellKind" | "cellLit" | "cellEditing" | "cellOutlined" | "cellBoxed" | "cellSlides" | "upSlides" | "upLit" | "downLit" | "pillOpen"): number[] {
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
    bytesOf("cellOutlined").join(""),
    bytesOf("pillOpen").join(""),
    bytesOf("cellLit").join(""),
    bytesOf("cellEditing").join(""),
    strings("keyText", "keyLen").join("|"),
    strings("valueText", "valueLen").join("|"),
    draftText(),
  ].join(",");
}

function rects(x: "pillX" | "boxX" | "cellX" | "upX" | "downX", y: "pillY" | "boxY" | "cellY" | "upY" | "downY",
  w: "pillW" | "boxW" | "cellW" | "upW" | "downW", h: "pillH" | "boxH" | "cellH" | "upH" | "downH"): R[] {
  const xs = sheetF32Run(x), ys = sheetF32Run(y), ws = sheetF32Run(w), hs = sheetF32Run(h);
  const out: R[] = [];
  if (!xs || !ys || !ws || !hs) return out;
  for (let i = 0; i < xs.length; i++) out.push({ x: xs[i]!, y: ys[i]!, w: ws[i]!, h: hs[i]! });
  return out;
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
  const outlined = bytesOf("cellOutlined"), boxed = bytesOf("cellBoxed");
  const slideX = sheetF32Run("slideX"), slideW = sheetF32Run("slideW");
  const cellSlides = bytesOf("cellSlides"), upSlides = bytesOf("upSlides");
  const upLit = bytesOf("upLit"), downLit = bytesOf("downLit");

  boxes.forEach((b, bi) => {
    if (b.w <= 0 || b.h <= 0) return;
    drawBox(c, b.x, b.y, b.w, b.h);
    c.save();
    c.beginPath();
    c.rect(b.x, b.y, b.w, b.h);
    c.clip();
    const sx = slideX?.[bi] ?? 0, sw = slideW?.[bi] ?? 0;
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
          slid(c, b, sx, sw, (cellSlides[i] ?? 0) !== 0, () => drawCell(c, r, {
            kind: kinds[i] ?? 0, key: keys[i] ?? "", value: isEditing ? `${draft}_` : vals[i] ?? "", editing: isEditing,
            outlined: (outlined[i] ?? 0) !== 0, boxed: (boxed[i] ?? 0) !== 0, lit: (lit[i] ?? 0) !== 0,
          }));
        });
      }
      if (upBox) {
        ups.forEach((r, i) => {
          if (upBox[i] !== bi) return;
          slid(c, b, sx, sw, (upSlides[i] ?? 0) !== 0, () => {
            drawArrow(c, r, true, (upLit[i] ?? 0) !== 0);
            const down = downs[i];
            if (down) drawArrow(c, down, false, (downLit[i] ?? 0) !== 0);
          });
        });
      }
    } finally {
      c.restore();
    }
  });
}
