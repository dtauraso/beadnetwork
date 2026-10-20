function bCellPoints(builder, rs) {
  const out = [];
  const had = new Set();
  const add = (p, color) => {
    const key = p.map(v => v.toFixed(6)).join();
    if (bLen(p) > 1e-9 && !had.has(key)) { had.add(key); out.push({ p, color }); }
  };
  bShown(builder).forEach(n => { if (n.known) add(bNodePos(builder, n), '--ink'); });
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
  const ends = new Map(bRowEnds(builder).map(e => [e.ab, e.T]));
  return bShown(builder).map(n => {
    const p = bNodePos(builder, n) || ends.get(n.via);
    if (!p) return `${n.name} in O's frame: no position`;
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

function bSphereBoxes(builder, h, changed) {
  const on = builder.spheres ||= {}, names = bShown(builder).map(n => n.name);
  const box = (label, checked, set) => {
    const b = h('input', { type: 'checkbox', checked });
    b.onchange = () => { set(b.checked); changed(); };
    return h('label', {}, [b, ` ${label}`]);
  };
  return [h('span', { textContent: 'sphere at' }),
    box('all', names.length > 0 && names.every(n => on[n]), v => names.forEach(n => { on[n] = v; })),
    ...names.map(n => box(n, !!on[n], v => { on[n] = v; }))];
}

function bDrawSpheres(builder, draw, ctx) {
  const P = bShown(builder).filter(n => n.known).map(n => bNodePos(builder, n));
  const cap = builder.rho * 2, cell = p => p.map(v => Math.floor(v / cap)), grid = new Map();
  let near = cap;
  P.forEach(p => {
    const [x, y, z] = cell(p);
    for (const dx of [-1, 0, 1]) for (const dy of [-1, 0, 1]) for (const dz of [-1, 0, 1])
      (grid.get(`${x + dx},${y + dy},${z + dz}`) || []).forEach(q => { near = Math.min(near, bLen(bSub(p, q))); });
    const key = `${x},${y},${z}`;
    if (!grid.has(key)) grid.set(key, []);
    grid.get(key).push(p);
  });
  const R = near / 2;
  const ring = (C, f) => {
    ctx.beginPath();
    for (let k = 0; k <= 48; k++) { const [x, y] = draw.proj(bAdd(C, f(k / 48 * 2 * Math.PI))); k ? ctx.lineTo(x, y) : ctx.moveTo(x, y); }
    ctx.stroke();
  };
  ctx.strokeStyle = 'hsl(210 60% 55%)'; ctx.lineWidth = 1.5;
  bShown(builder).filter(n => n.known && (builder.spheres || {})[n.name]).map(n => bNodePos(builder, n)).forEach(C => {
    for (let m = 0; m < 6; m++) ring(C, t => bCart(R, t, m * Math.PI / 6));
    for (let l = 1; l < 6; l++) ring(C, t => bCart(R, l * Math.PI / 6, t));
  });
}
