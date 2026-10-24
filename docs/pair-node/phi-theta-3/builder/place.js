const bC = builder => ({ r: builder.cr, phi: 2 * Math.PI / builder.Tphi, theta: 2 * Math.PI / builder.Ttheta });

function bPlace(builder, I) {
  const c = bC(builder);
  return bAdd(B_O, bCart(I.r * c.r, I.phi * c.phi, I.theta * c.theta));
}

function bIndexOf(builder, p) {
  const c = bC(builder), w = bSub(p, B_O), r = bLen(w);
  if (r < 1e-9) return { r: 0, phi: 0, theta: 0 };
  return { r: r / c.r, phi: Math.acos(Math.max(-1, Math.min(1, w[1] / r))) / c.phi, theta: Math.atan2(w[2], w[0]) / c.theta };
}

const bS = (builder, n) => (n && n.s) || builder.s;
const bNodePos = (builder, n) => (n ? bPlace(builder, n.idx) : null);

function bVec(builder, s, t) {
  const ang = x => x * 2 * Math.PI / (12 * s);
  const R = t.r * builder.rho, ph = ang(t.phi), th = ang(t.theta);
  return [R * Math.sin(ph) * Math.abs(Math.cos(th)), R * Math.cos(ph), R * Math.sin(ph) * Math.sin(th)];
}

function bSetAt(builder, n, p) {
  const J = bIndexOf(builder, p);
  n.idx = { ...J, theta: J.theta < 0 ? J.theta + builder.Ttheta : J.theta };
}
