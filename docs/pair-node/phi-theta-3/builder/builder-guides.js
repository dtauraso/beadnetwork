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

function bDrawGuides(st, draw, ctx, css, points) {
  const tick = 2 * Math.PI / (12 * st.s), e = 1e-6, muted = css('--muted');
  const near = v => [...new Set([Math.floor(v + e), Math.ceil(v - e)])];
  points.forEach(({ p }) => {
    bDrawAxes(draw, ctx, css, p, st.cr, false);
    points.forEach(({ p: q }) => {
      const w = bSub(q, p), r = bLen(w);
      if (r < e) return;
      const phi = Math.acos(Math.max(-1, Math.min(1, w[1] / r))) / tick, theta = Math.atan2(w[2], w[0]) / tick;
      near(phi).forEach(i => near(theta).forEach(j => {
        const [x0, y0] = draw.proj(p), end = bAdd(p, bCart(r, i * tick, j * tick)), [x1, y1] = draw.proj(end);
        ctx.save();
        ctx.setLineDash([3, 7]); ctx.lineWidth = 1.5; ctx.strokeStyle = muted;
        ctx.beginPath(); ctx.moveTo(x0, y0); ctx.lineTo(x1, y1); ctx.stroke();
        ctx.restore();
        ctx.fillStyle = muted; ctx.font = '18px -apple-system, sans-serif'; ctx.fillText(`φ ${i} θ ${((j % (12 * st.s)) + 12 * st.s) % (12 * st.s)}`, x1 + 6, y1 + 18);
      }));
    });
  });
}
