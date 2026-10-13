function bGridTargets(st, rs) {
  const out = [];
  const add = (key, label, p) => { if (bLen(p) > 1e-9 && !out.some(t => bLen(bSub(t.p, p)) < 1e-9)) out.push({ key, label, p }); };
  rs.forEach(({ steps }) => steps.forEach(s => s.draw.forEach(o => {
    if (o.square) add(`result:${o.label}`, `${o.label} (placed)`, o.square);
    else if (o.dot && o.label.startsWith(`${st.item}(`)) add(`result:${o.label}`, `${o.label} (placed)`, o.dot);
  })));
  st.nodes.filter(n => n.known).forEach(n => add(`node:${n.name}`, `node ${n.name}`, bPlace(st, n.idx)));
  return out;
}

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
  const arc = (f, color, width, alpha) => {
    const pts = []; for (let t = 0; t <= 16; t++) pts.push(f(t / 16));
    draw.curve(pts, [0, 0, 0], color, width, alpha);
  };
  const face = (k, i, j, color, width, alpha) => {
    [i, i + 1].forEach(ph => arc(t => at(k, ph, j + t), color, width, alpha));
    [j, j + 1].forEach(th => arc(t => at(k, i + t, th), color, width, alpha));
  };
  const rays = (k0, k1, i, j, color, width, alpha) => [[i, j], [i + 1, j], [i, j + 1], [i + 1, j + 1]]
    .forEach(([ph, th]) => draw.curve([at(k0, ph, th), at(k1, ph, th)], [0, 0, 0], color, width, alpha));
  return points.map(({ p, color }) => {
    const c = bCell(st, p), col = css(color);
    rays(0, c.K, c.i, c.j, wire, 1, 0.5);
    for (let k = 1; k < c.K; k++) face(k, c.i, c.j, wire, 1, 0.5);
    face(c.K - 1, c.i, c.j, col, 2.5, 1);
    face(c.K, c.i, c.j, col, 2.5, 1);
    rays(c.K - 1, c.K, c.i, c.j, col, 2.5, 1);
    return `cell I_r ${c.K - 1}–${c.K}, I_φ ${c.i}–${c.i + 1}, I_θ ${c.j}–${c.j + 1} holds (${bF(c.I.r)}, ${bF(c.I.phi)}, ${bF(c.I.theta)})`;
  });
}

function bDrawGrid(st, draw, ctx, wire, p) {
  const R = bLen(p), cphi = 2 * Math.PI / st.Tphi, cth = 2 * Math.PI / st.Ttheta;
  const kphi = Math.ceil(st.Tphi / 72), kth = Math.ceil(st.Ttheta / 72);
  const on = (ph, th) => bCart(R, ph, th);
  for (let i = kphi; i * cphi < Math.PI - 1e-9; i += kphi) {
    const pts = []; for (let t = 0; t <= 96; t++) pts.push(on(i * cphi, t * 2 * Math.PI / 96));
    draw.curve(pts, [0, 0, 0], wire, 1.2, 0.6);
  }
  for (let j = 0; j < st.Ttheta; j += kth) {
    const pts = []; for (let t = 0; t <= 48; t++) pts.push(on(t * Math.PI / 48, j * cth));
    draw.curve(pts, [0, 0, 0], wire, 1.2, 0.6);
  }
  const u = p.map(v => v / R);
  ctx.fillStyle = wire;
  for (let k = 1; k * st.cr <= R + 1e-9; k++) {
    const [x, y] = draw.proj(u.map(v => v * k * st.cr));
    ctx.beginPath(); ctx.arc(x, y, 4, 0, Math.PI * 2); ctx.fill();
  }
  const I = bIndexOf(st, p);
  return `radius |C − O| = ${bF(R)}: I_r ${bF(I.r)}, I_φ ${bF(I.phi)}, I_θ ${bF(I.theta)}` +
    `; a line every ${kphi > 1 ? `${kphi} φ steps` : 'φ step'} and ${kth > 1 ? `${kth} θ steps` : 'θ step'}; dots on the ray: each c_r`;
}
