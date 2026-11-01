function orderView(runs) {
  const results = orderCheck(runs);
  const cards = results.map(r => {
    const rows = r.links.map(l => el("tr", { class: l.sameDelta && l.deltaKept ? "" : "badrow" },
      num(l.name),
      el("td", { class: "num" }, ...r.orders.map((o, i) => el("div", {}, pressesName(o) + "  →  " + ticks(l.deltas[i])))),
      verdict(l.sameDelta, "yes", "no"),
      num(l.rewritten + " of " + r.orders.length),
      verdict(l.deltaKept, "yes", "no")));
    return el("div", { class: "seq" },
      el("h3", {}, "Sequence " + r.sequence),
      table(["link", "Δ φ θ r after each order", "same Δ in every order", "start rewritten by the move", "Δ kept when start was rewritten"], rows));
  });
  const bad = results.filter(r => !r.ok).map(r => "sequence " + r.sequence);
  const head = [el("h2", {}, "Order of presses (rule 7)"),
    el("p", { class: "muted" }, "Runs in the Runs box that press the same buttons on the same sequence in different orders are compared with each other.")];
  if (!results.length) return el("section", { class: "panel" }, ...head, el("p", { class: "muted" }, "No runs in the Runs box press the same buttons in different orders."));
  return el("section", { class: "panel" }, ...head,
    el("div", { class: "summary" }, chip(!bad.length, bad.length ? "Δ depends on order in " + bad.join(", ") : `Δ is the same in every order, in ${results.length} groups of runs`)),
    ...cards);
}
