const bDir = (builder, row, s = bS(builder, row.from)) => ({ ...row.start, phi: bWrap(s, row.start.phi + row.delta.phi), theta: bWrap(s, row.start.theta + row.delta.theta) });

function bDirs(builder, tree) {
  return [{ t: bDir(builder, tree.ab.row) }];
}

const bBack = (builder, row, a = bDir(builder, row)) => {
  const sa = bS(builder, row.from), sb = bS(builder, row.to);
  return { phi: bWrap(sb, a.phi * sb / sa + 6 * sb), theta: bWrap(sb, a.theta * sb / sa), r: a.r };
};

function bMeasure(builder, row, P = bNodePos(builder, row.from), T = bNodePos(builder, row.to)) {
  if (!P || !T) return null;
  const s = bS(builder, row.from), v = bSub(T, P), len = bLen(v), tick = 2 * Math.PI / (12 * s), half = 6 * s;
  const was = bDir(builder, row) || { phi: 0, theta: 0 };
  if (len < 1e-9) return { phi: was.phi, theta: was.theta, r: 0 };
  const side = Math.hypot(v[0], v[2]), phi = Math.atan2(side, v[1]) / tick;
  const theta = side < 1e-9 * len ? was.theta : Math.atan2(v[2], v[0]) / tick;
  const gap = (a, b) => { const d = bWrap(s, a - b); return Math.min(d, 12 * s - d); };
  const off = (p, t) => gap(p, was.phi) + gap(t, was.theta);
  const [p, t] = off(phi, theta) <= off(-phi, theta + half) ? [phi, theta] : [-phi, theta + half];
  return { phi: bWrap(s, p), theta: bWrap(s, t), r: len / builder.rho };
}

function bRowEnds(builder) {
  return [...bOpenStep(builder).edges].flatMap(tree => bDirs(builder, tree).map(({ t }) => {
    const { ab } = tree, row = ab.row, P = bNodePos(builder, ab.a);
    if (!P) return null;
    const d = bVec(builder, bS(builder, ab.a), t);
    return { ab, row, P, d, t, T: bNodePos(builder, ab.b) || bAdd(P, d) };
  })).filter(Boolean);
}

function bLinks(builder) {
  return bRowEnds(builder).flatMap(({ row, P, T, t }) => [
    { from: row.from, to: row.to, P, T, t, s: bS(builder, row.from) },
    { from: row.to, to: row.from, P: T, T: P, t: bBack(builder, row, t), s: bS(builder, row.to) },
  ]);
}
