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

function bCell(st, p) {
  const I = bIndexOf(st, p), e = 1e-9;
  const K = Math.max(1, Math.ceil(I.r - e));
  const i = Math.min(Math.floor(I.phi + e), Math.ceil(st.Tphi / 2) - 1);
  const j = Math.floor(I.theta + e);
  return { I, K, i, j, R: K * st.cr };
}

function bCellCorners(st, p) {
  const c = bCell(st, p), cphi = 2 * Math.PI / st.Tphi, cth = 2 * Math.PI / st.Ttheta;
  return [[0, 0], [1, 0], [0, 1], [1, 1]].map(([di, dj]) => bCart(c.R, (c.i + di) * cphi, (c.j + dj) * cth));
}

function bDrawCells(st, draw, ctx, css, points) {
  const cphi = 2 * Math.PI / st.Tphi, cth = 2 * Math.PI / st.Ttheta, wire = css('--wire');
  const at = (k, ph, th) => bCart(k * st.cr, ph * cphi, th * cth);
  const dotted = (pts, color) => {
    const ps = pts.map(draw.proj);
    ctx.save();
    ctx.setLineDash([3, 7]); ctx.lineCap = 'round'; ctx.lineWidth = 1.5; ctx.strokeStyle = color; ctx.globalAlpha = 0.8;
    ctx.beginPath(); ctx.moveTo(ps[0][0], ps[0][1]);
    ps.slice(1).forEach(p => ctx.lineTo(p[0], p[1]));
    ctx.stroke();
    ctx.restore();
  };
  const arc = (f, color, width, alpha) => {
    const pts = []; for (let t = 0; t <= 16; t++) pts.push(f(t / 16));
    if (width === 'dotted') dotted(pts, color); else draw.curve(pts, [0, 0, 0], color, width, alpha);
  };
  const face = (k, i, j, color, width, alpha) => {
    [i, i + 1].forEach(ph => arc(t => at(k, ph, j + t), color, width, alpha));
    [j, j + 1].forEach(th => arc(t => at(k, i + t, th), color, width, alpha));
  };
  const rays = (k0, k1, i, j, color, width, alpha) => [[i, j], [i + 1, j], [i, j + 1], [i + 1, j + 1]]
    .forEach(([ph, th]) => width === 'dotted'
      ? dotted([at(k0, ph, th), at(k1, ph, th)], color)
      : draw.curve([at(k0, ph, th), at(k1, ph, th)], [0, 0, 0], color, width, alpha));
  return points.map(({ p, color }) => {
    const c = bCell(st, p), col = css(color);
    if (c.K > 1) rays(0, c.K - 1, c.i, c.j, wire, 'dotted');
    for (let k = 1; k < c.K - 1; k++) face(k, c.i, c.j, wire, 'dotted');
    face(c.K - 1, c.i, c.j, col, 2.5, 1);
    face(c.K, c.i, c.j, col, 2.5, 1);
    rays(c.K - 1, c.K, c.i, c.j, col, 2.5, 1);
    return `cell I_r ${c.K - 1}–${c.K}, I_φ ${c.i}–${c.i + 1}, I_θ ${c.j}–${c.j + 1} holds (${bF(c.I.r)}, ${bF(c.I.phi)}, ${bF(c.I.theta)})`;
  });
}
