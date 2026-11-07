import { drawBox, canvasFont, roundRect } from "../../canvas-box";
import * as T from "../../chrome-theme";
import { decodeAt } from "../../leaf-text";
import { CARET_W, drawPill } from "../../Pills/pill";
import { drawArrow, drawCell, slid, type R } from "./draw-sheet-cell";
import { sheetBytes, sheetF32Run, sheetU32Run, sheetText } from "./sheet-leaves";

const TITLE_FONT_PX = 12;
const TITLE_INK = "#222";

function bytesOf(name: "cellKind" | "cellLit" | "cellEditing" | "pillBar" | "tabOn" | "cellOutlined" | "cellBoxed" | "cellSlides" | "upSlides" | "upLit" | "downLit" | "upDim" | "cellDim" | "pillOpen"): number[] {
  const v = sheetBytes(name);
  const out: number[] = [];
  if (v) for (let i = 0; i < v.byteLength; i++) out.push(v.getUint8(i));
  return out;
}

function strings(text: "pillText" | "tabText" | "titleText" | "keyText" | "valueText", lens: "pillLen" | "tabLen" | "titleLen" | "keyLen" | "valueLen"): string[] {
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
    bytesOf("cellDim").join(""),
    bytesOf("upLit").join(""),
    bytesOf("downLit").join(""),
    bytesOf("pillOpen").join(""),
    bytesOf("tabOn").join(""),
    Array.from(sheetF32Run("tabX") ?? []).join("."),
    bytesOf("cellLit").join(""),
    bytesOf("cellEditing").join(""),
    strings("keyText", "keyLen").join("|"),
    strings("valueText", "valueLen").join("|"),
    draftText(),
  ].join(",");
}

function rects(x: "pillX" | "tabX" | "boxX" | "cellX" | "upX" | "downX", y: "pillY" | "tabY" | "boxY" | "cellY" | "upY" | "downY",
  w: "pillW" | "tabW" | "boxW" | "cellW" | "upW" | "downW", h: "pillH" | "tabH" | "boxH" | "cellH" | "upH" | "downH"): R[] {
  const xs = sheetF32Run(x), ys = sheetF32Run(y), ws = sheetF32Run(w), hs = sheetF32Run(h);
  const out: R[] = [];
  if (!xs || !ys || !ws || !hs) return out;
  for (let i = 0; i < xs.length; i++) out.push({ x: xs[i]!, y: ys[i]!, w: ws[i]!, h: hs[i]! });
  return out;
}

type Tab = { r: R; label: string; on: boolean };

function drawTabBar(c: CanvasRenderingContext2D, bar: R, tabs: Tab[], folded: boolean): void {
  roundRect(c, bar.x + 0.5, bar.y + 0.5, bar.w - 1, bar.h - 1, T.RADIUS_CHIP);
  c.fillStyle = T.SURFACE;
  c.fill();
  c.strokeStyle = T.BORDER;
  c.lineWidth = 1;
  c.stroke();

  c.textAlign = "center";
  c.textBaseline = "middle";
  for (const t of tabs) {
    if (t.on) {
      roundRect(c, t.r.x, t.r.y, t.r.w, t.r.h, T.RADIUS_CHIP);
      c.fillStyle = T.HOVER_CHIP;
      c.fill();
    }
    c.globalAlpha = t.on ? 1 : T.DISABLED_OPACITY + 0.2;
    c.fillStyle = T.TEXT;
    c.font = canvasFont(T.FONT_SIZE, T.FONT_WEIGHT_LABEL);
    c.fillText(t.label, t.r.x + t.r.w / 2, t.r.y + t.r.h / 2);
  }
  c.globalAlpha = 1;

  c.fillStyle = T.TEXT;
  c.font = canvasFont(T.FONT_SIZE_GLYPH);
  c.fillText(folded ? "▲" : "▼", bar.x + bar.w - CARET_W / 2 - 3, bar.y + bar.h / 2);
}

export function drawSheets(c: CanvasRenderingContext2D): void {
  const pills = rects("pillX", "pillY", "pillW", "pillH");
  const labels = strings("pillText", "pillLen");
  const open = bytesOf("pillOpen");
  const bar = bytesOf("pillBar");
  const tabRects = rects("tabX", "tabY", "tabW", "tabH");
  const tabPill = sheetU32Run("tabPill"), tabLabels = strings("tabText", "tabLen"), tabOn = bytesOf("tabOn");
  pills.forEach((p, i) => {
    if (p.w <= 0) return;
    if (bar[i]) {
      const tabs: Tab[] = [];
      tabRects.forEach((r, ti) => {
        if (tabPill?.[ti] === i) tabs.push({ r, label: tabLabels[ti] ?? "", on: tabOn[ti] !== 0 });
      });
      drawTabBar(c, p, tabs, open[i] !== 0);
    }
    else drawPill(c, p.x, p.y, p.w, p.h, labels[i] ?? "", open[i] !== 0);
  });

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
  const outlined = bytesOf("cellOutlined"), boxed = bytesOf("cellBoxed"), dimCell = bytesOf("cellDim");
  const slideX = sheetF32Run("slideX"), slideW = sheetF32Run("slideW");
  const cellSlides = bytesOf("cellSlides"), upSlides = bytesOf("upSlides");
  const upLit = bytesOf("upLit"), downLit = bytesOf("downLit"), upDim = bytesOf("upDim");

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
            outlined: (outlined[i] ?? 0) !== 0, boxed: (boxed[i] ?? 0) !== 0, lit: (lit[i] ?? 0) !== 0, dim: (dimCell[i] ?? 0) !== 0,
          }));
        });
      }
      if (upBox) {
        ups.forEach((r, i) => {
          if (upBox[i] !== bi) return;
          slid(c, b, sx, sw, (upSlides[i] ?? 0) !== 0, () => {
            const dim = (upDim[i] ?? 0) !== 0;
            drawArrow(c, r, true, (upLit[i] ?? 0) !== 0, dim);
            const down = downs[i];
            if (down) drawArrow(c, down, false, (downLit[i] ?? 0) !== 0, dim);
          });
        });
      }
    } finally {
      c.restore();
    }
  });
}
