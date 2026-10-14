function bDrawAxes(draw, ctx, css, P, L, named) {
  const q = Math.PI / 2, muted = css('--muted');
  const label = (p, text, color) => {
    const [x, y] = draw.proj(p);
    ctx.fillStyle = color; ctx.font = '24px -apple-system, sans-serif'; ctx.fillText(text, x + 6, y - 6);
  };
  [[bAdd(P, bCart(L, 0, 0)), 'φ = 0'], [bAdd(P, bCart(L, q, 0)), 'θ = 0']].forEach(([e, n]) => {
    draw.arrow(P, e, muted, 2);
    if (named) label(e, n, muted);
  });
  if (named) [[t => bAdd(P, bCart(0.7 * L, t * q, 0)), 'φ', '--phi'], [t => bAdd(P, bCart(0.7 * L, q, t * q)), 'θ', '--theta']].forEach(([f, n, c]) => {
    const pts = Array.from({ length: 17 }, (_, i) => f(i / 16)), color = css(c);
    ctx.strokeStyle = color; ctx.lineWidth = 2; ctx.beginPath();
    pts.slice(0, 15).map(draw.proj).forEach(([x, y], i) => (i ? ctx.lineTo(x, y) : ctx.moveTo(x, y)));
    ctx.stroke();
    draw.arrow(pts[14], pts[16], color, 2);
    if (named) label(f(0.5), n, color);
  });
}

function bRowEnds(st) {
  return st.rows.map(row => {
    const P = bNodePos(st, row.from);
    if (!P) return null;
    const d = bRowVectors(st, row.to).find(v => v.row === row).d;
    return { row, P, d, T: bNodePos(st, row.to) || bAdd(P, d) };
  }).filter(Boolean);
}

function bPreset() {
  const p = {
    s: 1, rho: 0.5, Tphi: 12, Ttheta: 12, cr: 0.5,
    nodes: [
      { name: 'a', known: true, idx: { r: 2, phi: 3, theta: 0 } },
      { name: 'b', known: true, idx: { r: 0, phi: 0, theta: 0 } },
      { name: 'c', known: true, idx: { r: 0, phi: 0, theta: 0 } },
    ],
    rows: [
      { from: 'a', to: 'b', start: { phi: 2, theta: 0, r: 2 }, end: { phi: 0, theta: 0 } },
      { from: 'b', to: 'c', start: { phi: 6, theta: 0, r: 2 }, end: { phi: 0, theta: 0 } },
      { from: 'a', to: 'c', start: { phi: 4, theta: 0, r: 2 }, end: { phi: 0, theta: 0 } },
    ],
    item: 'b', on: { atob: true, pair: true }, step: 1,
  };
  const A = bPlace(p, p.nodes[0].idx);
  [[1, 0], [2, 2]].forEach(([i, r]) => { p.nodes[i].idx = bIndexOf(p, bAdd(A, bRowVectors(p, p.nodes[i].name).find(v => v.row === p.rows[r]).d)); });
  return p;
}

function bView(st, row) {
  const P = bNodePos(st, row.from), T = bNodePos(st, row.to);
  if (!P || !T) return null;
  const v = bSub(T, P), len = bLen(v), tick = 2 * Math.PI / (12 * st.s), half = 6 * st.s;
  if (len < 1e-9) return { phi: row.start.phi, theta: row.start.theta, r: 0 };
  let phi = Math.acos(Math.max(-1, Math.min(1, v[1] / len))) / tick;
  let theta = Math.hypot(v[0], v[2]) < 1e-9 * len ? row.start.theta : bWrap(st, Math.atan2(v[2], v[0]) / tick);
  if (row.start.phi > half) { phi = 2 * half - phi; theta = bWrap(st, theta + half); }
  return { phi, theta, r: len / st.rho };
}

function bSetStart(st, row, k, value) {
  const whole = x => Math.abs(x - Math.round(x)) < 1e-6;
  const name = { phi: 'φ', theta: 'θ', r: 'r' };
  const said = `${row.from}→${row.to} ${name[k]} ${bN(value)}`;
  if (!whole(value)) return `${said}: not a whole number, so it stays at ${bN(row.start[k])}`;
  const n = st.nodes.find(x => x.name === row.to), P = bNodePos(st, row.from);
  const idx = n && { ...n.idx };
  const start = { ...row.start, [k]: value };
  if (n && n.known && P) {
    const tick = 2 * Math.PI / (12 * st.s);
    const J = bIndexOf(st, bAdd(P, bCart(start.r * st.rho, start.phi * tick, start.theta * tick)));
    n.idx = { ...J, theta: J.theta < 0 ? J.theta + st.Ttheta : J.theta };
  }
  const others = st.rows.filter(r => r !== row && n && (r.from === n.name || r.to === n.name))
    .map(r => ({ r, t: bView(st, r) })).filter(o => o.t);
  const bad = others.filter(({ t }) => !['phi', 'theta', 'r'].every(q => whole(t[q])));
  if (bad.length) {
    if (n) n.idx = idx;
    return `${said}: ${bad.map(({ r, t }) => `${r.from}→${r.to} would be φ ${bF(t.phi)}, θ ${bF(t.theta)}, r ${bF(t.r)}`).join('; ')} — not whole ticks, so it stays at ${bN(row.start[k])}`;
  }
  row.start = start;
  others.forEach(({ r, t }) => { r.start = { phi: Math.round(t.phi) % (12 * st.s), theta: Math.round(t.theta) % (12 * st.s), r: Math.round(t.r) }; });
  return '';
}

function bDrawGuides(st, draw, ctx, css, points) {
  const tick = 2 * Math.PI / (12 * st.s), e = 1e-6, a = 0.7 * st.cr;
  const muted = css('--muted');
  const line = (pts, color, width, dash = []) => {
    ctx.save(); ctx.setLineDash(dash); ctx.lineWidth = width; ctx.strokeStyle = color; ctx.beginPath();
    pts.map(draw.proj).forEach(([x, y], k) => (k ? ctx.lineTo(x, y) : ctx.moveTo(x, y)));
    ctx.stroke(); ctx.restore();
  };
  const arc = (f, color) => {
    const pts = Array.from({ length: 49 }, (_, k) => f(k / 48));
    line(pts.slice(0, 47), color, 2);
    draw.arrow(pts[46], pts[48], color, 2);
  };
  points.forEach(({ p }) => bDrawAxes(draw, ctx, css, p, st.cr, false));
  const seen = {};
  bLinks(st).forEach(({ from, P, t }) => {
    const k = seen[from] = (seen[from] ?? -1) + 1, R = a * (1 + 0.4 * k);
    if (t.r < e) return;
    if (t.phi > e) arc(s => bAdd(P, bCart(R, s * t.phi * tick, 0)), css('--phi'));
    if (t.theta > e) arc(s => bAdd(P, bCart(R, Math.PI / 2, s * t.theta * tick)), css('--theta'));
  });
  bRowEnds(st).forEach(({ P, T }) => line([P, T], muted, 2, [3, 7]));
}
