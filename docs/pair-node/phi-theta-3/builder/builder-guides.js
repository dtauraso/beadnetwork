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
    const P = bNodePos(st, row.from), v = bRowVectors(st, row.to).find(x => x.row === row);
    if (!P || !v) return null;
    return { row, P, d: v.d, T: bNodePos(st, row.to) || bAdd(P, v.d) };
  }).filter(Boolean);
}

function bPreset() {
  const p = {
    s: 1, rho: 0.5, Tphi: 12, Ttheta: 12, cr: 0.5,
    nodes: [
      { name: 'a', s: 1, known: true, idx: { r: 2, phi: 3, theta: 0 } },
      { name: 'b', s: 1, known: true, idx: { r: 0, phi: 0, theta: 0 } },
      { name: 'c', s: 1, known: true, idx: { r: 0, phi: 0, theta: 0 } },
    ],
    rows: [
      { from: 'a', to: 'b', kind: 'angle', start: { phi: 2, theta: 0, r: 2 }, end: { phi: 0, theta: 0 } },
      { from: 'b', to: 'c', kind: 'change', start: { phi: 0, theta: 0, r: 2 }, end: { phi: 2, theta: 0 } },
      { from: 'a', to: 'c', kind: 'base', start: { phi: 0, theta: 0, r: 0 }, end: { phi: 0, theta: 0 } },
    ],
    item: 'c', on: { atob: true, pair: true }, step: 1,
  };
  [[1, 0], [2, 1]].forEach(([i, r]) => {
    const row = p.rows[r], P = bPlace(p, p.nodes.find(n => n.name === row.from).idx);
    p.nodes[i].idx = bIndexOf(p, bAdd(P, bRowVectors(p, row.to).find(v => v.row === row).d));
  });
  p.rows[2].base = bView(p, p.rows[2]);
  return p;
}

function bView(st, row) {
  const P = bNodePos(st, row.from), T = bNodePos(st, row.to);
  if (!P || !T) return null;
  const s = bS(st, row.from), v = bSub(T, P), len = bLen(v), tick = 2 * Math.PI / (12 * s), half = 6 * s;
  const was = bDir(st, row) || { phi: 0, theta: 0 };
  if (len < 1e-9) return { phi: was.phi, theta: was.theta, r: 0 };
  let phi = Math.acos(Math.max(-1, Math.min(1, v[1] / len))) / tick;
  let theta = Math.hypot(v[0], v[2]) < 1e-9 * len ? was.theta : bWrap(s, Math.atan2(v[2], v[0]) / tick);
  if (was.phi > half) { phi = 2 * half - phi; theta = bWrap(s, theta + half); }
  return { phi, theta, r: len / st.rho };
}

function bOffTicks(st, rows) {
  return rows.filter(r => bKind(r) !== 'base').flatMap(r => {
    const a = bDir(st, r), b = bBack(st, r), out = [], own = bKind(r) === 'change';
    const ends = own ? `, own φ ${bF(r.end.phi)}, own θ ${bF(r.end.theta)}` : '';
    if (![a.phi, a.theta, a.r, ...(own ? [r.end.phi, r.end.theta] : [])].every(bWhole)) out.push(`${r.from}→${r.to} at ${r.from} (s ${bS(st, r.from)}) would be φ ${bF(a.phi)}, θ ${bF(a.theta)}, r ${bF(a.r)}${ends}`);
    if (![b.phi, b.theta].every(bWhole)) out.push(`${r.from} seen from ${r.to} (s ${bS(st, r.to)}) would be φ ${bF(b.phi)}, θ ${bF(b.theta)}`);
    return out;
  });
}

function bSettle(st, keep, rows, said, was) {
  const bad = bOffTicks(st, rows);
  if (bad.length) {
    [st.nodes, st.rows] = JSON.parse(keep);
    return `${said}: ${bad.join('; ')} — not whole ticks, so it stays at ${was}`;
  }
  rows.filter(r => bKind(r) !== 'base').forEach(r => {
    const w = t => bWrap(bS(st, r.from), Math.round(t));
    r.start = { phi: w(r.start.phi), theta: w(r.start.theta), r: Math.round(r.start.r) };
    r.end = { phi: w(r.end.phi), theta: w(r.end.theta) };
  });
  return '';
}

function bSetStart(st, row, k, value, part = 'start') {
  const was = bN(row[part][k]), said = `${row.from}→${row.to} ${part === 'end' ? 'own ' : ''}${{ phi: 'φ', theta: 'θ', r: 'r' }[k]} ${bN(value)}`;
  if (!bWhole(value)) return `${said}: not a whole number, so it stays at ${was}`;
  const keep = JSON.stringify([st.nodes, st.rows]);
  row[part][k] = value;
  const n = st.nodes.find(x => x.name === row.to), P = bNodePos(st, row.from);
  let rows = [row];
  if (n && n.known && P) {
    const J = bIndexOf(st, bAdd(P, bRowVectors(st, row.to).find(v => v.row === row).d));
    n.idx = { ...J, theta: J.theta < 0 ? J.theta + st.Ttheta : J.theta };
    rows = st.rows.filter(r => r.from === n.name || r.to === n.name);
    rows.filter(r => r !== row).forEach(r => bReview(st, r));
  }
  return bSettle(st, keep, rows, said, was);
}

function bSetS(st, node, value) {
  const was = bS(st, node.name), said = `${node.name}'s s ${bN(value)}`;
  if (!bWhole(value) || value < 1) return `${said}: not a whole number of at least 1, so it stays at ${was}`;
  const keep = JSON.stringify([st.nodes, st.rows]), k = value / was;
  node.s = value;
  const rows = st.rows.filter(r => r.from === node.name || r.to === node.name);
  rows.filter(r => r.from === node.name).forEach(r => {
    r.start = { ...r.start, phi: r.start.phi * k, theta: r.start.theta * k };
    r.end = { phi: r.end.phi * k, theta: r.end.theta * k };
    if (r.base) r.base = { ...r.base, phi: r.base.phi * k, theta: r.base.theta * k };
  });
  return bSettle(st, keep, rows, said, was);
}

function bDrawGuides(st, draw, ctx, css, points) {
  const e = 1e-6, a = 0.7 * st.cr;
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
  bLinks(st).forEach(({ from, P, t, s: ns }) => {
    const k = seen[from] = (seen[from] ?? -1) + 1, R = a * (1 + 0.4 * k), tick = 2 * Math.PI / (12 * ns);
    if (t.r < e) return;
    if (t.phi > e) arc(s => bAdd(P, bCart(R, s * t.phi * tick, 0)), css('--phi'));
    if (t.theta > e) arc(s => bAdd(P, bCart(R, Math.PI / 2, s * t.theta * tick)), css('--theta'));
  });
  bRowEnds(st).forEach(({ P, T }) => line([P, T], muted, 2, [3, 7]));
}
