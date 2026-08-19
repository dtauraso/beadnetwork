export function drawScrollHint(
  c: CanvasRenderingContext2D,
  x: number, y: number, w: number, h: number,
  scroll: number, maxScroll: number,
  background: string,
  fade = 12,
): void {
  if (maxScroll <= 0) return;

  if (scroll > 0.5) {
    const g = c.createLinearGradient(0, y, 0, y + fade);
    g.addColorStop(0, background);
    g.addColorStop(1, "rgba(0,0,0,0)");
    c.fillStyle = g;
    c.fillRect(x + 1, y, w - 2, fade);
  }
  if (scroll < maxScroll - 0.5) {
    const g = c.createLinearGradient(0, y + h - fade, 0, y + h);
    g.addColorStop(0, "rgba(0,0,0,0)");
    g.addColorStop(1, background);
    c.fillStyle = g;
    c.fillRect(x + 1, y + h - fade, w - 2, fade);
  }
}
