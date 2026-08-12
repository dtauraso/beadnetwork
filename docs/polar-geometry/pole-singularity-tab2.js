import {
  R, R_WORLD, SP, D2R, R2D, TAU, cssv, sph, meridian, parallel, path3, dashPath3, dot3,
  arrowHead, label3, clamp, rig, ball, setTxt,
  norm, cross, dot, vsub, vscale, rotAbout,
} from "./pole-singularity-math.js";
import { nudge1 } from "./pole-singularity-tab1.js";

// ============ TAB 2 scene ============
export const r2 = rig(document.getElementById("c2"));
export const ref = { th: 78 * D2R, ph: -1.9 };          // reference direction (recedes from the target)
export const tgt = { th: 9 * D2R, ph: 0.7 };            // draggable target
let showCells = true;                      // draw the ref-frame snapping cells
export function setShowCells(v) { showCells = v; }
function vec(o) { return sph(o.th, o.ph); }
let refE1 = null, refPrevU = null;           // ref-frame azimuth origin, parallel-transported
export function draw2() {
  const r = r2; r.g.clearRect(0, 0, r.w, r.h); ball(r, r.cx, r.cy, r.S);
  // faint +y φ grid for contrast
  for (let a = 0; a < 360; a += 30) path3(r, meridian(a * D2R), cssv("--amber-soft"), 1, 0.07);
  for (let d = 30; d <= 150; d += 30) path3(r, parallel(d * D2R), cssv("--grid"), 1, 0.07);
  dot3(r, sph(0, 0), 3, cssv("--muted"));

  // THE FIX: the reference keeps its distance. Each frame it slides along the geodesic so
  // the target stays between MIN and 180−MIN away — so ref × target never collapses and the
  // singular configuration simply never happens. Drag the target anywhere, even over +y.
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
  const fade = clamp((deg - EPS) / (15 * D2R), 0, 1);   // stays ~1 now — ref never lets it degenerate

  // ref-frame basis (pole = ref), PARALLEL-TRANSPORTED as the ref moves so the cells/wedge
  // translate WITH the ref instead of spinning (the "sphere rotates in reverse" artifact).
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
  const psiR = Math.atan2(dot(t, e2), dot(t, e1));            // bearing ψ (radians), about the ref frame
  const P = (c, a) => ({
    x: Math.cos(c) * u.x + Math.sin(c) * (Math.cos(a) * e1.x + Math.sin(a) * e2.x),
    y: Math.cos(c) * u.y + Math.sin(c) * (Math.cos(a) * e1.y + Math.sin(a) * e2.y),
    z: Math.cos(c) * u.z + Math.sin(c) * (Math.cos(a) * e1.z + Math.sin(a) * e2.z)
  });
  // SNAPPING CELLS in the ref frame: range bands (const c) + bearing meridians (const ψ).
  // They converge on the ref (kept away), so where the target sits they never collapse.
  const SC = 20 * D2R, SPS = 20 * D2R;
  if (showCells) {
    for (let c = SC; c < Math.PI - 1e-3; c += SC) { const a = []; for (let i = 0; i <= 64; i++)a.push(P(c, i / 64 * TAU)); path3(r, a, cssv("--teal-soft"), 1, 0.09); }
    for (let b = 0; b < TAU - 1e-3; b += SPS) { const a = []; for (let i = 0; i <= 48; i++)a.push(P(i / 48 * Math.PI, b)); path3(r, a, cssv("--teal-soft"), 1, 0.09); }
    const bp = ((psiR % TAU) + TAU) % TAU, ci = Math.floor(ang / SC) * SC, bi = Math.floor(bp / SPS) * SPS;   // target's own cell
    [ci, ci + SC].forEach(c => { if (c <= 0 || c >= Math.PI) return; const a = []; for (let i = 0; i <= 64; i++)a.push(P(c, i / 64 * TAU)); path3(r, a, cssv("--teal"), 1.7, 0.18); });
    [bi, bi + SPS].forEach(b => { const a = []; for (let i = 0; i <= 48; i++)a.push(P(i / 48 * Math.PI, b)); path3(r, a, cssv("--teal"), 1.7, 0.18); });
  }

  let vv = { x: t.x - dot(t, u) * u.x, y: t.y - dot(t, u) * u.y, z: t.z - dot(t, u) * u.z };
  const vm = Math.hypot(vv.x, vv.y, vv.z);
  if (fade > 0 && vm > 1e-4) {
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
    // measurement axis = ref × target, oriented to a stable hemisphere so it doesn't pop
    let ax = norm(cross(u, t)); if (ax.y < 0) { ax = { x: -ax.x, y: -ax.y, z: -ax.z }; }
    dashPath3(r, [{ x: -ax.x * R, y: -ax.y * R, z: -ax.z * R }, { x: ax.x * R, y: ax.y * R, z: ax.z * R }], cssv("--muted"), 1.6, 0.3);
    dot3(r, ax, 3.5, cssv("--muted")); dot3(r, { x: -ax.x, y: -ax.y, z: -ax.z }, 3.5, cssv("--muted"));
    label3(r, ax, "axis", cssv("--muted"), 9, -2);
    r.g.globalAlpha = 1;
  }
  // degenerate marker: target sitting on ref (or its antipode)
  if (fade < 1) {
    r.g.globalAlpha = 1 - fade;
    const s = r.scr(t); r.g.strokeStyle = cssv("--danger"); r.g.lineWidth = 2; r.g.setLineDash([4, 4]);
    r.g.beginPath(); r.g.arc(s.x, s.y, 16, 0, TAU); r.g.stroke(); r.g.setLineDash([]);
    r.g.font = "12px " + cssv("--mono"); r.g.fillStyle = cssv("--danger"); r.g.textAlign = "left"; r.g.textBaseline = "middle";
    r.g.globalAlpha = 1;
  }
  // ref + target pucks
  dot3(r, u, 6, cssv("--teal"), cssv("--panel")); label3(r, u, "ref", cssv("--teal"), 9, -2);
  dot3(r, t, 7, cssv("--ink"), cssv("--panel")); label3(r, t, "target", cssv("--ink"), 9, -2);
  // r_edge: the offset magnitude (node → neighbor), drawn as the radius out to the target
  dashPath3(r, [{ x: 0, y: 0, z: 0 }, { x: t.x, y: t.y, z: t.z }], cssv("--ink"), 1.2, 0.3);
  label3(r, { x: t.x * 0.52, y: t.y * 0.52, z: t.z * 0.52 }, "r_edge", cssv("--ink"), 8, -2);
  // ψ is a CENTRAL azimuth about the ref axis — same kind of angle as φ (about +y), just a
  // different axis. Draw it BOTH ways: a wedge at the centre (the definition) and the surface
  // arc / bearing it maps to (the navigation reading) — same number.
  if (fade > 0) {
    const pd = psiR * R2D;
    // (1) central wedge: two radii in the plane ⟂ ref, through the centre
    const rad = 0.5;
    const d1 = { x: Math.cos(psiR) * e1.x + Math.sin(psiR) * e2.x, y: Math.cos(psiR) * e1.y + Math.sin(psiR) * e2.y, z: Math.cos(psiR) * e1.z + Math.sin(psiR) * e2.z };
    dashPath3(r, [{ x: 0, y: 0, z: 0 }, { x: e1.x * rad, y: e1.y * rad, z: e1.z * rad }], cssv("--psi"), 1.2, 0.3);   // ψ=0 radius
    path3(r, [{ x: 0, y: 0, z: 0 }, { x: d1.x * rad, y: d1.y * rad, z: d1.z * rad }], cssv("--psi"), 1.7, 0.35);        // ψ radius
    const Nw = Math.max(2, Math.round(Math.abs(psiR) / (4 * D2R))), wedge = [];
    for (let i = 0; i <= Nw; i++) { const a = psiR * i / Nw; wedge.push({ x: (Math.cos(a) * e1.x + Math.sin(a) * e2.x) * rad * 0.62, y: (Math.cos(a) * e1.y + Math.sin(a) * e2.y) * rad * 0.62, z: (Math.cos(a) * e1.z + Math.sin(a) * e2.z) * rad * 0.62 }); }
    path3(r, wedge, cssv("--psi"), 2.6, 0.4);
    if (wedge.length > 2) arrowHead(r, wedge[wedge.length - 2], wedge[wedge.length - 1], cssv("--psi"));
    label3(r, { x: d1.x * rad, y: d1.y * rad, z: d1.z * rad }, "ψ", cssv("--psi"), 8, -2);
    // (2) same angle, read on the surface as the departure bearing
    const cpsi = 26 * D2R;
    const zero = []; for (let i = 0; i <= 20; i++) zero.push(P(cpsi * i / 20, 0));
    dashPath3(r, zero, cssv("--psi"), 1.1, 0.2);
    const N = Math.max(2, Math.round(Math.abs(psiR) / (3 * D2R))), sweep = [];
    for (let i = 0; i <= N; i++) sweep.push(P(cpsi, psiR * i / N));
    path3(r, sweep, cssv("--psi"), 2.4, 0.25);
  }

  // instruments — sensitivity vs the target's colatitude from +y
  const th = tgt.th;
  const ll = nudge1 / (R_WORLD * Math.sin(th)) / SP;   // φ about fixed +y
  const gc = nudge1 / (R_WORLD) / SP;                // great-circle range: uniform
  setTxt("c2theta", (th * R2D).toFixed(1) + "°");
  setTxt("c2ll", ll.toFixed(2)); setTxt("c2gc", gc.toFixed(2)); setTxt("c2ratio", (ll / gc).toFixed(0));

  // ---- target's stored great-circle properties (basis e1/e2 + psiR computed above) ----
  // r_edge is the NEIGHBOR distance (offset = neighbor − thisNode), not the scene-sphere radius.
  const psi = psiR * R2D;
  const sepDeg = deg * R2D, safe = deg >= EPS;
  setTxt("c2r", "10.0 wu"); setTxt("c2c", (ang * R2D).toFixed(1) + "°");
  setTxt("c2psi", (psi >= 0 ? "" : "−") + Math.abs(psi).toFixed(1) + "°");
  setTxt("c2gain", "0.100 /wu");
  const sepEl = document.getElementById("c2sep"); sepEl.textContent = sepDeg.toFixed(1) + "°";
  sepEl.style.color = safe ? "var(--ok)" : "var(--danger)";
  const hp = document.getElementById("c2health");
  hp.className = "pill " + (safe ? "ok" : "bad"); hp.textContent = safe ? "frame safe" : "degenerate";
  const cap = 4; // meter full-scale = 4 cells
  document.getElementById("c2llbar").style.width = clamp(ll / cap, 0, 1) * 100 + "%";
  document.getElementById("c2gcbar").style.width = clamp(gc / cap, 0, 1) * 100 + "%";
  const st = document.getElementById("c2status");
  if (fade < 0.5) { st.className = "status bad"; st.innerHTML = '<span class="ico"></span>the frame\'s one singular point — put ref away from your data'; }
  else if (ll / gc > 3) { st.className = "status bad"; st.innerHTML = '<span class="ico"></span>φ grid ' + (ll / gc).toFixed(0) + '× more sensitive — teal unmoved'; }
  else { st.className = "status ok"; st.innerHTML = '<span class="ico"></span>both calm here — drag the target to the pole'; }
}
