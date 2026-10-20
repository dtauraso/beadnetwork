function bCellPoints(builder, rs) {
  const out = [];
  const had = new Set();
  const add = (p, color) => {
    const key = p.map(v => v.toFixed(6)).join();
    if (bLen(p) > 1e-9 && !had.has(key)) { had.add(key); out.push({ p, color }); }
  };
  bShown(builder).forEach(n => add(bNodePos(builder, n), '--ink'));
  rs.forEach(({ m, steps }) => steps.slice(0, builder.step).forEach(s => s.draw.forEach(o => {
    if (o.square) add(o.square, m.color);
    else if (o.dot && o.label.startsWith(`${builder.item.name}(`)) add(o.dot, m.color);
  })));
  return out;
}

function bLinks(builder) {
  return bRowEnds(builder).flatMap(({ row, P, T, t }) => [
    { from: row.from, to: row.to, P, T, t, s: bS(builder, row.from) },
    { from: row.to, to: row.from, P: T, T: P, t: bBack(builder, row, t), s: bS(builder, row.to) },
  ]);
}

function bGlobalLines(builder) {
  return bShown(builder).map(n => {
    const p = bNodePos(builder, n);
    const r = bLen(p), phi = r < 1e-9 ? 0 : Math.acos(Math.max(-1, Math.min(1, p[1] / r)));
    let th = Math.atan2(p[2], p[0]);
    if (th < -1e-9) th += 2 * Math.PI;
    return `${n.name} in O's frame: r ${bF(r)}, φ ${bF(phi)} rad, θ ${bF(Math.max(0, th))} rad`;
  });
}
