const $ = id => document.getElementById(id);
const fmt = x => (Math.abs(x) < 5e-7 ? 0 : x).toFixed(6);
const xyz = v => v ? [v.x, v.y, v.z].map(fmt).join("  ") : "—";
const axis = k => ({ phi: "φ", theta: "θ", r: "r", both: "both" })[k];
const presses = Array.isArray(window.SCENE_STEPS) ? window.SCENE_STEPS : [];
const pressName = p => `sequence ${p.sequence}: ${axis(p.button.k)} ${p.button.sign > 0 ? "+" : "−"}`;
let original = "";

function recorded() {
  const failed = [];
  for (const p of presses) {
    const stepOk = checkStep(p.before, p.after).every(r => r.ok);
    const closes = checkSequence(p.after).links.every(r => r.gap === undefined || r.closes);
    if (!stepOk || !closes) failed.push(pressName(p));
  }
  return failed;
}

function parse(text, label) {
  const data = JSON.parse(text);
  if (!Array.isArray(data)) throw new Error(label + " should be a list of sequences, the shape readscene -json prints.");
  for (const s of data) {
    if (!s.frame || !Array.isArray(s.nodes) || !Array.isArray(s.links)) {
      throw new Error(label + ": sequence " + (s.sequence ?? "?") + " is missing frame, nodes or links.");
    }
  }
  return data;
}

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

const tally = { closed: 0, open: 0, stepOk: 0, stepBad: 0 };

function count(links) {
  for (const r of links) if (r.gap !== undefined) (r.closes ? tally.closed++ : tally.open++);
}

function nodeRows(seq, st) {
  return seq.nodes.map(n => el("tr", {},
    el("td", { class: "num" }, n.id), el("td", { class: "num" }, n.s),
    el("td", { class: "num" }, n.place ? fmt(n.place.phi) : "—"),
    el("td", { class: "num" }, n.place ? fmt(n.place.theta) : "—"),
    el("td", { class: "num" }, n.place ? fmt(n.place.r) : "—"),
    el("td", { class: "num" }, xyz(st.positions[n.id]))));
}

function linkRows(st) {
  return st.links.map(r => {
    const l = r.link, has = r.gap !== undefined;
    return el("tr", { class: has && !r.closes ? "badrow" : "" },
      el("td", { class: "num" }, l.from + "→" + l.to),
      el("td", {}, l.picked ? "yes" : "no"),
      el("td", { class: "num" }, [l.start.phi, l.start.theta, l.start.r].map(fmt).join("  ")),
      el("td", { class: "num" }, `${l.delta.phi}  ${l.delta.theta}  ${l.delta.r}`),
      el("td", { class: "num" }, [r.points.phi, r.points.theta, r.points.r].map(fmt).join("  ") + (r.flipped ? "  (z flipped)" : "")),
      el("td", { class: "num" }, xyz(r.vector)),
      el("td", { class: "num" }, xyz(r.end)),
      el("td", { class: "num" }, xyz(st.positions[l.to])),
      el("td", { class: "num" }, has ? r.gap.toExponential(1) : "—"),
      el("td", { class: has ? (r.closes ? "ok" : "bad") : "" }, has ? (r.closes ? "yes" : "no") : "a node has no place"));
  });
}

function stepRows(seq, next, ns) {
  return checkStep(seq, next).map(r => {
    r.ok ? tally.stepOk++ : tally.stepBad++;
    const after = ns.links.find(x => x.link.from === r.link.from && x.link.to === r.link.to);
    const closes = after && after.gap !== undefined ? (after.closes ? "yes" : "no") : "—";
    const changed = AXES.filter(k => r.old.delta[k] !== r.link.delta[k]);
    return el("tr", { class: r.ok ? "" : "badrow" },
      el("td", { class: "num" }, r.link.from + "→" + r.link.to),
      el("td", { class: "num" }, changed.length ? changed.map(k => `Δ${axis(k)} ${r.old.delta[k]}→${r.link.delta[k]}`).join(", ") : "none"),
      el("td", {}, r.startKept ? "yes" : "no"),
      el("td", {}, r.aNodeMoved ? "yes" : "no"),
      el("td", { class: closes === "yes" ? "ok" : "bad" }, closes),
      el("td", { class: r.ok ? "ok" : "bad" }, r.why));
  });
}

function sequenceCard(seq, after) {
  const st = checkSequence(seq);
  count(st.links);
  const f = seq.frame;
  const card = el("article", { class: "panel" },
    el("div", { class: "seqhead" }, el("h2", {}, "Sequence " + seq.sequence), el("span", { class: "muted mono" }, seq.name || "")),
    el("p", { class: "muted mono" }, `ρ=${f.rho}  constant-r=${f.constantR}  max-index-φ=${f.maxIndexPhi}  max-index-θ=${f.maxIndexTheta}`),
    el("h3", {}, "Nodes (rule 4)"),
    table(["node", "s", "place φ", "place θ", "place r", "position x  y  z"], nodeRows(seq, st)),
    el("h3", {}, "Links (rules 1, 3 and 5)"),
    table(["link", "selected", "start φ θ r", "Δ φ θ r", "points φ θ r", "vector x y z", "a + vector", "b is at", "gap", "closes"], linkRows(st)));
  const next = after && after.find(s => s.sequence === seq.sequence);
  if (next) {
    const ns = checkSequence(next);
    count(ns.links);
    card.append(el("h3", {}, "Before → after (rule 6, and rule 5 on the after state)"),
      table(["link", "Δ changes", "start kept", "touches a moved node", "after closes", "verdict"], stepRows(seq, next, ns)));
  }
  return card;
}

function render() {
  const out = $("results");
  out.replaceChildren();
  let before, after = null;
  try {
    before = parse($("before").value, "Before");
    if ($("after").value.trim()) after = parse($("after").value, "After");
  } catch (e) {
    out.append(el("div", { class: "panel" }, el("p", { class: "error" }, "Could not read the data. " + e.message)));
    $("status").textContent = "";
    return;
  }
  Object.assign(tally, { closed: 0, open: 0, stepOk: 0, stepBad: 0 });
  const cards = before.map(seq => sequenceCard(seq, after));
  const chips = el("div", { class: "summary" },
    el("span", { class: "chip " + (tally.open ? "fail" : "pass") }, `${tally.closed} links close · ${tally.open} do not`));
  if (after) chips.append(el("span", { class: "chip " + (tally.stepBad ? "fail" : "pass") }, `rule 6: ${tally.stepOk} links follow it · ${tally.stepBad} do not`));
  if (presses.length) {
    const failed = recorded();
    chips.append(el("span", { class: "chip " + (failed.length ? "fail" : "pass") },
      `recorded presses: ${presses.length - failed.length} of ${presses.length} follow rules 5 and 6` + (failed.length ? " · failing: " + failed.join(", ") : "")));
  }
  out.append(el("div", { class: "panel" }, el("h2", {}, "Results"), chips), ...cards);
  $("status").textContent = "Checked " + before.length + " sequence" + (before.length === 1 ? "" : "s") + ".";
}

function loadFile(input, target) {
  input.addEventListener("change", () => {
    const f = input.files && input.files[0];
    if (!f) return;
    const r = new FileReader();
    r.onload = () => { $(target).value = r.result; render(); };
    r.readAsText(f);
  });
}

$("check").addEventListener("click", render);
$("reset").addEventListener("click", () => { $("before").value = original; $("after").value = ""; render(); });
loadFile($("beforeFile"), "before");
loadFile($("afterFile"), "after");
$("press").append(el("option", { value: "" }, presses.length ? "choose one" : "scene-steps.js did not load"),
  ...presses.map((p, i) => el("option", { value: String(i) }, pressName(p))));
$("press").addEventListener("change", () => {
  const p = presses[$("press").value];
  if (!p) return;
  $("before").value = JSON.stringify([p.before], null, 2);
  $("after").value = JSON.stringify([p.after], null, 2);
  render();
  $("status").textContent = "Loaded " + pressName(p) + "." + (p.refused.length ? " Refused: " + p.refused.join("; ") : "");
});
$("source").textContent = RULES_SOURCE;
for (const pre of document.querySelectorAll("pre[data-fn]")) {
  pre.textContent = pre.dataset.fn.split(" ").map(n => String(window[n])).join("\n\n");
}

if (Array.isArray(window.SCENE_EXPORT)) {
  original = JSON.stringify(window.SCENE_EXPORT, null, 2);
  $("before").value = original;
  render();
} else {
  $("results").replaceChildren(el("div", { class: "panel" }, el("p", { class: "error" }, "scene-export.js did not load, so there is no data yet. Paste an export into Before and press Check.")));
}
