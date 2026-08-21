import { canvasFont, roundRect } from "../../../webview/canvas-box";
import { decodeAt } from "../../../webview/leaf-text";
import {
  nodeRuleGroup, nodeDragActive, drawCheckbox,
  INK, PILL_EDGE, PILL_INK, INACTIVE_ALPHA, HEAD_FONT_PX,
} from "./rules-values";
import {
  rulesF32, rulesI32, rulesF32Run, rulesI32Run, rulesU32Run, rulesText,
} from "./rules-leaves";

export function drawSharedMenu(c: CanvasRenderingContext2D): void {
  const w = rulesF32("menuW");
  const h = rulesF32("menuH");
  if (w <= 0 || h <= 0) return;
  const x = rulesF32("menuX");
  const y = rulesF32("menuY");

  roundRect(c, x + 0.5, y + 0.5, w - 1, h - 1, 6);
  c.fillStyle = "#fff";
  c.fill();
  c.strokeStyle = PILL_EDGE;
  c.lineWidth = 1;
  c.stroke();

  const anchorRow = rulesI32("menuAnchorRow");
  const group = anchorRow >= 0 ? nodeRuleGroup(anchorRow) : { id: -1, size: 0 };

  c.fillStyle = PILL_INK;
  c.font = canvasFont(HEAD_FONT_PX);
  c.textAlign = "left";
  c.textBaseline = "middle";
  c.fillText(`shared by ${group.size}`, x + 6, y + 10);

  const rx = rulesF32Run("menuRowX");
  const ry = rulesF32Run("menuRowY");
  const rw = rulesF32Run("menuRowW");
  const rh = rulesF32Run("menuRowH");
  const cx = rulesF32Run("menuCheckX");
  const cy = rulesF32Run("menuCheckY");
  const labelData = rulesText("menuLabelData");
  const labelLen = rulesU32Run("menuLabelLen");
  const rows = rulesI32Run("menuNodeRow");
  if (!rx || !ry || !rw || !rh || !cx || !cy || !labelData || !labelLen || !rows) return;

  let allOn = true;
  for (const row of rows) {
    if (row >= 0 && nodeRuleGroup(row).id === group.id) {
      if (!nodeDragActive(row)) allOn = false;
    }
  }

  let off = 0;
  for (let i = 0; i < rx.length; i++) {
    const label = decodeAt(labelData, off, labelLen[i]!);
    off += labelLen[i]!;
    const row = rows[i]!;
    const member = row < 0 || nodeRuleGroup(row).id === group.id;
    const on = row < 0 ? allOn : nodeDragActive(row);

    c.globalAlpha = member ? 1 : INACTIVE_ALPHA;
    drawCheckbox(c, cx[i]!, cy[i]!, 13, 13, on);
    c.fillStyle = INK;
    c.font = canvasFont(HEAD_FONT_PX, row === anchorRow ? 600 : undefined);
    c.textAlign = "left";
    c.fillText(label, cx[i]! + 13 + 6, ry[i]! + rh[i]! / 2);
    c.globalAlpha = 1;
  }
}
