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

function bSyncRows(st) {
  let changed = false;
  st.rows.forEach(row => [row.start, row.end].forEach(o => Object.keys(o).forEach(k => {
    if (Number.isInteger(o[k])) return;
    o[k] = Math.round(o[k]);
    changed = true;
  })));
  const first = st.nodes.find(n => n.known);
  if (!first) return changed;
  const pos = { [first.name]: bPlace(st, first.idx) };
  const ds = st.nodes.flatMap(n => bRowVectors(st, n.name));
  for (let more = true; more;) {
    more = false;
    ds.forEach(({ row, d }) => {
      if (pos[row.from] && !pos[row.to]) { pos[row.to] = bAdd(pos[row.from], d); more = true; }
      if (pos[row.to] && !pos[row.from]) { pos[row.from] = bSub(pos[row.to], d); more = true; }
    });
  }
  st.nodes.filter(n => n.known && n !== first && pos[n.name]).forEach(n => {
    if (bLen(bSub(bPlace(st, n.idx), pos[n.name])) < 1e-9) return;
    const I = bIndexOf(st, pos[n.name]);
    n.idx = { ...I, theta: I.theta < 0 ? I.theta + st.Ttheta : I.theta };
    changed = true;
  });
  return changed;
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
  bLinks(st).forEach(({ P, t }) => {
    if (t.r < e) return;
    if (t.phi > e) arc(s => bAdd(P, bCart(a, s * t.phi * tick, t.theta * tick)), css('--phi'));
    if (t.theta > e) arc(s => bAdd(P, bCart(a, Math.PI / 2, s * t.theta * tick)), css('--theta'));
  });
  bRowEnds(st).forEach(({ P, T }) => line([P, T], muted, 2, [3, 7]));
}
