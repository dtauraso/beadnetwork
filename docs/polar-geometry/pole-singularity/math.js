export const R = 1, TAU = Math.PI * 2, D2R = Math.PI / 180, R2D = 180 / Math.PI;
export const R_WORLD = 10, SP_DEG = 15, SP = SP_DEG * D2R;   // illustrative φ grid cell
export const cssv = n => getComputedStyle(document.documentElement).getPropertyValue(n).trim();
export const isDark = () => (document.documentElement.getAttribute("data-theme") || (matchMedia("(prefers-color-scheme:dark)").matches ? "dark" : "light")) === "dark";
export const clamp = (v, a, b) => v < a ? a : v > b ? b : v;

// ---- camera (shared orbit) ----
export const cam = { az: 0.7, el: 0.42 };
export function rot(p) { // world -> camera space (orthographic). depth>0 = toward viewer
  let x = p.x * Math.cos(cam.az) + p.z * Math.sin(cam.az);
  let z = -p.x * Math.sin(cam.az) + p.z * Math.cos(cam.az);
  let y = p.y;
  let y2 = y * Math.cos(cam.el) - z * Math.sin(cam.el);
  let z2 = y * Math.sin(cam.el) + z * Math.cos(cam.el);
  return { x, y: y2, d: z2 };
}
export function unrot(X, Y, Z) { // camera space -> world
  let y = Y * Math.cos(cam.el) + Z * Math.sin(cam.el);
  let z1 = -Y * Math.sin(cam.el) + Z * Math.cos(cam.el);
  let x = X * Math.cos(cam.az) - z1 * Math.sin(cam.az);
  let z = X * Math.sin(cam.az) + z1 * Math.cos(cam.az);
  return { x, y, z };
}
export const sph = (th, ph) => ({ x: R * Math.sin(th) * Math.cos(ph), y: R * Math.cos(th), z: R * Math.sin(th) * Math.sin(ph) });

// ---- a tiny scene rig per canvas ----
export function rig(cv) {
  const g = cv.getContext("2d"); let w, h, cx, cy, S;
  function size() {
    const dpr = Math.min(devicePixelRatio || 1, 2); const cw = cv.clientWidth || 600; const ch = Math.round(cw * 0.875);
    cv.width = Math.round(cw * dpr); cv.height = Math.round(ch * dpr); g.setTransform(dpr, 0, 0, dpr, 0, 0);
    w = cw; h = ch; cx = w * 0.46; cy = h * 0.5; S = Math.min(w, h) * 0.40;
  }
  const scr = p => { const c = rot(p); return { x: cx + c.x * S, y: cy - c.y * S, d: c.d, front: c.d >= 0 }; };
  return { g, size, scr, get w() { return w }, get h() { return h }, get cx() { return cx }, get cy() { return cy }, get S() { return S } };
}

export function ball(r, cx, cy, S) { // shaded sphere fill
  const R2 = R * S, lx = cx - R2 * 0.4, ly = cy - R2 * 0.45;
  const grd = r.g.createRadialGradient(lx, ly, R2 * 0.1, cx, cy, R2 * 1.15);
  if (isDark()) { grd.addColorStop(0, "#2b313d"); grd.addColorStop(.55, "#1a1e27"); grd.addColorStop(1, "#0c0e13"); }
  else { grd.addColorStop(0, "#ffffff"); grd.addColorStop(.6, "#e9e5da"); grd.addColorStop(1, "#cfcabb"); }
  // soft ground shadow
  r.g.save(); r.g.globalAlpha = isDark() ? 0.35 : 0.16; r.g.fillStyle = "#000";
  r.g.beginPath(); r.g.ellipse(cx, cy + R2 * 0.98, R2 * 0.82, R2 * 0.16, 0, 0, TAU); r.g.fill(); r.g.restore();
  r.g.beginPath(); r.g.arc(cx, cy, R2, 0, TAU); r.g.fillStyle = grd; r.g.fill();
  r.g.lineWidth = 1; r.g.strokeStyle = isDark() ? "rgba(120,130,150,.25)" : "rgba(90,90,110,.18)"; r.g.stroke();
}
// draw a 3D polyline with front/back depth alpha
export function path3(r, pts, col, wFront, aBack) {
  for (let i = 1; i < pts.length; i++) {
    const a = r.scr(pts[i - 1]), b = r.scr(pts[i]);
    const front = (a.d + b.d) >= 0;
    r.g.globalAlpha = front ? 1 : aBack; r.g.lineWidth = front ? wFront : wFront * 0.8;
    r.g.strokeStyle = col; r.g.beginPath(); r.g.moveTo(a.x, a.y); r.g.lineTo(b.x, b.y); r.g.stroke();
  }
  r.g.globalAlpha = 1;
}
// dashed 3D polyline (construction lines), front/back depth alpha
export function dashPath3(r, pts, col, wFront, aBack) {
  r.g.save(); r.g.setLineDash([5, 5]);
  for (let i = 1; i < pts.length; i++) {
    const a = r.scr(pts[i - 1]), b = r.scr(pts[i]); const front = (a.d + b.d) >= 0;
    r.g.globalAlpha = front ? 0.95 : aBack; r.g.lineWidth = wFront; r.g.strokeStyle = col;
    r.g.beginPath(); r.g.moveTo(a.x, a.y); r.g.lineTo(b.x, b.y); r.g.stroke();
  }
  r.g.restore(); r.g.globalAlpha = 1;
}
export function dot3(r, p, rad, col, ring) {
  const s = r.scr(p);
  r.g.globalAlpha = s.front ? 1 : 0.32; r.g.beginPath(); r.g.arc(s.x, s.y, rad, 0, TAU); r.g.fillStyle = col; r.g.fill();
  if (ring) { r.g.lineWidth = 2; r.g.strokeStyle = ring; r.g.stroke(); } r.g.globalAlpha = 1; return s;
}
export function arrowHead(r, pA, pB, col) {
  const a = r.scr(pA), b = r.scr(pB); if (!b.front) return;
  const ang = Math.atan2(b.y - a.y, b.x - a.x), L = 8.5;
  r.g.globalAlpha = 1; r.g.fillStyle = col; r.g.beginPath(); r.g.moveTo(b.x, b.y);
  r.g.lineTo(b.x - L * Math.cos(ang - 0.42), b.y - L * Math.sin(ang - 0.42));
  r.g.lineTo(b.x - L * Math.cos(ang + 0.42), b.y - L * Math.sin(ang + 0.42));
  r.g.closePath(); r.g.fill();
}
export function label3(r, p, txt, col, dx, dy) {
  const s = r.scr(p); if (!s.front) return;
  r.g.globalAlpha = 1; r.g.font = "12px " + cssv("--mono"); r.g.fillStyle = col; r.g.textAlign = "left"; r.g.textBaseline = "middle";
  r.g.fillText(txt, s.x + (dx || 8), s.y + (dy || 0));
}
export const meridian = ph => { const a = []; for (let i = 0; i <= 64; i++)a.push(sph(i / 64 * Math.PI, ph)); return a; };
export const parallel = th => { const a = []; for (let i = 0; i <= 72; i++)a.push(sph(th, i / 72 * TAU)); return a; };

// ---- unproject screen -> sphere front point -> (theta,phi) ----
// unproject to the sphere. Inside the silhouette (ρ≤1) → near hemisphere. Past it (ρ>1) we
// reflect: effective radius 2−ρ on the FAR hemisphere — so dragging outward from centre sweeps
// front-pole → equator → back-pole in ONE motion, all the way around, no state.
export function pickHemi(r, mx, my) {
  let X = (mx - r.cx) / r.S, Y = -(my - r.cy) / r.S; const rho = Math.hypot(X, Y);
  let back = false;
  if (rho > 1) { back = true; const rr = Math.max(0, 2 - rho), s = rho > 0 ? rr / rho : 0; X *= s; Y *= s; }
  const q = X * X + Y * Y, Zm = Math.sqrt(Math.max(0, 1 - q)), Z = back ? -Zm : Zm;
  const p = unrot(X, Y, Z); const th = Math.acos(clamp(p.y, -1, 1)); const ph = Math.atan2(p.z, p.x);
  return { th: clamp(th, 0.6 * D2R, Math.PI - 0.6 * D2R), ph };
}

// ---- vector algebra shared by tab 2's great-circle construction ----
export function norm(v) { const m = Math.hypot(v.x, v.y, v.z) || 1; return { x: v.x / m, y: v.y / m, z: v.z / m }; }
export function cross(a, b) { return { x: a.y * b.z - a.z * b.y, y: a.z * b.x - a.x * b.z, z: a.x * b.y - a.y * b.x }; }
export function dot(a, b) { return a.x * b.x + a.y * b.y + a.z * b.z; }
export const vsub = (a, b) => ({ x: a.x - b.x, y: a.y - b.y, z: a.z - b.z });
export const vscale = (a, s) => ({ x: a.x * s, y: a.y * s, z: a.z * s });
export function rotAbout(v, k, ang) {
  const c = Math.cos(ang), s = Math.sin(ang), kv = dot(k, v), cr = cross(k, v);
  return { x: v.x * c + cr.x * s + k.x * kv * (1 - c), y: v.y * c + cr.y * s + k.y * kv * (1 - c), z: v.z * c + cr.z * s + k.z * kv * (1 - c) };
}

export function setTxt(id, t) { document.getElementById(id).textContent = t; }

// ---- interaction: shared across both canvases' bind() (one drag flag, not one per tab) ----
export const interaction = { dragging: false };
