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
  [[t => bAdd(P, bCart(0.7 * L, t * q, 0)), 'φ', '--phi'], [t => bAdd(P, bCart(0.7 * L, q, t * q)), 'θ', '--theta']].forEach(([f, n, c]) => {
    const pts = Array.from({ length: 17 }, (_, i) => f(i / 16)), color = css(c);
    ctx.strokeStyle = color; ctx.lineWidth = 2; ctx.beginPath();
    pts.slice(0, 15).map(draw.proj).forEach(([x, y], i) => (i ? ctx.lineTo(x, y) : ctx.moveTo(x, y)));
    ctx.stroke();
    draw.arrow(pts[14], pts[16], color, 2);
    if (named) label(f(0.5), n, color);
  });
}

function bTicks(st, w) {
  const n = 12 * st.s, tick = 2 * Math.PI / n, r = bLen(w);
  if (r < 1e-9) return { phi: 0, theta: 0, r: 0 };
  const theta = Math.atan2(w[2], w[0]) / tick;
  return { phi: Math.acos(Math.max(-1, Math.min(1, w[1] / r))) / tick, theta: theta < -1e-9 ? theta + n : Math.max(0, theta), r: r / st.rho };
}

function bRowEnds(st) {
  return st.rows.map(row => {
    const P = bNodePos(st, row.from);
    if (!P) return null;
    const d = bRowVectors(st, row.to).find(v => v.row === row).d;
    return { row, P, d, T: bNodePos(st, row.to) || bAdd(P, d) };
  }).filter(Boolean);
}

function bSyncRows(st) {
  let changed = false;
  st.rows.forEach(row => [row.start, row.end].forEach(o => Object.keys(o).forEach(k => {
    if (Number.isInteger(o[k])) return;
    o[k] = Math.round(o[k]);
    changed = true;
  })));
  const placed = new Set();
  bRowEnds(st).forEach(({ row, P, d, T }) => {
    const n = st.nodes.find(x => x.name === row.to);
    if (!n || !n.known || placed.has(n.name)) return;
    placed.add(n.name);
    if (bLen(bSub(bAdd(P, d), T)) < 1e-9) return;
    n.idx = bIndexOf(st, bAdd(P, d));
    changed = true;
  });
  return changed;
}

function bDrawGuides(st, draw, ctx, css, points) {
  const tick = 2 * Math.PI / (12 * st.s), e = 1e-6, a = 0.45 * st.cr;
  const muted = css('--muted');
  const line = (pts, color, width, dash = []) => {
    ctx.save(); ctx.setLineDash(dash); ctx.lineWidth = width; ctx.strokeStyle = color; ctx.beginPath();
    pts.map(draw.proj).forEach(([x, y], k) => (k ? ctx.lineTo(x, y) : ctx.moveTo(x, y)));
    ctx.stroke(); ctx.restore();
  };
  const arc = (f, color) => line(Array.from({ length: 17 }, (_, k) => f(k / 16)), color, 3);
  points.forEach(({ p }) => bDrawAxes(draw, ctx, css, p, st.cr, false));
  bRowEnds(st).forEach(({ P, T }) => {
    const w = bSub(T, P), t = bTicks(st, w), r = bLen(w);
    if (r < e) return;
    arc(s => bAdd(P, bCart(a, s * t.phi * tick, t.theta * tick)), css('--phi'));
    arc(s => bAdd(P, bCart(a, Math.PI / 2, s * t.theta * tick)), css('--theta'));
    line([P, T], muted, 2, [3, 7]);
  });
}
