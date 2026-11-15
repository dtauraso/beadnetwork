import { canvasFont } from "../../canvas-box";

type R = { x: number; y: number; w: number; h: number };
const TURN = 12;

export type DialLook = { axis: number; start: number; turn: number; ghost: number; inc: number; labels: boolean };

const DIAL_TRACK = "#dde1e6", DIAL_START = "#9aa0a8", DIAL_INK = "#222";

export function drawDial(c: CanvasRenderingContext2D, r: R, d: DialLook, ink: string): void {
  const size = Math.min(r.w, r.h), cx = r.x + r.w / 2, cy = r.y + r.h / 2;
  const insp = size >= 50, mini = insp ? size < 34 : size < 30;
  const rad = size / 2 - (d.labels ? 13 : mini ? 2 : insp ? 5 : 4);
  if (rad <= 6) return;
  const at = (u: number, rr = rad): [number, number] => {
    const t = (u / TURN) * 2 * Math.PI;
    return [cx + rr * Math.sin(t), cy - rr * Math.cos(t)];
  };
  const arc = (from: number, span: number, rr: number): void => {
    if (span === 0) return;
    const s = Math.max(-11.999, Math.min(11.999, span));
    const a0 = (from / TURN) * 2 * Math.PI - Math.PI / 2;
    c.beginPath();
    c.arc(cx, cy, rr, a0, a0 + (s / TURN) * 2 * Math.PI, s < 0);
    c.stroke();
  };
  const dot = (x: number, y: number, rr: number): void => {
    c.beginPath();
    c.arc(x, y, rr, 0, 2 * Math.PI);
  };
  c.setLineDash([]);
  c.lineCap = "butt";
  c.strokeStyle = DIAL_TRACK;
  c.lineWidth = insp ? (mini ? 1.5 : 1) : 1.5;
  dot(cx, cy, rad);
  c.stroke();
  if (!mini) {
    for (let u = 0; u < TURN; u += insp ? 1 : 3) {
      const major = u % 3 === 0, o = insp ? (major ? 4 : 2.5) : 0;
      const [x0, y0] = insp ? at(u, rad - o) : at(u, rad - 3), [x1, y1] = insp ? at(u, rad + o) : at(u, rad + 2);
      c.strokeStyle = major ? DIAL_START : DIAL_TRACK;
      c.lineWidth = insp && major ? 1.3 : 1;
      c.beginPath();
      c.moveTo(x0, y0);
      c.lineTo(x1, y1);
      c.stroke();
    }
  }
  if (d.labels) {
    c.fillStyle = DIAL_START;
    c.font = canvasFont(9);
    c.textAlign = "center";
    c.textBaseline = "middle";
    for (const u of [0, 3, 6, 9]) {
      const [x, y] = at(u, rad + 9);
      c.fillText(String(u), x, y);
    }
  }
  const w = mini ? 2.5 : insp ? (size > 100 ? 5 : 3.5) : 3;
  c.lineCap = "round";
  if (d.start) {
    c.strokeStyle = DIAL_START;
    c.lineWidth = insp && !mini ? 2 : 1.5;
    arc(0, d.start, rad - (mini ? 0 : w + 1));
  }
  c.strokeStyle = ink;
  c.lineWidth = w;
  if (d.turn) {
    if (d.turn < 0) c.setLineDash([3, 2.5]);
    arc(d.start, d.turn, rad);
    c.setLineDash([]);
  }
  if (d.ghost) {
    c.globalAlpha = 0.45;
    c.setLineDash([2, 3]);
    arc(d.start + d.turn, d.ghost, rad);
    c.setLineDash([]);
    c.globalAlpha = 1;
  }
  if (d.inc > 0 && d.turn && !mini) {
    const n = Math.abs(d.turn / d.inc);
    c.fillStyle = DIAL_INK;
    c.globalAlpha = 0.9;
    for (let k = 1; k < n && k < 48; k++) {
      const [x, y] = at(d.start + Math.sign(d.turn) * d.inc * k);
      dot(x, y, insp ? 1.3 : 1.2);
      c.fill();
    }
    c.globalAlpha = 1;
  }
  const end = d.start + d.turn;
  const [ex, ey] = at(end, rad * (insp ? (mini ? 0.95 : 0.78) : 0.8));
  c.lineCap = "butt";
  c.strokeStyle = DIAL_INK;
  c.lineWidth = insp ? (mini ? 1.2 : 1.4) : 1.2;
  c.globalAlpha = 0.55;
  c.beginPath();
  c.moveTo(cx, cy);
  c.lineTo(ex, ey);
  c.stroke();
  c.globalAlpha = 0.6;
  c.fillStyle = DIAL_INK;
  dot(cx, cy, insp ? (mini ? 1.2 : 2) : 1.3);
  c.fill();
  c.globalAlpha = 1;
  if (insp && !mini && d.start) {
    const [sx, sy] = at(d.start);
    c.fillStyle = DIAL_START;
    dot(sx, sy, 2.6);
    c.fill();
  }
  const endR = mini ? 2.2 : insp ? (size > 100 ? 6 : 4) : 3.2;
  const [hx, hy] = at(end);
  c.fillStyle = ink;
  dot(hx, hy, endR);
  c.fill();
  if (d.ghost) {
    const [gx, gy] = at(end + d.ghost);
    c.globalAlpha = 0.7;
    c.strokeStyle = ink;
    c.lineWidth = 1.5;
    dot(gx, gy, insp ? endR : 3.2);
    c.stroke();
    c.globalAlpha = 1;
  }
}
