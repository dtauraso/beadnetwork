const $ = id => document.getElementById(id);
const fmt = x => (Math.abs(x) < 5e-7 ? 0 : x).toFixed(6);
const xyz = v => v ? [v.x, v.y, v.z].map(fmt).join("  ") : "—";
const ticks = d => d ? `${d.phi}  ${d.theta}  ${d.r}` : "—";
const axisName = k => ({ phi: "φ", theta: "θ", r: "r", both: "both" })[k];
const pressesName = presses => presses.map(b => axisName(b.k) + (b.sign > 0 ? "+" : "−")).join(" ");

function el(tag, attrs, ...kids) {
  const e = document.createElement(tag);
  for (const [k, v] of Object.entries(attrs || {})) {
    if (k === "class") e.className = v; else e.setAttribute(k, v);
  }
  for (const k of kids) e.append(k instanceof Node ? k : document.createTextNode(String(k)));
  return e;
}

function table(head, rows) {
  const t = el("table", {}, el("thead", {}, el("tr", {}, ...head.map(h => el("th", {}, h)))));
  const body = el("tbody");
  for (const r of rows) body.append(r);
  t.append(body);
  return el("div", { class: "scroll" }, t);
}

const num = x => el("td", { class: "num" }, x);
const verdict = (ok, yes, no) => el("td", { class: ok ? "ok" : "bad" }, ok ? yes : no);
const chip = (ok, text) => el("span", { class: "chip " + (ok ? "pass" : "fail") }, text);
