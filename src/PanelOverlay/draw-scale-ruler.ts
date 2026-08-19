// PLACEMENT: src/PanelOverlay/ — drawn by PanelOverlay onto the same surface as the panels,

const RULER_CSS = 200;
const RULER_H = 28;
const RULER_X = 12;
const RULER_Y = 300;

export function drawScaleRuler(c: CanvasRenderingContext2D, viewCssW: number, viewCssH: number): void {
  c.save();
  c.fillStyle = "rgba(255,0,0,0.85)";
  c.fillRect(RULER_X, RULER_Y, RULER_CSS, RULER_H);

  c.fillStyle = "#000";
  c.fillRect(RULER_X, RULER_Y, 2, RULER_H);
  c.fillRect(RULER_X + RULER_CSS - 2, RULER_Y, 2, RULER_H);

  c.fillStyle = "#fff";
  c.font = "bold 14px -apple-system, system-ui, sans-serif";
  c.textAlign = "left";
  c.textBaseline = "middle";
  c.fillText(`${RULER_CSS}css of ${viewCssW}x${viewCssH}`, RULER_X + 6, RULER_Y + RULER_H / 2);
  c.restore();
}
