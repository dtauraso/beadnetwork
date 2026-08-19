import { columnF32, columnI32 } from "../Buffer/column-values";
import { panelFont, roundRect } from "../PanelOverlay/panel-box";
import { readF32Run, readI32Run, readU32Run, readText, decodeAt } from "../PanelOverlay/panel-columns";
import {
  nodeRuleGroup, nodeDragActive, drawCheckbox,
  INK, PILL_EDGE, PILL_INK, INACTIVE_ALPHA, HEAD_FONT_PX,
} from "./rules-values";
import {
  COL_STREAM_RULES_PANEL_MENU_ANCHOR_ROW, COL_STREAM_RULES_PANEL_MENU_X,
  COL_STREAM_RULES_PANEL_MENU_Y, COL_STREAM_RULES_PANEL_MENU_W,
  COL_STREAM_RULES_PANEL_MENU_H, COL_STREAM_RULES_PANEL_MENU_ROW_X,
  COL_STREAM_RULES_PANEL_MENU_ROW_Y, COL_STREAM_RULES_PANEL_MENU_ROW_W,
  COL_STREAM_RULES_PANEL_MENU_ROW_H, COL_STREAM_RULES_PANEL_MENU_CHECK_X,
  COL_STREAM_RULES_PANEL_MENU_CHECK_Y, COL_STREAM_RULES_PANEL_MENU_LABEL_DATA,
  COL_STREAM_RULES_PANEL_MENU_LABEL_LEN, COL_STREAM_RULES_PANEL_MENU_NODE_ROW,
} from "./columns-gen";

export function drawSharedMenu(c: CanvasRenderingContext2D): void {
  const w = columnF32(COL_STREAM_RULES_PANEL_MENU_W);
  const h = columnF32(COL_STREAM_RULES_PANEL_MENU_H);
  if (w <= 0 || h <= 0) return;
  const x = columnF32(COL_STREAM_RULES_PANEL_MENU_X);
  const y = columnF32(COL_STREAM_RULES_PANEL_MENU_Y);

  roundRect(c, x + 0.5, y + 0.5, w - 1, h - 1, 6);
  c.fillStyle = "#fff";
  c.fill();
  c.strokeStyle = PILL_EDGE;
  c.lineWidth = 1;
  c.stroke();

  const anchorRow = columnI32(COL_STREAM_RULES_PANEL_MENU_ANCHOR_ROW);
  const group = anchorRow >= 0 ? nodeRuleGroup(anchorRow) : { id: -1, size: 0 };

  c.fillStyle = PILL_INK;
  c.font = panelFont(HEAD_FONT_PX);
  c.textAlign = "left";
  c.textBaseline = "middle";
  c.fillText(`shared by ${group.size}`, x + 6, y + 10);

  const rx = readF32Run(COL_STREAM_RULES_PANEL_MENU_ROW_X);
  const ry = readF32Run(COL_STREAM_RULES_PANEL_MENU_ROW_Y);
  const rw = readF32Run(COL_STREAM_RULES_PANEL_MENU_ROW_W);
  const rh = readF32Run(COL_STREAM_RULES_PANEL_MENU_ROW_H);
  const cx = readF32Run(COL_STREAM_RULES_PANEL_MENU_CHECK_X);
  const cy = readF32Run(COL_STREAM_RULES_PANEL_MENU_CHECK_Y);
  const labelData = readText(COL_STREAM_RULES_PANEL_MENU_LABEL_DATA);
  const labelLen = readU32Run(COL_STREAM_RULES_PANEL_MENU_LABEL_LEN);
  const rows = readI32Run(COL_STREAM_RULES_PANEL_MENU_NODE_ROW);
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
    c.font = panelFont(HEAD_FONT_PX, row === anchorRow ? 600 : undefined);
    c.textAlign = "left";
    c.fillText(label, cx[i]! + 13 + 6, ry[i]! + rh[i]! / 2);
    c.globalAlpha = 1;
  }
}
