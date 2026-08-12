import {
  sph, meridian, parallel, path3, dot3,
  clamp, rig, ball,
  norm, cross, dot, vsub, vscale, rotAbout, cssv, D2R,
} from "./math.js";
import {
  drawTab2Cells, drawTab2Arc, drawTab2Degenerate, drawTab2Pucks, drawTab2Psi, drawTab2Instruments,
} from "./tab2-render.js";

export const r2 = rig(document.getElementById("c2"));
export const ref = { th: 78 * D2R, ph: -1.9 };
export const tgt = { th: 9 * D2R, ph: 0.7 };
let showCells = true;
export function setShowCells(v) { showCells = v; }
function vec(o) { return sph(o.th, o.ph); }
let refE1 = null, refPrevU = null;

export function draw2() {
  const r = r2; r.g.clearRect(0, 0, r.w, r.h); ball(r, r.cx, r.cy, r.S);
  for (let a = 0; a < 360; a += 30) path3(r, meridian(a * D2R), cssv("--amber-soft"), 1, 0.07);
  for (let d = 30; d <= 150; d += 30) path3(r, parallel(d * D2R), cssv("--grid"), 1, 0.07);
  dot3(r, sph(0, 0), 3, cssv("--muted"));

  {
    const tt = norm(vec(tgt)); let uu = norm(vec(ref)); const MIN = 15 * D2R;
    const a = Math.acos(clamp(dot(uu, tt), -1, 1)); const d = clamp(a, MIN, Math.PI - MIN);
    let tan = { x: uu.x - dot(uu, tt) * tt.x, y: uu.y - dot(uu, tt) * tt.y, z: uu.z - dot(uu, tt) * tt.z };
    let tm = Math.hypot(tan.x, tan.y, tan.z);
    if (tm < 1e-4) { const hlp = Math.abs(tt.y) < 0.9 ? { x: 0, y: 1, z: 0 } : { x: 1, y: 0, z: 0 }; tan = cross(tt, hlp); tm = Math.hypot(tan.x, tan.y, tan.z); }
    tan = { x: tan.x / tm, y: tan.y / tm, z: tan.z / tm };
    const nr = { x: Math.cos(d) * tt.x + Math.sin(d) * tan.x, y: Math.cos(d) * tt.y + Math.sin(d) * tan.y, z: Math.cos(d) * tt.z + Math.sin(d) * tan.z };
    ref.th = Math.acos(clamp(nr.y, -1, 1)); ref.ph = Math.atan2(nr.z, nr.x);
  }
  const u = norm(vec(ref)), t = norm(vec(tgt));
  const ang = Math.acos(clamp(dot(u, t), -1, 1));
  const deg = Math.min(ang, Math.PI - ang);
  const EPS = 9 * D2R;
  const fade = clamp((deg - EPS) / (15 * D2R), 0, 1);

  if (!refE1) {
    let up = Math.abs(u.y) < 0.9 ? { x: 0, y: 1, z: 0 } : { x: 1, y: 0, z: 0 };
    refE1 = norm(vsub(up, vscale(u, dot(up, u)))); refPrevU = { x: u.x, y: u.y, z: u.z };
  }
  else {
    const ax = cross(refPrevU, u), axm = Math.hypot(ax.x, ax.y, ax.z);
    if (axm > 1e-6) {
      const a2 = Math.acos(clamp(dot(refPrevU, u), -1, 1));
      refE1 = rotAbout(refE1, { x: ax.x / axm, y: ax.y / axm, z: ax.z / axm }, a2);
    }
    refE1 = norm(vsub(refE1, vscale(u, dot(refE1, u)))); refPrevU = { x: u.x, y: u.y, z: u.z };
  }
  const e1 = refE1, e2 = cross(u, e1);
  const psiR = Math.atan2(dot(t, e2), dot(t, e1));
  const P = (c, a) => ({
    x: Math.cos(c) * u.x + Math.sin(c) * (Math.cos(a) * e1.x + Math.sin(a) * e2.x),
    y: Math.cos(c) * u.y + Math.sin(c) * (Math.cos(a) * e1.y + Math.sin(a) * e2.y),
    z: Math.cos(c) * u.z + Math.sin(c) * (Math.cos(a) * e1.z + Math.sin(a) * e2.z)
  });

  drawTab2Cells(r, P, showCells, ang, psiR);
  drawTab2Arc(r, u, t, ang, fade);
  drawTab2Degenerate(r, t, fade);
  drawTab2Pucks(r, u, t);
  drawTab2Psi(r, P, e1, e2, psiR, fade);
  drawTab2Instruments(tgt, ang, deg, EPS, psiR, fade);
}
