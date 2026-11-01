const shipped = {
  before: Array.isArray(window.SCENE_EXPORT) ? JSON.stringify(window.SCENE_EXPORT, null, 2) : "",
  runs: JSON.stringify([...(window.SCENE_STEPS || []), ...(window.SCENE_ORDERS || [])], null, 1),
};
let runs = [];
let chosen = null;

function parseSequences(text, label) {
  const data = JSON.parse(text);
  if (!Array.isArray(data)) throw new Error(label + " should be a list of sequences, the shape readscene -json prints.");
  for (const s of data) {
    if (!s.frame || !Array.isArray(s.nodes) || !Array.isArray(s.links)) {
      throw new Error(label + ": sequence " + (s.sequence ?? "?") + " is missing frame, nodes or links.");
    }
  }
  return data;
}

function parseRuns(text) {
  if (!text.trim()) return [];
  const data = JSON.parse(text);
  if (!Array.isArray(data)) throw new Error("Runs should be a list, the shape readscene -steps and -orders print.");
  for (const run of data) {
    if (!Array.isArray(run.presses) || !run.before || !run.after) throw new Error("Runs: each run needs presses, before and after.");
    parseSequences(JSON.stringify([run.before, run.after]), "Runs, " + pressName(run));
  }
  return data;
}

function fillPresses() {
  $("press").replaceChildren(el("option", { value: "" }, runs.length ? "choose one" : "the Runs box is empty"),
    ...runs.map((run, i) => el("option", { value: String(i) }, pressName(run))));
}

function render() {
  const out = $("results");
  out.replaceChildren();
  let before, after = null;
  try {
    before = parseSequences($("before").value, "Before");
    if ($("after").value.trim()) after = parseSequences($("after").value, "After");
    runs = parseRuns($("runs").value);
  } catch (e) {
    out.append(el("div", { class: "panel" }, el("p", { class: "error" }, "Could not read the data. " + e.message)));
    return;
  }
  fillPresses();
  let closed = 0, open = 0, pressOk = 0, pressBad = 0;
  const cards = before.map(seq => {
    const v = sequenceView(seq);
    closed += v.closed;
    open += v.open;
    const next = after && after.find(s => s.sequence === seq.sequence);
    if (next && chosen) {
      const p = pressView(seq, next, chosen.presses);
      p.ok ? pressOk++ : pressBad++;
      v.card.append(p.part);
    } else if (next) {
      v.card.append(el("p", { class: "muted" }, "Load a run below, so the page knows which buttons to predict for After."));
    }
    return v.card;
  });
  const chips = el("div", { class: "summary" }, chip(!open, `${closed} links close · ${open} do not`));
  if (pressOk + pressBad) chips.append(chip(!pressBad, `rule 6: the prediction ${pressBad ? "does not match" : "matches"} After`));
  if (runs.length) chips.append(runsChip(runs));
  out.append(el("div", { class: "panel" }, el("h2", {}, "Results"), chips), ...cards, orderView(runs));
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

function choosePress() {
  chosen = runs[$("press").value] || null;
  if (!chosen) return;
  const picked = $("press").value;
  $("before").value = JSON.stringify([chosen.before], null, 2);
  $("after").value = JSON.stringify([chosen.after], null, 2);
  render();
  $("press").value = picked;
  $("status").textContent = "Loaded " + pressName(chosen) + "." + (chosen.refused.length ? " Refused: " + chosen.refused.join("; ") : "");
}

function reset() {
  chosen = null;
  $("before").value = shipped.before;
  $("after").value = "";
  $("runs").value = shipped.runs;
  $("status").textContent = "";
  render();
}

$("check").addEventListener("click", render);
$("reset").addEventListener("click", reset);
loadFile($("beforeFile"), "before");
loadFile($("afterFile"), "after");
loadFile($("runsFile"), "runs");
$("press").addEventListener("change", choosePress);
codeView();
goView();
if (shipped.before) {
  reset();
} else {
  $("results").replaceChildren(el("div", { class: "panel" }, el("p", { class: "error" }, "scene-export.js did not load, so there is no data yet. Paste an export into Before and press Check.")));
}
