const pressName = run => `sequence ${run.sequence}: ${pressesName(run.presses)}`;

function pressView(before, after, presses) {
  const c = compare(before, after, presses);
  const linkRows = c.links.map(l => el("tr", { class: l.same ? "" : "badrow" },
    num(l.name), num(ticks(l.was)), num(ticks(l.expected)), num(ticks(l.recorded)), verdict(l.same, "yes", "no")));
  const nodeRows = c.nodes.map(n => el("tr", { class: n.same ? "" : "badrow" },
    num(n.id), num(xyz(n.expected)), num(xyz(n.recorded)), num(n.gap === null ? "—" : n.gap.toExponential(1)), verdict(n.same, "yes", "no")));
  const part = el("div", { class: "seq" },
    el("h3", {}, `Press ${pressesName(presses)}: predicted from Before, recorded in After (rule 6)`),
    table(["link", "Δ before", "Δ predicted", "Δ recorded", "same"], linkRows),
    table(["node", "predicted x  y  z", "recorded x  y  z", "gap", "same"], nodeRows),
    el("p", { class: c.closes ? "muted" : "error" }, c.closes ? "Every link in After closes (rule 5)." : "A link in After does not close (rule 5)."));
  return { part, ok: c.ok };
}

function runsChip(runs) {
  const failed = runs.filter(run => !compare(run.before, run.after, run.presses).ok).map(pressName);
  return chip(!failed.length, `runs: ${runs.length - failed.length} of ${runs.length} match the prediction (rule 6)`
    + (failed.length ? " · failing: " + failed.join(", ") : ""));
}
