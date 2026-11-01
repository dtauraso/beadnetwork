function nodeRows(seq, st) {
  return seq.nodes.map(n => el("tr", {},
    num(n.id), num(n.s),
    num(n.place ? fmt(n.place.phi) : "—"),
    num(n.place ? fmt(n.place.theta) : "—"),
    num(n.place ? fmt(n.place.r) : "—"),
    num(xyz(st.positions[n.id]))));
}

function linkRows(st) {
  return st.links.map(r => {
    const l = r.link, has = r.gap !== undefined;
    return el("tr", { class: has && !r.closes ? "badrow" : "" },
      num(l.from + "→" + l.to),
      el("td", {}, l.picked ? "yes" : "no"),
      num([l.start.phi, l.start.theta, l.start.r].map(fmt).join("  ")),
      num(ticks(l.delta)),
      num([r.points.phi, r.points.theta, r.points.r].map(fmt).join("  ") + (r.flipped ? "  (z flipped)" : "")),
      num(xyz(r.vector)),
      num(xyz(r.end)),
      num(xyz(st.positions[l.to])),
      num(has ? r.gap.toExponential(1) : "—"),
      has ? verdict(r.closes, "yes", "no") : el("td", {}, "a node has no place"));
  });
}

function sequenceView(seq) {
  const st = checkSequence(seq);
  const f = seq.frame;
  const card = el("article", { class: "panel" },
    el("div", { class: "seqhead" }, el("h2", {}, "Sequence " + seq.sequence), el("span", { class: "muted mono" }, seq.name || "")),
    el("p", { class: "muted mono" }, `ρ=${f.rho}  constant-r=${f.constantR}  max-index-φ=${f.maxIndexPhi}  max-index-θ=${f.maxIndexTheta}`),
    el("h3", {}, "Nodes (rule 4)"),
    table(["node", "s", "place φ", "place θ", "place r", "position x  y  z"], nodeRows(seq, st)),
    el("h3", {}, "Links (rules 1, 3 and 5)"),
    table(["link", "selected", "start φ θ r", "Δ φ θ r", "points φ θ r", "vector x y z", "a + vector", "b is at", "gap", "closes"], linkRows(st)));
  const rows = st.links.filter(r => r.gap !== undefined);
  return { card, closed: rows.filter(r => r.closes).length, open: rows.filter(r => !r.closes).length };
}
