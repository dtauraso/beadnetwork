function atobScene({ proj, curve, arrow, dot, label, triangle }, ctx, S, css) {
  const O = [0, 0, 0];
  const rad = d => d * Math.PI / 180;
  const add = (p, q) => p.map((v, i) => v + q[i]);
  const toCart = (R, ph, th) => [R * Math.sin(ph) * Math.cos(th), R * Math.cos(ph), R * Math.sin(ph) * Math.sin(th)];

  function grid(C, r, wire, width, alpha) {
    const on = (ph, th) => add(C, toCart(r, rad(ph), rad(th)));
    for (let ph = 30; ph <= 150; ph += 30) {
      const pts = []; for (let t = 0; t <= 360; t += 6) pts.push(on(ph, t));
      curve(pts, C, wire, width, alpha);
    }
    for (let th = 0; th < 180; th += 30) {
      const pts = []; for (let p = 0; p <= 360; p += 6) pts.push(on(p, th));
      curve(pts, C, wire, width, alpha);
    }
  }

  function circles(C, r, phi, theta, phiColor, thColor, width) {
    const on = (ph, th) => add(C, toCart(r, rad(ph), rad(th)));
    const phiPath = []; for (let p = 0; p <= 360; p += 2) phiPath.push(on(p, theta));
    curve(phiPath, C, phiColor, width);
    const thPath = []; for (let t = 0; t <= 360; t += 2) thPath.push(on(phi, t));
    curve(thPath, C, thColor, width);
  }

  function sphere(C, r, phi, theta, phiColor, thColor) {
    grid(C, r, css('--wire'), 0.8, 0.3);
    circles(C, r, phi, theta, phiColor, thColor, 4);
  }

  function sharedCircle(C, s) {
    const { u, h, rho } = s;
    const cross = (p, q) => [p[1] * q[2] - p[2] * q[1], p[2] * q[0] - p[0] * q[2], p[0] * q[1] - p[1] * q[0]];
    const e1 = cross(u, Math.abs(u[1]) < 0.9 ? [0, 1, 0] : [1, 0, 0]), n1 = Math.hypot(...e1);
    const f1 = e1.map(v => v / n1), f2 = cross(u, f1);
    const pts = [];
    for (let t = 0; t <= 360; t += 3) {
      const c = Math.cos(rad(t)), sn = Math.sin(rad(t));
      pts.push(add(C, u.map((v, i) => h * v + rho * (c * f1[i] + sn * f2[i]))));
    }
    ctx.setLineDash([12, 8]);
    curve(pts, add(C, s.m), css('--circle'), 4);
    ctx.setLineDash([]);
  }

  function sharedSphere(C, s, angles) {
    const centre = add(C, s.m);
    grid(centre, s.rho, css('--circle'), 1.2, 0.5);
    if (angles) circles(centre, s.rho, angles[1], angles[2], css('--circle'), css('--circle'), 2.5);
  }

  function vectors(pair, d, b) {
    const color = css(pair.color), n = pair.node;
    triangle([O, pair.C, b], color);
    arrow(O, pair.C, css('--muted'), 2, true);
    arrow(O, b, css('--global'), 3, true);
    if (pair.via) {
      const o = add(pair.C, pair.via.v);
      arrow(pair.C, o, color, 3, true);
      label(add(pair.C, pair.via.v.map(v => v / 2)), `2 d_${n}${pair.via.to}`, color, -12, -14);
      arrow(o, b, color, 2, true);
    }
    arrow(pair.C, b, color, 5);
    label(add(pair.C, d.map(v => v / 2)), `3 d_${n}b`, color, pair.side * 12, pair.side * 26);
    const [nx, ny] = proj(pair.C);
    ctx.beginPath(); ctx.arc(nx, ny, 0.14 * S, 0, Math.PI * 2);
    ctx.fillStyle = css('--card'); ctx.globalAlpha = 0.85; ctx.fill(); ctx.globalAlpha = 1;
    ctx.lineWidth = 6; ctx.strokeStyle = color; ctx.stroke();
    label(pair.C, `1 node ${n}`, color, pair.side * (0.14 * S + 8), 0.14 * S + 10);
    dot(pair.C, color);
  }

  return { grid, sphere, sharedCircle, sharedSphere, vectors };
}
