import { canvasFont, roundRect } from "../../canvas-box";
import { decodeAt } from "../../leaf-text";
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
  rulesBytes, rulesF32, rulesU8, rulesF32Run, rulesI32Run, rulesU32Run, rulesText,
} from "./rules-leaves";

export function rulesPanelKey(): string {
  const nodeRows = rulesI32Run("rowNodeRow");
  const values = rulesU32Run("rowValue");
  const edgeRows = rulesI32Run("rowEdgeRow");
  const checks = rulesBytes("rowCheck");
  const parts: string[] = [
    String(rulesU8("open")),
    String(rulesF32("boxH")),
    String(rulesF32("scrollY")),
    String(rulesU8("menuOpen")),
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
  const boxW = rulesF32("boxW");
  const boxH = rulesF32("boxH");
  const toggleText = rulesText("toggleText");
  if (boxW <= 0 || boxH <= 0 || !toggleText) return;

  roundRect(
    c,
    rulesF32("boxX") + 0.5, rulesF32("boxY") + 0.5,
    boxW - 1, boxH - 1, 6,
  );
  c.fillStyle = PANEL_BG;
  c.fill();
  c.strokeStyle = PANEL_EDGE;
  c.lineWidth = 1;
  c.stroke();

  c.fillStyle = MUTED;
  c.font = canvasFont(HEAD_FONT_PX);
  c.textAlign = "left";
  c.textBaseline = "middle";
  c.fillText(
    decodeAt(toggleText, 0, toggleText.length),
    rulesF32("toggleX"),
    rulesF32("toggleY") + rulesF32("toggleH") / 2,
  );

  if (rulesU8("open") === 0) return;

  const clipY = rulesF32("clipY");
  const clipH = rulesF32("clipH");
  const boxX = rulesF32("boxX");
  c.save();
  c.beginPath();
  c.rect(boxX, clipY, boxW, clipH);
  c.clip();
  drawRows(c);
  drawThetaDraft(c);
  c.restore();

  if (rulesU8("menuOpen") !== 0) drawSharedMenu(c);
}

function drawRows(c: CanvasRenderingContext2D): void {
  const kinds = rulesBytes("rowKind");
  const x = rulesF32Run("rowX");
  const y = rulesF32Run("rowY");
  const w = rulesF32Run("rowW");
  const h = rulesF32Run("rowH");
  const textData = rulesText("rowTextData");
  const textLen = rulesU32Run("rowTextLen");
  const glyphData = rulesText("rowGlyphData");
  const glyphLen = rulesU32Run("rowGlyphLen");
  const free = rulesBytes("rowFree");
  const nodeRows = rulesI32Run("rowNodeRow");
  const edgeRows = rulesI32Run("rowEdgeRow");
  const checks = rulesBytes("rowCheck");
  const ckx = rulesF32Run("rowCheckX");
  const cky = rulesF32Run("rowCheckY");
  const ckw = rulesF32Run("rowCheckW");
  const ckh = rulesF32Run("rowCheckH");
  const values = rulesU32Run("rowValue");
  const vx = rulesF32Run("rowValueX");
  const vy = rulesF32Run("rowValueY");
  const shx = rulesF32Run("rowSharedX");
  const shy = rulesF32Run("rowSharedY");
  const shw = rulesF32Run("rowSharedW");
  const shh = rulesF32Run("rowSharedH");
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
      c.font = canvasFont(FONT_PX, 600);
      c.fillText(text, left, mid);
      const nameW = c.measureText(text).width;
      if (glyph) {
        c.fillStyle = MUTED;
        c.font = canvasFont(FONT_PX);
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
        c.font = canvasFont(10);
        c.textAlign = "center";
        c.fillText(`⇄ shared ×${size}`, shx[i]! + shw[i]! / 2, mid);
      }
      c.globalAlpha = 1;
      continue;
    }

    if (kind === ROW_HOLDER_HEAD) {
      c.fillStyle = MUTED;
      c.font = canvasFont(FONT_PX);
      c.fillText(text, x[i]! + ckw[i]! + 6, mid);
      c.globalAlpha = 1;
      continue;
    }

    if (kind === ROW_LINE) {
      const value = values[i]!;
      const shown = value === VAL_NONE ? { text, free: free.getUint8(i) !== 0 } : valueText(value, nodeRow, edgeRow);
      c.fillStyle = shown.free ? FREE_GLYPH : GLYPH_INK;
      c.font = canvasFont(FONT_PX);
      c.fillText(glyph, x[i]!, mid);
      if (!isRowEditing(i)) {
        c.fillStyle = shown.free ? FREE_INK : INK;
        c.fillText(shown.text, vx[i]!, vy[i]! + h[i]! / 2);
      }
      c.globalAlpha = 1;
      continue;
    }

    c.fillStyle = free.getUint8(i) !== 0 ? FREE_INK : "#444";
    c.font = canvasFont(FONT_PX);
    c.fillText(text, x[i]!, mid);
    c.globalAlpha = 1;
  }
}

export { rulesDraftOpen } from "./draw-theta-draft";
