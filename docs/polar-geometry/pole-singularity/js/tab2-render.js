import {
  R, R_WORLD, SP, R2D, D2R, TAU, cssv, path3, dashPath3, dot3,
  arrowHead, label3, clamp, setTxt,
  norm, cross, dot,
} from "./math.js";
import { nudge1 } from "./tab1.js";

export function drawTab2Cells(r, P, showCells, ang, psiR) {
  const SC = 20 * D2R, SPS = 20 * D2R;
  if (!showCells) return;
  for (let c = SC; c < Math.PI - 1e-3; c += SC) { const a = []; for (let i = 0; i <= 64; i++)a.push(P(c, i / 64 * TAU)); path3(r, a, cssv("--teal-soft"), 1, 0.09); }
  for (let b = 0; b < TAU - 1e-3; b += SPS) { const a = []; for (let i = 0; i <= 48; i++)a.push(P(i / 48 * Math.PI, b)); path3(r, a, cssv("--teal-soft"), 1, 0.09); }
  const bp = ((psiR % TAU) + TAU) % TAU, ci = Math.floor(ang / SC) * SC, bi = Math.floor(bp / SPS) * SPS;
  [ci, ci + SC].forEach(c => { if (c <= 0 || c >= Math.PI) return; const a = []; for (let i = 0; i <= 64; i++)a.push(P(c, i / 64 * TAU)); path3(r, a, cssv("--teal"), 1.7, 0.18); });
  [bi, bi + SPS].forEach(b => { const a = []; for (let i = 0; i <= 48; i++)a.push(P(i / 48 * Math.PI, b)); path3(r, a, cssv("--teal"), 1.7, 0.18); });
}

export function drawTab2Arc(r, u, t, ang, fade) {
  let vv = { x: t.x - dot(t, u) * u.x, y: t.y - dot(t, u) * u.y, z: t.z - dot(t, u) * u.z };
  const vm = Math.hypot(vv.x, vv.y, vv.z);
  if (fade <= 0 || vm <= 1e-4) return;
  vv = { x: vv.x / vm, y: vv.y / vm, z: vv.z / vm };
  r.g.globalAlpha = fade;
  const circ = []; for (let i = 0; i <= 96; i++) { const a = i / 96 * TAU; circ.push({ x: u.x * Math.cos(a) + vv.x * Math.sin(a), y: u.y * Math.cos(a) + vv.y * Math.sin(a), z: u.z * Math.cos(a) + vv.z * Math.sin(a) }); }
  path3(r, circ, cssv("--teal-soft"), 1.6, 0.16);
  const arc = [];
  for (let i = 0; i <= 48; i++) { const a = ang * i / 48; arc.push({ x: u.x * Math.cos(a) + vv.x * Math.sin(a), y: u.y * Math.cos(a) + vv.y * Math.sin(a), z: u.z * Math.cos(a) + vv.z * Math.sin(a) }); }
  path3(r, arc, cssv("--teal"), 4.5, 0.4);
  if (arc.length > 2) arrowHead(r, arc[arc.length - 2], arc[arc.length - 1], cssv("--teal"));
  const midc = { x: u.x * Math.cos(ang / 2) + vv.x * Math.sin(ang / 2), y: u.y * Math.cos(ang / 2) + vv.y * Math.sin(ang / 2), z: u.z * Math.cos(ang / 2) + vv.z * Math.sin(ang / 2) };
  label3(r, midc, "c", cssv("--teal"), 8, -6);
  let ax = norm(cross(u, t)); if (ax.y < 0) { ax = { x: -ax.x, y: -ax.y, z: -ax.z }; }
  dashPath3(r, [{ x: -ax.x * R, y: -ax.y * R, z: -ax.z * R }, { x: ax.x * R, y: ax.y * R, z: ax.z * R }], cssv("--muted"), 1.6, 0.3);
  dot3(r, ax, 3.5, cssv("--muted")); dot3(r, { x: -ax.x, y: -ax.y, z: -ax.z }, 3.5, cssv("--muted"));
  label3(r, ax, "axis", cssv("--muted"), 9, -2);
  r.g.globalAlpha = 1;
}

export function drawTab2Degenerate(r, t, fade) {
  if (fade >= 1) return;
  r.g.globalAlpha = 1 - fade;
  const s = r.scr(t); r.g.strokeStyle = cssv("--danger"); r.g.lineWidth = 2; r.g.setLineDash([4, 4]);
  r.g.beginPath(); r.g.arc(s.x, s.y, 16, 0, TAU); r.g.stroke(); r.g.setLineDash([]);
  r.g.font = "12px " + cssv("--mono"); r.g.fillStyle = cssv("--danger"); r.g.textAlign = "left"; r.g.textBaseline = "middle";
  r.g.globalAlpha = 1;
}

export function drawTab2Pucks(r, u, t) {
  dot3(r, u, 6, cssv("--teal"), cssv("--panel")); label3(r, u, "ref", cssv("--teal"), 9, -2);
  dot3(r, t, 7, cssv("--ink"), cssv("--panel")); label3(r, t, "target", cssv("--ink"), 9, -2);
  dashPath3(r, [{ x: 0, y: 0, z: 0 }, { x: t.x, y: t.y, z: t.z }], cssv("--ink"), 1.2, 0.3);
  label3(r, { x: t.x * 0.52, y: t.y * 0.52, z: t.z * 0.52 }, "r_edge", cssv("--ink"), 8, -2);
}

export function drawTab2Psi(r, P, e1, e2, psiR, fade) {
  if (fade <= 0) return;
  const rad = 0.5;
  const d1 = { x: Math.cos(psiR) * e1.x + Math.sin(psiR) * e2.x, y: Math.cos(psiR) * e1.y + Math.sin(psiR) * e2.y, z: Math.cos(psiR) * e1.z + Math.sin(psiR) * e2.z };
  dashPath3(r, [{ x: 0, y: 0, z: 0 }, { x: e1.x * rad, y: e1.y * rad, z: e1.z * rad }], cssv("--psi"), 1.2, 0.3);
  path3(r, [{ x: 0, y: 0, z: 0 }, { x: d1.x * rad, y: d1.y * rad, z: d1.z * rad }], cssv("--psi"), 1.7, 0.35);
  const Nw = Math.max(2, Math.round(Math.abs(psiR) / (4 * D2R))), wedge = [];
  for (let i = 0; i <= Nw; i++) { const a = psiR * i / Nw; wedge.push({ x: (Math.cos(a) * e1.x + Math.sin(a) * e2.x) * rad * 0.62, y: (Math.cos(a) * e1.y + Math.sin(a) * e2.y) * rad * 0.62, z: (Math.cos(a) * e1.z + Math.sin(a) * e2.z) * rad * 0.62 }); }
  path3(r, wedge, cssv("--psi"), 2.6, 0.4);
  if (wedge.length > 2) arrowHead(r, wedge[wedge.length - 2], wedge[wedge.length - 1], cssv("--psi"));
  label3(r, { x: d1.x * rad, y: d1.y * rad, z: d1.z * rad }, "ψ", cssv("--psi"), 8, -2);
  const cpsi = 26 * D2R;
  const zero = []; for (let i = 0; i <= 20; i++) zero.push(P(cpsi * i / 20, 0));
  dashPath3(r, zero, cssv("--psi"), 1.1, 0.2);
  const N = Math.max(2, Math.round(Math.abs(psiR) / (3 * D2R))), sweep = [];
  for (let i = 0; i <= N; i++) sweep.push(P(cpsi, psiR * i / N));
  path3(r, sweep, cssv("--psi"), 2.4, 0.25);
}

export function drawTab2Instruments(tgt, ang, deg, EPS, psiR, fade) {
  const th = tgt.th;
  const ll = nudge1 / (R_WORLD * Math.sin(th)) / SP;
  const gc = nudge1 / (R_WORLD) / SP;
  setTxt("c2theta", (th * R2D).toFixed(1) + "°");
  setTxt("c2ll", ll.toFixed(2)); setTxt("c2gc", gc.toFixed(2)); setTxt("c2ratio", (ll / gc).toFixed(0));
  const psi = psiR * R2D;
  const sepDeg = deg * R2D, safe = deg >= EPS;
  setTxt("c2r", "10.0 wu"); setTxt("c2c", (ang * R2D).toFixed(1) + "°");
  setTxt("c2psi", (psi >= 0 ? "" : "−") + Math.abs(psi).toFixed(1) + "°");
  setTxt("c2gain", "0.100 /wu");
  const sepEl = document.getElementById("c2sep"); sepEl.textContent = sepDeg.toFixed(1) + "°";
  sepEl.style.color = safe ? "var(--ok)" : "var(--danger)";
  const hp = document.getElementById("c2health");
  hp.className = "pill " + (safe ? "ok" : "bad"); hp.textContent = safe ? "frame safe" : "degenerate";
  const cap = 4;
  document.getElementById("c2llbar").style.width = clamp(ll / cap, 0, 1) * 100 + "%";
  document.getElementById("c2gcbar").style.width = clamp(gc / cap, 0, 1) * 100 + "%";
  const st = document.getElementById("c2status");
  if (fade < 0.5) { st.className = "status bad"; st.innerHTML = '<span class="ico"></span>the frame\'s one singular point — put ref away from your data'; }
  else if (ll / gc > 3) { st.className = "status bad"; st.innerHTML = '<span class="ico"></span>φ grid ' + (ll / gc).toFixed(0) + '× more sensitive — teal unmoved'; }
  else { st.className = "status ok"; st.innerHTML = '<span class="ico"></span>both calm here — drag the target to the pole'; }
}
