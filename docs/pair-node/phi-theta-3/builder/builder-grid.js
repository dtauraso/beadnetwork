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

function bCell(builder, t, s) {
  const n = 12 * s, e = 1e-9;
  const I = { r: t.r * builder.rho / builder.cr, phi: t.phi * builder.Tphi / n, theta: t.theta * builder.Ttheta / n };
  const K = Math.max(1, Math.ceil(I.r - e));
  return { I, K, i: Math.floor(I.phi + e) % builder.Tphi, j: Math.floor(I.theta + e) % builder.Ttheta, R: K * builder.cr };
}

function bCellCorners(builder, P, t, s) {
  const c = bCell(builder, t, s), cphi = 2 * Math.PI / builder.Tphi, cth = 2 * Math.PI / builder.Ttheta;
  return [[0, 0], [1, 0], [0, 1], [1, 1]].map(([di, dj]) => bAdd(P, bCart(c.R, (c.i + di) * cphi, (c.j + dj) * cth)));
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

function bDrawCells(builder, draw, ctx, css, links) {
  const cphi = 2 * Math.PI / builder.Tphi, cth = 2 * Math.PI / builder.Ttheta, wire = css('--wire'), ink = css('--ink');
  const line = (pts, color, width, alpha) => {
    const ps = pts.map(draw.proj);
    ctx.save();
    ctx.lineCap = 'round'; ctx.lineWidth = width; ctx.strokeStyle = color; ctx.globalAlpha = alpha;
    ctx.beginPath(); ctx.moveTo(ps[0][0], ps[0][1]);
    ps.slice(1).forEach(p => ctx.lineTo(p[0], p[1]));
    ctx.stroke();
    ctx.restore();
  };
  return links.map(({ from, to, P, t, s }) => {
    const at = (k, ph, th) => bAdd(P, bCart(k * builder.cr, ph * cphi, th * cth));
    const arc = (f, color, width, alpha) => line(Array.from({ length: 17 }, (_, t) => f(t / 16)), color, width, alpha);
    const face = (k, i, j, color, width, alpha) => {
      [i, i + 1].forEach(ph => arc(t => at(k, ph, j + t), color, width, alpha));
      [j, j + 1].forEach(th => arc(t => at(k, i + t, th), color, width, alpha));
    };
    const rays = (k0, k1, i, j, color, width, alpha) => [[i, j], [i + 1, j], [i, j + 1], [i + 1, j + 1]]
      .forEach(([ph, th]) => line([at(k0, ph, th), at(k1, ph, th)], color, width, alpha));
    const c = bCell(builder, t, s);
    rays(0, c.K, c.i, c.j, wire, 1, 0.5);
    for (let k = 1; k < c.K; k++) face(k, c.i, c.j, wire, 1, 0.5);
    face(c.K - 1, c.i, c.j, ink, 2, 0.8);
    face(c.K, c.i, c.j, ink, 2, 0.8);
    rays(c.K - 1, c.K, c.i, c.j, ink, 2, 0.8);
    return `${to.name} in ${from.name}'s frame: I_r ${bF(c.I.r)}, I_φ ${bF(c.I.phi)}, I_θ ${bF(c.I.theta)}`;
  });
}
