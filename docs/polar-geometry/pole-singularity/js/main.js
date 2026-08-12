import { cam, sph, clamp, pickHemi, isDark, interaction, setTxt } from "./math.js";
import { r1, n1, draw1, setNudge1, setJit1, isJit1 } from "./tab1.js";
import { r2, tgt, draw2, setShowCells } from "./tab2.js";

// ---- interaction: orbit vs drag-node ----
function bind(cv, r, node, redraw) {
  let mode = null, lastX = 0, lastY = 0;
  const local = e => { const b = cv.getBoundingClientRect(); return [(e.clientX - b.left) * (r.w / b.width), (e.clientY - b.top) * (r.h / b.height)]; };
  cv.addEventListener("pointerdown", e => {
    cv.setPointerCapture(e.pointerId); const [x, y] = local(e);
    const s = r.scr(sph(node.th, node.ph)); interaction.dragging = true;
    // grab the node whether it's on the near OR far side (its screen dot is drawn either way)
    mode = (Math.hypot(x - s.x, y - s.y) < 22) ? "node" : "orbit"; lastX = x; lastY = y;
    if (mode === "node") { const p = pickHemi(r, x, y); node.th = p.th; node.ph = p.ph; } redraw();
  });
  cv.addEventListener("pointermove", e => {
    if (!interaction.dragging) return; const [x, y] = local(e);
    if (mode === "node") { const p = pickHemi(r, x, y); node.th = p.th; node.ph = p.ph; }
    else { cam.az += (x - lastX) * 0.008; cam.el = clamp(cam.el - (y - lastY) * 0.008, -1.35, 1.35); }
    lastX = x; lastY = y; redraw();
  });
  const up = () => { interaction.dragging = false; mode = null; if (isJit1()) draw1(); };
  cv.addEventListener("pointerup", up); cv.addEventListener("pointercancel", up);
}
bind(document.getElementById("c1"), r1, n1, draw1);
bind(document.getElementById("c2"), r2, tgt, draw2);

// ---- controls ----
document.getElementById("c1nudge").addEventListener("input", e => { setNudge1(+e.target.value); setTxt("c1nv", (+e.target.value).toFixed(2)); draw1(); draw2(); });
document.getElementById("c1jit").addEventListener("change", e => { setJit1(e.target.checked); draw1(); });
document.getElementById("c2cells").addEventListener("change", e => { setShowCells(e.target.checked); draw2(); });
document.getElementById("themebtn").addEventListener("click", () => { const d = isDark(); document.documentElement.setAttribute("data-theme", d ? "light" : "dark"); draw1(); draw2(); });

// ---- tabs ----
const tabs = [document.getElementById("tab1"), document.getElementById("tab2")];
const panels = [document.getElementById("panel1"), document.getElementById("panel2")];
tabs.forEach((t, i) => t.addEventListener("click", () => {
  tabs.forEach((x, j) => x.setAttribute("aria-selected", j === i));
  panels.forEach((p, j) => { if (j === i) p.setAttribute("data-active", ""); else p.removeAttribute("data-active"); });
  if (i === 0) { r1.size(); draw1(); } else { r2.size(); draw2(); }
}));

function boot() { r1.size(); r2.size(); draw1(); draw2(); }
addEventListener("resize", () => { r1.size(); r2.size(); draw1(); draw2(); });
boot();
