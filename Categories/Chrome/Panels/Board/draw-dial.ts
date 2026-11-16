import { canvasFont } from "../../canvas-box";

type R = { x: number; y: number; w: number; h: number };
const TURN = 12;

export type DialLook = { axis: number; start: number; turn: number; inc: number; labels: boolean };

const GUIDE = "#dde1e6", TICK = "#c8ccd2", LABEL = "#555", START = "#9aa0a8", INK = "#222", ACCENT = "#1565c0";

export function drawDial(c: CanvasRenderingContext2D, r: R, d: DialLook, col: string): void {
  const size = Math.min(r.w, r.h), cx = r.x + r.w / 2, cy = r.y + r.h / 2, rad = size / 2 - 1.5;
  if (rad <= 4) return;
  const at = (u: number, rr: number): [number, number] => {
    const t = (u / TURN) * 2 * Math.PI;
    return [cx + rr * Math.sin(t), cy - rr * Math.cos(t)];
  };
  const arc = (from: number, span: number, rr: number): void => {
    const s = Math.max(-11.999, Math.min(11.999, span));
    const a0 = (from / TURN) * 2 * Math.PI - Math.PI / 2;
    c.beginPath();
    c.arc(cx, cy, rr, a0, a0 + (s / TURN) * 2 * Math.PI, s < 0);
    c.stroke();
  };
  const disc = (x: number, y: number, rr: number): void => {
    c.beginPath();
    c.arc(x, y, rr, 0, 2 * Math.PI);
  };
  const hand = (u: number, ink: string, w: number, rr: number): void => {
    const [x, y] = at(u, rr);
    c.strokeStyle = ink;
    c.lineWidth = w;
    c.beginPath();
    c.moveTo(cx, cy);
    c.lineTo(x, y);
    c.stroke();
  };
  c.save();
  c.setLineDash([]);
  disc(cx, cy, rad);
  c.fillStyle = "#fff";
  c.fill();
  c.strokeStyle = GUIDE;
  c.lineWidth = 1;
  c.stroke();
  if (size >= 40) {
    c.strokeStyle = TICK;
    for (let u = 0; u < TURN; u++) {
      const [x0, y0] = at(u, rad), [x1, y1] = at(u, rad - (u % 3 ? 3 : 6));
      c.beginPath();
      c.moveTo(x0, y0);
      c.lineTo(x1, y1);
      c.stroke();
    }
  }
  if (d.labels && size >= 70) {
    c.fillStyle = LABEL;
    c.font = canvasFont(10);
    c.textAlign = "center";
    c.textBaseline = "middle";
    for (const u of [0, 3, 6, 9]) {
      const [x, y] = at(u, rad - 13);
      c.fillText(String(u), x, y);
    }
  }
  const sw = Math.max(2, size / 14), track = rad - sw, reach = rad - sw * 1.5, knob = size / 16;
  const end = d.start + d.turn;
  c.strokeStyle = col;
  c.lineWidth = sw;
  if (d.turn) {
    c.lineCap = "round";
    if (d.turn < 0) c.setLineDash([3, 3]);
    arc(d.start, d.turn, track);
    c.setLineDash([]);
  }
  c.lineCap = "butt";
  hand(d.start, START, Math.max(1.2, size / 30), reach);
  if (d.turn) {
    c.lineCap = "round";
    hand(end, INK, Math.max(1.5, size / 24), reach);
    if (size >= 40) {
      const [x, y] = at(end, reach);
      disc(x, y, knob);
      c.fillStyle = ACCENT;
      c.fill();
    }
  }
  disc(cx, cy, Math.max(1.2, size / 28));
  c.fillStyle = INK;
  c.fill();
  c.restore();
}
