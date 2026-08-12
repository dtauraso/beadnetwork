import { R_WORLD, SP_DEG, SP, D2R, R2D, cssv, sph, meridian, path3, parallel, dot3, label3, clamp, rig, setTxt, ball, interaction } from "./math.js";

// ============ TAB 1 scene ============
export const r1 = rig(document.getElementById("c1"));
export const n1 = { th: 5 * D2R, ph: 0.6 };
export let nudge1 = 0.05;
let jit1 = false, jitP = 0;
export function setNudge1(v) { nudge1 = v; }
export function setJit1(v) { jit1 = v; }
export function isJit1() { return jit1; }
export function draw1() {
  const r = r1; r.g.clearRect(0, 0, r.w, r.h); ball(r, r.cx, r.cy, r.S);
  // parallels (faint)
  for (let d = 30; d <= 150; d += 30) path3(r, parallel(d * D2R), cssv("--grid"), 1, 0.10);
  // φ cells = meridians converging at pole (amber)
  for (let a = 0; a < 360; a += SP_DEG) path3(r, meridian(a * D2R), cssv("--amber-soft"), 1.1, 0.10);
  // pole marker
  dot3(r, sph(0, 0), 4, cssv("--muted")); label3(r, sph(0, 0), "+y", cssv("--muted"), 8, -2);

  // node phi with optional world-jitter
  let ph = n1.ph;
  if (jit1) { jitP += 0.14; const wob = 0.018 * Math.sin(jitP * 1.7) + 0.013 * Math.sin(jitP * 0.9); ph += wob / (R_WORLD * Math.sin(n1.th)); }

  // the node's own φ-cell: highlight its two bounding meridians (they pinch at the pole)
  const k = Math.floor(ph / SP);
  path3(r, meridian(k * SP), cssv("--amber"), 2.4, 0.18);
  path3(r, meridian((k + 1) * SP), cssv("--amber"), 2.4, 0.18);

  // fixed world nudge -> Δφ, swept as a thick arc along the node's parallel
  const dphi = nudge1 / (R_WORLD * Math.sin(n1.th));
  const cellWorld = R_WORLD * Math.sin(n1.th) * SP;
  const sweep = []; for (let i = 0; i <= 40; i++) sweep.push(sph(n1.th, ph - dphi + (2 * dphi) * i / 40));
  path3(r, sweep, cssv("--amber"), 6, 0.5);
  const dqp = Math.round((ph + dphi) / SP) - Math.round(ph / SP);

  // node puck (glossy)
  dot3(r, sph(n1.th, ph), 7, cssv("--ink"), cssv("--panel"));

  // instruments
  setTxt("c1theta", (n1.th * R2D).toFixed(1) + "°"); setTxt("c1dphi", (dphi * R2D).toFixed(1) + "°");
  setTxt("c1dqp", (dqp >= 0 ? "" : "−") + Math.abs(dqp)); setTxt("c1sectorlab", Math.abs(dqp) === 1 ? "cell" : "cells");
  document.getElementById("c1dqpbar").style.width = clamp(Math.abs(dqp) / 12, 0, 1) * 100 + "%";
  // cell-width meter: full at equator cell (r*1*sp), shrinks toward pole
  const cellMax = R_WORLD * SP;
  document.getElementById("c1cellbar").style.width = clamp(cellWorld / cellMax, 0, 1) * 100 + "%";
  const st = document.getElementById("c1status");
  if (Math.abs(dqp) === 0) { st.className = "status ok"; st.innerHTML = '<span class="ico"></span>absorbed — the grid hides the nudge'; }
  else { st.className = "status bad"; st.innerHTML = '<span class="ico"></span>amplified — a ' + nudge1.toFixed(2) + ' nudge flips the stored cell by ' + Math.abs(dqp); }
  if (jit1 && !interaction.dragging) requestAnimationFrame(draw1);
}
