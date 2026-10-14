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
  return [...st.rows].map(row => {
    const P = bNodePos(st, row.from), t = bDir(st, row);
    if (!P || !t) return null;
    const d = bVec(st, bS(st, row.from), t);
    return { row, P, d, T: bNodePos(st, row.to) || bAdd(P, d) };
  }).filter(Boolean);
}

function bPreset() {
  const a = bMake({ name: 'a', s: 1, known: true, idx: { r: 2, phi: 3, theta: 0 } });
  const b = bMake({ name: 'b', s: 1, known: true, idx: { r: 0, phi: 0, theta: 0 } });
  const c = bMake({ name: 'c', s: 1, known: true, idx: { r: 0, phi: 0, theta: 0 } });
  const p = {
    s: 1, rho: 0.5, Tphi: 12, Ttheta: 12, cr: 0.5,
    nodes: new Set([a, b, c]), rows: new Set(), made: 3,
    item: c, on: { atob: true, pair: true }, step: 1,
  };
  const rows = [
    { from: a, to: b, kind: 'angle', start: { phi: 2, theta: 0, r: 2 }, end: { phi: 0, theta: 0 } },
    { from: b, to: c, kind: 'angle', start: { phi: 4, theta: 0, r: 2 }, end: { phi: 0, theta: 0 } },
    { from: a, to: c, kind: 'copy', start: { phi: 0, theta: 0, r: 0 }, end: { phi: 0, theta: 0 } },
  ];
  rows.forEach(r => bAddRow(p, r));
  rows.slice(0, 2).forEach(r => bRefer(p, r));
  bSend(p, a);
  bSetCopy(p, rows[2], '');
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
  return rows.filter(r => !bFloat(r)).flatMap(r => {
    const a = bDir(st, r), b = bBack(st, r), out = [], own = bKind(r) === 'change';
    const ends = own ? `, own φ ${bF(r.end.phi)}, own θ ${bF(r.end.theta)}` : '';
    if (![a.phi, a.theta, a.r, ...(own ? [r.end.phi, r.end.theta] : [])].every(bWhole)) out.push(`${r.from.name}→${r.to.name} at ${r.from.name} (s ${bS(st, r.from)}) would be φ ${bF(a.phi)}, θ ${bF(a.theta)}, r ${bF(a.r)}${ends}`);
    if (![b.phi, b.theta].every(bWhole)) out.push(`${r.from.name} seen from ${r.to.name} (s ${bS(st, r.to)}) would be φ ${bF(b.phi)}, θ ${bF(b.theta)}`);
    return out;
  });
}

function bSettle(st, keep, rows, said, was) {
  const bad = bOffTicks(st, rows);
  if (bad.length) {
    bRestore(keep);
    return `${said}: ${bad.join('; ')} — not whole ticks, so it stays at ${was}`;
  }
  bDone();
  rows.filter(r => !bFloat(r)).forEach(r => {
    const w = t => bWrap(bS(st, r.from), Math.round(t));
    r.start = { phi: w(r.start.phi), theta: w(r.start.theta), r: Math.round(r.start.r) };
    r.end = { phi: w(r.end.phi), theta: w(r.end.theta) };
  });
  return '';
}

function bSetStart(st, row, k, value, part = 'start') {
  const was = bN(row[part][k]), said = `${row.from.name}→${row.to.name} ${part === 'end' ? 'own ' : ''}${{ phi: 'φ', theta: 'θ', r: 'r' }[k]} ${bN(value)}`;
  if (!bWhole(value)) return `${said}: not a whole number, so it stays at ${was}`;
  const keep = bKeep();
  bSave(row);
  row[part][k] = value;
  if (!bMove(st, row)) {
    bRestore(keep);
    return `${said}: the nodes' references run in a loop, so it stays at ${was}`;
  }
  return bSettle(st, keep, [row], said, was);
}

function bSetS(st, node, value) {
  const was = bS(st, node), said = `${node.name}'s s ${bN(value)}`;
  if (!bWhole(value) || value < 1) return `${said}: not a whole number of at least 1, so it stays at ${was}`;
  const keep = bKeep(), k = value / was;
  bSave(node);
  node.s = value;
  const rows = [...node.rowsOut, ...node.rowsIn];
  node.rowsOut.forEach(r => {
    bSave(r);
    r.start = { ...r.start, phi: r.start.phi * k, theta: r.start.theta * k };
    r.end = { phi: r.end.phi * k, theta: r.end.theta * k };
    if (r.base) r.base = { ...r.base, phi: r.base.phi * k, theta: r.base.theta * k };
    if (r.ref) r.ref = { phi: r.ref.phi * k, theta: r.ref.theta * k };
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
    const k = seen[from.name] = (seen[from.name] ?? -1) + 1, R = a * (1 + 0.4 * k), tick = 2 * Math.PI / (12 * ns);
    if (t.r < e) return;
    if (t.phi > e) arc(s => bAdd(P, bCart(R, s * t.phi * tick, 0)), css('--phi'));
    if (t.theta > e) arc(s => bAdd(P, bCart(R, Math.PI / 2, s * t.theta * tick)), css('--theta'));
  });
  bRowEnds(st).forEach(({ P, T }) => line([P, T], muted, 2, [3, 7]));
}
