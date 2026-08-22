
export const BOX_FILL = "#fff";
export const BOX_EDGE = "#ddd";
export const BOX_RADIUS = 6;

export const CANVAS_FONT = "ui-sans-serif, system-ui, sans-serif";

export function canvasFont(px: number, weight?: number | "bold"): string {
  return `${weight ? `${weight} ` : ""}${px}px ${CANVAS_FONT}`;
}

export function roundRect(
  c: CanvasRenderingContext2D,
  x: number,
  y: number,
  w: number,
  h: number,
  r: number,
): void {
  const rr = Math.min(r, w / 2, h / 2);
  c.beginPath();
  c.moveTo(x + rr, y);
  c.arcTo(x + w, y, x + w, y + h, rr);
  c.arcTo(x + w, y + h, x, y + h, rr);
  c.arcTo(x, y + h, x, y, rr);
  c.arcTo(x, y, x + w, y, rr);
  c.closePath();
}

export function drawBox(
  c: CanvasRenderingContext2D,
  x: number,
  y: number,
  w: number,
  h: number,
): void {
  if (w <= 0 || h <= 0) return;
  roundRect(c, x + 0.5, y + 0.5, w - 1, h - 1, BOX_RADIUS);
  c.fillStyle = BOX_FILL;
  c.fill();
  c.strokeStyle = BOX_EDGE;
  c.lineWidth = 1;
  c.stroke();
}
