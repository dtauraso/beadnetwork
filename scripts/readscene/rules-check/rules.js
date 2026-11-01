const TAU = 2 * Math.PI;
const CLOSE = 1e-6;

function wrap(t, T) {
  return ((t % T) + T) % T;
}

function linkVector(frame, s, link) {
  const T = 12 * s;
  const phi = wrap(link.start.phi + link.delta.phi, T);
  const theta = wrap(link.start.theta + link.delta.theta, T);
  const r = link.start.r + link.delta.r;
  const a = t => t * TAU / T;
  const v = [
    r * frame.rho * Math.sin(a(phi)) * Math.cos(a(theta)),
    r * frame.rho * Math.cos(a(phi)),
    r * frame.rho * Math.sin(a(phi)) * Math.sin(a(theta)),
  ];
  const flipped = phi > T / 2 && Math.abs(phi - theta) < 1e-9;
  if (flipped) v[2] = -v[2];
  return { phi, theta, r, v, flipped };
}

function nodePosition(frame, place) {
  const R = place.r * frame.constantR;
  const phi = place.phi * TAU / frame.maxIndexPhi;
  const theta = place.theta * TAU / frame.maxIndexTheta;
  return [R * Math.sin(phi) * Math.cos(theta), R * Math.cos(phi), R * Math.sin(phi) * Math.sin(theta)];
}

function gap(p, q) {
  return Math.hypot(p[0] - q[0], p[1] - q[1], p[2] - q[2]);
}

function checkState(seq) {
  const pos = {}, scale = {};
  for (const n of seq.nodes) {
    scale[n.id] = n.s;
    if (n.place) pos[n.id] = nodePosition(seq.frame, n.place);
  }
  const links = seq.links.map(l => {
    const lv = linkVector(seq.frame, scale[l.to], l);
    const row = { link: l, ...lv };
    if (pos[l.from] && pos[l.to]) {
      row.end = pos[l.from].map((c, i) => c + lv.v[i]);
      row.gap = gap(row.end, pos[l.to]);
      row.closes = row.gap <= CLOSE * Math.max(1, Math.hypot(...pos[l.to]));
    }
    return row;
  });
  return { pos, scale, links };
}

function checkStep(before, after) {
  const key = l => l.from + "→" + l.to;
  const was = Object.fromEntries(before.links.map(l => [key(l), l]));
  const placeOf = seq => Object.fromEntries(seq.nodes.filter(n => n.place).map(n => [n.id, nodePosition(seq.frame, n.place)]));
  const p0 = placeOf(before), p1 = placeOf(after);
  const moved = id => p0[id] && p1[id] ? gap(p0[id], p1[id]) > CLOSE * Math.max(1, Math.hypot(...p1[id])) : false;
  const scale = Object.fromEntries(after.nodes.map(n => [n.id, n.s]));
  const same = (a, b) => Math.abs(a - b) < 1e-9;
  return after.links.filter(l => was[key(l)]).map(l => {
    const b = was[key(l)], T = 12 * scale[l.to];
    const changed = [];
    for (const k of ["phi", "theta", "r"]) {
      if (b.delta[k] === l.delta[k]) continue;
      const up = k === "r" ? b.delta[k] + b.increment[k] : wrap(b.delta[k] + b.increment[k], T);
      const down = k === "r" ? b.delta[k] - b.increment[k] : wrap(b.delta[k] - b.increment[k], T);
      changed.push({ k, from: b.delta[k], to: l.delta[k], byIncrement: l.delta[k] === up || l.delta[k] === down });
    }
    const startSame = same(b.start.phi, l.start.phi) && same(b.start.theta, l.start.theta) && same(b.start.r, l.start.r);
    const touchesMoved = moved(l.from) || moved(l.to);
    let ok, why;
    if (changed.length) {
      ok = b.picked && startSame && changed.every(c => c.byIncrement);
      why = !b.picked ? "Δ changed on a link that was not selected"
        : !startSame ? "Δ changed and the start changed too"
        : !changed.every(c => c.byIncrement) ? "Δ changed by something other than its increment"
        : "stepped: Δ by its increment, start kept";
    } else if (!startSame) {
      ok = touchesMoved;
      why = touchesMoved ? "not stepped: start recalculated, a node it touches moved" : "start changed, but neither of its nodes moved";
    } else {
      ok = true;
      why = "unchanged";
    }
    return { link: l, changed, startSame, touchesMoved, ok, why };
  });
}

const RULES_SOURCE = [
  "const TAU = 2 * Math.PI;", "const CLOSE = 1e-6;",
  wrap, linkVector, nodePosition, gap, checkState, checkStep,
].map(String).join("\n\n");
