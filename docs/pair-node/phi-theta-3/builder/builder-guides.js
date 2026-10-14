function bDrawGuides(st, draw, ctx, css, points) {
  const c = bC(st), e = 1e-6, top = st.Tphi / 2;
  const path = (f, from, to, color, width, dash = []) => {
    const n = Math.max(2, Math.ceil(Math.abs(to - from) * 8));
    const ps = Array.from({ length: n + 1 }, (_, k) => draw.proj(f(from + (to - from) * k / n)));
    ctx.save();
    ctx.setLineDash(dash); ctx.lineWidth = width; ctx.strokeStyle = color;
    ctx.beginPath(); ps.forEach(([x, y], k) => (k ? ctx.lineTo(x, y) : ctx.moveTo(x, y))); ctx.stroke();
    ctx.restore();
  };
  const label = (p, text, color, size) => {
    const [x, y] = draw.proj(p);
    ctx.fillStyle = color; ctx.font = `${size}px -apple-system, sans-serif`; ctx.fillText(text, x + 6, y - 6);
  };
  const near = v => [...new Set([Math.floor(v + e), Math.ceil(v - e)])];
  const muted = css('--muted'), phi = css('--phi'), theta = css('--theta');
  points.forEach(({ p }) => {
    const I = bIndexOf(st, p), R = I.r * c.r;
    if (R < e) return;
    const merid = th => i => bCart(R, i * c.phi, th);
    const lat = ph => j => bCart(R, ph, j * c.theta);
    near(I.theta).forEach(j => {
      path(merid(j * c.theta), Math.max(0, I.phi - 1), Math.min(top, I.phi + 1), muted, 1.5, [3, 7]);
      label(bCart(R, Math.min(top, I.phi + 1) * c.phi, j * c.theta), `I_θ ${j}`, muted, 18);
    });
    near(I.phi).forEach(i => {
      if (i <= e || i >= top - e) return;
      path(lat(i * c.phi), I.theta - 1, I.theta + 1, muted, 1.5, [3, 7]);
      label(bCart(R, i * c.phi, (I.theta + 1) * c.theta), `I_φ ${i}`, muted, 18);
    });
    path(t => bCart(t, I.phi * c.phi, I.theta * c.theta), 0, R, muted, 1.5);
    label(bCart(R / 2, I.phi * c.phi, I.theta * c.theta), `I_r ${bF(I.r)}`, muted, 20);
    if (I.phi > e) {
      path(merid(I.theta * c.theta), 0, I.phi, phi, 2);
      label(bCart(R, I.phi / 2 * c.phi, I.theta * c.theta), `I_φ ${bF(I.phi)}`, phi, 20);
    }
    if (Math.abs(I.theta) > e && I.phi > e && I.phi < top - e) {
      path(lat(I.phi * c.phi), 0, I.theta, theta, 2);
      label(bCart(R, I.phi * c.phi, I.theta / 2 * c.theta), `I_θ ${bF(I.theta)}`, theta, 20);
    }
  });
}
