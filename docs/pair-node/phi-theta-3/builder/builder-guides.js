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
  return [...st.trees].flatMap(tree => bDirs(st, tree).map(({ t: was }) => {
    const { ab } = tree, row = ab.row, P = bNodePos(st, ab.a);
    if (!P) return null;
    const t = B_AT && (B_AT.has(ab.a) || B_AT.has(ab.b)) ? bView(st, row) || was : was;
    const d = bVec(st, bS(st, ab.a), t);
    return { ab, row, P, d, t, T: bNodePos(st, ab.b) || bAdd(P, d) };
  })).filter(Boolean);
}

function bPreset() {
  const a = bMake({ name: 'a', s: 1, known: true, idx: { r: 2, phi: 3, theta: 0 } });
  const b = bMake({ name: 'b', s: 1, known: true, idx: { r: 0, phi: 0, theta: 0 } });
  const c = bMake({ name: 'c', s: 1, known: true, idx: { r: 0, phi: 0, theta: 0 } });
  const p = {
    s: 1, rho: 0.5, Tphi: 12, Ttheta: 12, cr: 0.5,
    nodes: new Set([a, b, c]), trees: new Set(), picked: [], made: 3,
    item: c, on: { atob: true, pair: true }, step: 1,
  };
  const rows = [
    { from: a, to: b, start: { phi: 2, theta: 0, r: 2 } },
    { from: b, to: c, start: { phi: 4, theta: 0, r: 2 } },
  ];
  const [ab, bc] = rows.map(r => bAddTree(p, r));
  [ab, bc].forEach(t => { bRefer(p, t.ab); bPut(p, t.ab.row); });
  return p;
}

function bView(st, row, P = bNodePos(st, row.from), T = bNodePos(st, row.to)) {
  if (!P || !T) return null;
  const s = bS(st, row.from), v = bSub(T, P), len = bLen(v), tick = 2 * Math.PI / (12 * s), half = 6 * s;
  const was = bDir(st, row) || { phi: 0, theta: 0 };
  if (len < 1e-9) return { phi: was.phi, theta: was.theta, r: 0 };
  let phi = Math.acos(Math.max(-1, Math.min(1, v[1] / len))) / tick;
  let theta = Math.hypot(v[0], v[2]) < 1e-9 * len ? was.theta : bWrap(s, Math.atan2(v[2], v[0]) / tick);
  if (was.phi > half) { phi = 2 * half - phi; theta = bWrap(s, theta + half); }
  return { phi, theta, r: len / st.rho };
}

function bSetStart(st, tree, k, value) {
  tree.ab.row.start[k] = value;
  bMove(st, tree);
  return '';
}

function bSetS(st, node, value) {
  const was = bS(st, node), k = value / was;
  if (!bWhole(value) || value < 1) return `${node.name}'s s ${bN(value)}: not a whole number of at least 1, so it stays at ${was}`;
  node.s = value;
  node.out.forEach(t => {
    const r = t.ab.row;
    r.start = { ...r.start, phi: r.start.phi * k, theta: r.start.theta * k };
  });
  return '';
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
