const ANGLE_ARC_REACH = { theta: 0.46, phi: 0.64 };
const ANGLE_ARC_STEPS = 28;

function normalize3(v) {
  const l = Math.hypot(v.x, v.y, v.z) || 1;
  return { x: v.x / l, y: v.y / l, z: v.z / l };
}

function slerp(u, v, ang, t) {
  const s = Math.sin(ang);
  if (s < 1e-9) return u;
  const a = Math.sin((1 - t) * ang) / s, b = Math.sin(t * ang) / s;
  return normalize3({
    x: u.x * a + v.x * b,
    y: u.y * a + v.y * b,
    z: u.z * a + v.z * b,
  });
}

function angleArc(g, ball, from, to, reach, glyph) {
  const dot = Math.max(-1, Math.min(1, from.x * to.x + from.y * to.y + from.z * to.z));
  const ang = Math.acos(dot);

  const at = (t) => {
    const seat = seatFromDir(slerp(from, to, ang, t));
    return ballPointAt(ball, seat.a, seat.b, ball.r * reach);
  };

  if (ang >= 1e-6) {
    const pts = [];
    for (let k = 0; k <= ANGLE_ARC_STEPS; k++) {
      const [x, y] = at(k / ANGLE_ARC_STEPS);
      pts.push(`${k ? 'L' : 'M'}${x.toFixed(2)} ${y.toFixed(2)}`);
    }
    g.appendChild(tag('path', { d: pts.join(' '), class: 'ring-angle' }));
  }

  const [lx, ly, lz] = at(0.5);
  const t = tag('text', { x: lx, y: ly - 5, class: `ring-angle-label${backish(lz)}` });
  t.textContent = `${(ang / Math.PI).toFixed(2)}π ${glyph}`;
  g.appendChild(t);
}

function sphereAngles(g, ball, spec, incoming) {
  for (const which of ['theta', 'phi']) {
    if (!spec[which] || spec[which].axis === undefined) continue;
    const [a, b] = sphereAt(spec.points, spec[which].axis, which);
    angleArc(g, ball, incoming, unitPoint(a, b), ANGLE_ARC_REACH[which],
      sphereGlyph(which, ball.sub));
  }
}
