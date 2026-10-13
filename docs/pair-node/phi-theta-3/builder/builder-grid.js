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
