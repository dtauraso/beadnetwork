import { columnBytes, columnF32, columnI32, columnU8 } from "../Buffer/column-values";
import { panelFont, roundRect } from "../PanelOverlay/panel-box";
import { readF32Run, readI32Run, readU32Run, readText, decodeAt } from "../PanelOverlay/panel-columns";
import {
  ROW_NODE_HEAD, ROW_HOLDER_HEAD, ROW_LINE,
  CHECK_NONE, VAL_NONE,
  nodeRuleGroup, checkValue, valueText, drawCheckbox,
  PANEL_BG, PANEL_EDGE, INK, NODE_INK, MUTED, FREE_INK, GLYPH_INK, FREE_GLYPH,
  EDIT_BG, EDIT_EDGE, PILL_BG, PILL_EDGE, PILL_INK, INACTIVE_ALPHA, FONT_PX, HEAD_FONT_PX,
} from "./rules-values";
import { drawSharedMenu } from "./draw-shared-menu";
import { drawThetaDraft, draftText, isRowEditing } from "./draw-theta-draft";
import {
  COL_STREAM_RULES_PANEL_BOX_X, COL_STREAM_RULES_PANEL_BOX_Y,
  COL_STREAM_RULES_PANEL_BOX_W, COL_STREAM_RULES_PANEL_BOX_H,
  COL_STREAM_RULES_PANEL_CLIP_Y, COL_STREAM_RULES_PANEL_CLIP_H,
  COL_STREAM_RULES_PANEL_SCROLL_Y,
  COL_STREAM_RULES_PANEL_OPEN,
  COL_STREAM_RULES_PANEL_TOGGLE_X, COL_STREAM_RULES_PANEL_TOGGLE_Y,
  COL_STREAM_RULES_PANEL_TOGGLE_H, COL_STREAM_RULES_PANEL_TOGGLE_TEXT,
  COL_STREAM_RULES_PANEL_ROW_KIND,
  COL_STREAM_RULES_PANEL_ROW_X, COL_STREAM_RULES_PANEL_ROW_Y,
  COL_STREAM_RULES_PANEL_ROW_W, COL_STREAM_RULES_PANEL_ROW_H,
  COL_STREAM_RULES_PANEL_ROW_TEXT_DATA, COL_STREAM_RULES_PANEL_ROW_TEXT_LEN,
  COL_STREAM_RULES_PANEL_ROW_GLYPH_DATA, COL_STREAM_RULES_PANEL_ROW_GLYPH_LEN,
  COL_STREAM_RULES_PANEL_ROW_FREE,
  COL_STREAM_RULES_PANEL_ROW_NODE_ROW, COL_STREAM_RULES_PANEL_ROW_EDGE_ROW,
  COL_STREAM_RULES_PANEL_ROW_CHECK,
  COL_STREAM_RULES_PANEL_ROW_CHECK_X, COL_STREAM_RULES_PANEL_ROW_CHECK_Y,
  COL_STREAM_RULES_PANEL_ROW_CHECK_W, COL_STREAM_RULES_PANEL_ROW_CHECK_H,
  COL_STREAM_RULES_PANEL_ROW_VALUE,
  COL_STREAM_RULES_PANEL_ROW_VALUE_X, COL_STREAM_RULES_PANEL_ROW_VALUE_Y,
  COL_STREAM_RULES_PANEL_ROW_SHARED_X, COL_STREAM_RULES_PANEL_ROW_SHARED_Y,
  COL_STREAM_RULES_PANEL_ROW_SHARED_W, COL_STREAM_RULES_PANEL_ROW_SHARED_H,
  COL_STREAM_RULES_PANEL_MENU_OPEN,
} from "../Buffer/column-streams-gen";

export function rulesPanelKey(): string {
  const nodeRows = readI32Run(COL_STREAM_RULES_PANEL_ROW_NODE_ROW);
  const values = readU32Run(COL_STREAM_RULES_PANEL_ROW_VALUE);
  const edgeRows = readI32Run(COL_STREAM_RULES_PANEL_ROW_EDGE_ROW);
  const checks = columnBytes(COL_STREAM_RULES_PANEL_ROW_CHECK);
  const parts: string[] = [
    String(columnU8(COL_STREAM_RULES_PANEL_OPEN)),
    String(columnF32(COL_STREAM_RULES_PANEL_BOX_H)),
    String(columnF32(COL_STREAM_RULES_PANEL_SCROLL_Y)),
    String(columnU8(COL_STREAM_RULES_PANEL_MENU_OPEN)),
    draftText(),
  ];
  if (nodeRows && values && edgeRows && checks) {
    for (let i = 0; i < nodeRows.length; i++) {
      const v = values[i]!;
      if (v !== VAL_NONE) parts.push(valueText(v, nodeRows[i]!, edgeRows[i]!).text);
      const ck = checks.getUint8(i);
      if (ck !== CHECK_NONE) parts.push(checkValue(ck, nodeRows[i]!, edgeRows[i]!) ? "1" : "0");
      parts.push(String(nodeRuleGroup(nodeRows[i]!).size));
    }
  }
  return parts.join(",");
}

export function drawRulesPanel(c: CanvasRenderingContext2D): void {
  const boxW = columnF32(COL_STREAM_RULES_PANEL_BOX_W);
  const boxH = columnF32(COL_STREAM_RULES_PANEL_BOX_H);
  const toggleText = readText(COL_STREAM_RULES_PANEL_TOGGLE_TEXT);
  if (boxW <= 0 || boxH <= 0 || !toggleText) return;

  roundRect(
    c,
    columnF32(COL_STREAM_RULES_PANEL_BOX_X) + 0.5, columnF32(COL_STREAM_RULES_PANEL_BOX_Y) + 0.5,
    boxW - 1, boxH - 1, 6,
  );
  c.fillStyle = PANEL_BG;
  c.fill();
  c.strokeStyle = PANEL_EDGE;
  c.lineWidth = 1;
  c.stroke();

  c.fillStyle = MUTED;
  c.font = panelFont(HEAD_FONT_PX);
  c.textAlign = "left";
  c.textBaseline = "middle";
  c.fillText(
    decodeAt(toggleText, 0, toggleText.length),
    columnF32(COL_STREAM_RULES_PANEL_TOGGLE_X),
    columnF32(COL_STREAM_RULES_PANEL_TOGGLE_Y) + columnF32(COL_STREAM_RULES_PANEL_TOGGLE_H) / 2,
  );

  if (columnU8(COL_STREAM_RULES_PANEL_OPEN) === 0) return;

  const clipY = columnF32(COL_STREAM_RULES_PANEL_CLIP_Y);
  const clipH = columnF32(COL_STREAM_RULES_PANEL_CLIP_H);
  const boxX = columnF32(COL_STREAM_RULES_PANEL_BOX_X);
  c.save();
  c.beginPath();
  c.rect(boxX, clipY, boxW, clipH);
  c.clip();
  drawRows(c);
  drawThetaDraft(c);
  c.restore();

  if (columnU8(COL_STREAM_RULES_PANEL_MENU_OPEN) !== 0) drawSharedMenu(c);
}

function drawRows(c: CanvasRenderingContext2D): void {
  const kinds = columnBytes(COL_STREAM_RULES_PANEL_ROW_KIND);
  const x = readF32Run(COL_STREAM_RULES_PANEL_ROW_X);
  const y = readF32Run(COL_STREAM_RULES_PANEL_ROW_Y);
  const w = readF32Run(COL_STREAM_RULES_PANEL_ROW_W);
  const h = readF32Run(COL_STREAM_RULES_PANEL_ROW_H);
  const textData = readText(COL_STREAM_RULES_PANEL_ROW_TEXT_DATA);
  const textLen = readU32Run(COL_STREAM_RULES_PANEL_ROW_TEXT_LEN);
  const glyphData = readText(COL_STREAM_RULES_PANEL_ROW_GLYPH_DATA);
  const glyphLen = readU32Run(COL_STREAM_RULES_PANEL_ROW_GLYPH_LEN);
  const free = columnBytes(COL_STREAM_RULES_PANEL_ROW_FREE);
  const nodeRows = readI32Run(COL_STREAM_RULES_PANEL_ROW_NODE_ROW);
  const edgeRows = readI32Run(COL_STREAM_RULES_PANEL_ROW_EDGE_ROW);
  const checks = columnBytes(COL_STREAM_RULES_PANEL_ROW_CHECK);
  const ckx = readF32Run(COL_STREAM_RULES_PANEL_ROW_CHECK_X);
  const cky = readF32Run(COL_STREAM_RULES_PANEL_ROW_CHECK_Y);
  const ckw = readF32Run(COL_STREAM_RULES_PANEL_ROW_CHECK_W);
  const ckh = readF32Run(COL_STREAM_RULES_PANEL_ROW_CHECK_H);
  const values = readU32Run(COL_STREAM_RULES_PANEL_ROW_VALUE);
  const vx = readF32Run(COL_STREAM_RULES_PANEL_ROW_VALUE_X);
  const vy = readF32Run(COL_STREAM_RULES_PANEL_ROW_VALUE_Y);
  const shx = readF32Run(COL_STREAM_RULES_PANEL_ROW_SHARED_X);
  const shy = readF32Run(COL_STREAM_RULES_PANEL_ROW_SHARED_Y);
  const shw = readF32Run(COL_STREAM_RULES_PANEL_ROW_SHARED_W);
  const shh = readF32Run(COL_STREAM_RULES_PANEL_ROW_SHARED_H);
  if (!kinds || !x || !y || !w || !h || !textData || !textLen || !glyphData || !glyphLen) return;
  if (!free || !nodeRows || !edgeRows || !checks || !ckx || !cky || !ckw || !ckh) return;
  if (!values || !vx || !vy || !shx || !shy || !shw || !shh) return;

  let textOff = 0, glyphOff = 0;
  for (let i = 0; i < x.length; i++) {
    const text = decodeAt(textData, textOff, textLen[i]!);
    textOff += textLen[i]!;
    const glyph = decodeAt(glyphData, glyphOff, glyphLen[i]!);
    glyphOff += glyphLen[i]!;

    const kind = kinds.getUint8(i);
    const nodeRow = nodeRows[i]!;
    const edgeRow = edgeRows[i]!;
    const check = checks.getUint8(i);
    const mid = y[i]! + h[i]! / 2;

    const on = check === CHECK_NONE ? true : checkValue(check, nodeRow, edgeRow);
    c.globalAlpha = check !== CHECK_NONE && !on ? INACTIVE_ALPHA : 1;
    c.textBaseline = "middle";
    c.textAlign = "left";

    if (check !== CHECK_NONE) {
      drawCheckbox(c, ckx[i]!, cky[i]!, ckw[i]!, ckh[i]!, on);
    }

    if (kind === ROW_NODE_HEAD) {
      const left = x[i]! + ckw[i]! + 6;
      c.fillStyle = NODE_INK;
      c.font = panelFont(FONT_PX, 600);
      c.fillText(text, left, mid);
      const nameW = c.measureText(text).width;
      if (glyph) {
        c.fillStyle = MUTED;
        c.font = panelFont(FONT_PX);
        c.fillText(`· ${glyph}`, left + nameW + 6, mid);
      }
      const size = nodeRuleGroup(nodeRow).size;
      if (size > 1) {
        roundRect(c, shx[i]!, shy[i]! + 1, shw[i]!, shh[i]! - 2, 8);
        c.fillStyle = PILL_BG;
        c.fill();
        c.strokeStyle = PILL_EDGE;
        c.lineWidth = 1;
        c.stroke();
        c.fillStyle = PILL_INK;
        c.font = panelFont(10);
        c.textAlign = "center";
        c.fillText(`⇄ shared ×${size}`, shx[i]! + shw[i]! / 2, mid);
      }
      c.globalAlpha = 1;
      continue;
    }

    if (kind === ROW_HOLDER_HEAD) {
      c.fillStyle = MUTED;
      c.font = panelFont(FONT_PX);
      c.fillText(text, x[i]! + ckw[i]! + 6, mid);
      c.globalAlpha = 1;
      continue;
    }

    if (kind === ROW_LINE) {
      const value = values[i]!;
      const shown = value === VAL_NONE ? { text, free: free.getUint8(i) !== 0 } : valueText(value, nodeRow, edgeRow);
      c.fillStyle = shown.free ? FREE_GLYPH : GLYPH_INK;
      c.font = panelFont(FONT_PX);
      c.fillText(glyph, x[i]!, mid);
      if (!isRowEditing(i)) {
        c.fillStyle = shown.free ? FREE_INK : INK;
        c.fillText(shown.text, vx[i]!, vy[i]! + h[i]! / 2);
      }
      c.globalAlpha = 1;
      continue;
    }

    c.fillStyle = free.getUint8(i) !== 0 ? FREE_INK : "#444";
    c.font = panelFont(FONT_PX);
    c.fillText(text, x[i]!, mid);
    c.globalAlpha = 1;
  }
}

export { rulesDraftOpen } from "./draw-theta-draft";
