function bCellPoints(st, rs) {
  const out = [];
  const add = (p, color) => { if (bLen(p) > 1e-9 && !out.some(t => bLen(bSub(t.p, p)) < 1e-9)) out.push({ p, color }); };
  st.nodes.filter(n => n.known).forEach(n => add(bPlace(st, n.idx), '--ink'));
  rs.forEach(({ m, steps }) => steps.slice(0, st.step).forEach(s => s.draw.forEach(o => {
    if (o.square) add(o.square, m.color);
    else if (o.dot && o.label.startsWith(`${st.item}(`)) add(o.dot, m.color);
  })));
  return out;
}

function bLinks(st) {
  return bRowEnds(st).flatMap(({ row, P, T }) => [
    { from: row.from, to: row.to, P, T },
    { from: row.to, to: row.from, P: T, T: P },
  ]);
}

function bCell(st, P, T) {
  const c = bC(st), w = bSub(T, P), r = bLen(w), e = 1e-9;
  const th = r < e ? 0 : Math.atan2(w[2], w[0]) / c.theta;
  const I = r < e ? { r: 0, phi: 0, theta: 0 }
    : { r: r / c.r, phi: Math.acos(Math.max(-1, Math.min(1, w[1] / r))) / c.phi, theta: th < -e ? th + st.Ttheta : Math.max(0, th) };
  const K = Math.max(1, Math.ceil(I.r - e));
  const i = Math.min(Math.floor(I.phi + e), Math.ceil(st.Tphi / 2) - 1);
  const j = Math.floor(I.theta + e);
  return { I, K, i, j, R: K * st.cr };
}

function bCellCorners(st, P, T) {
  const c = bCell(st, P, T), cphi = 2 * Math.PI / st.Tphi, cth = 2 * Math.PI / st.Ttheta;
  return [[0, 0], [1, 0], [0, 1], [1, 1]].map(([di, dj]) => bAdd(P, bCart(c.R, (c.i + di) * cphi, (c.j + dj) * cth)));
}

function bGlobalLines(st) {
  const ends = bRowEnds(st);
  return st.nodes.map(n => {
    const p = bNodePos(st, n.name) || (ends.find(e => e.row.to === n.name) || {}).T;
    if (!p) return `${n.name} in O's frame: no position`;
    const r = bLen(p), phi = r < 1e-9 ? 0 : Math.acos(Math.max(-1, Math.min(1, p[1] / r)));
    let th = Math.atan2(p[2], p[0]);
    if (th < -1e-9) th += 2 * Math.PI;
    return `${n.name} in O's frame: r ${bF(r)}, φ ${bF(phi)} rad, θ ${bF(Math.max(0, th))} rad`;
  });
}

function bDrawCells(st, draw, ctx, css, links) {
  const cphi = 2 * Math.PI / st.Tphi, cth = 2 * Math.PI / st.Ttheta, wire = css('--wire'), ink = css('--ink');
  const line = (pts, color, width, alpha) => {
    const ps = pts.map(draw.proj);
    ctx.save();
    ctx.lineCap = 'round'; ctx.lineWidth = width; ctx.strokeStyle = color; ctx.globalAlpha = alpha;
    ctx.beginPath(); ctx.moveTo(ps[0][0], ps[0][1]);
    ps.slice(1).forEach(p => ctx.lineTo(p[0], p[1]));
    ctx.stroke();
    ctx.restore();
  };
  return links.map(({ from, to, P, T }) => {
    const at = (k, ph, th) => bAdd(P, bCart(k * st.cr, ph * cphi, th * cth));
    const arc = (f, color, width, alpha) => line(Array.from({ length: 17 }, (_, t) => f(t / 16)), color, width, alpha);
    const face = (k, i, j, color, width, alpha) => {
      [i, i + 1].forEach(ph => arc(t => at(k, ph, j + t), color, width, alpha));
      [j, j + 1].forEach(th => arc(t => at(k, i + t, th), color, width, alpha));
    };
    const rays = (k0, k1, i, j, color, width, alpha) => [[i, j], [i + 1, j], [i, j + 1], [i + 1, j + 1]]
      .forEach(([ph, th]) => line([at(k0, ph, th), at(k1, ph, th)], color, width, alpha));
    const c = bCell(st, P, T);
    rays(0, c.K, c.i, c.j, wire, 1, 0.5);
    for (let k = 1; k < c.K; k++) face(k, c.i, c.j, wire, 1, 0.5);
    face(c.K - 1, c.i, c.j, ink, 2, 0.8);
    face(c.K, c.i, c.j, ink, 2, 0.8);
    rays(c.K - 1, c.K, c.i, c.j, ink, 2, 0.8);
    return `${to} in ${from}'s frame: I_r ${bF(c.I.r)}, I_φ ${bF(c.I.phi)}, I_θ ${bF(c.I.theta)}`;
  });
}
