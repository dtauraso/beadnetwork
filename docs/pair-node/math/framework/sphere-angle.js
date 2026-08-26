const ANGLE_ARC_REACH = 0.55;
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

function angleArc(g, ball, from, to, reach, glyph, sign) {
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
  t.textContent = `${sign < 0 ? '−' : ''}${(ang / Math.PI).toFixed(2)}π ${glyph}`;
  g.appendChild(t);
}

const RING_NORMAL = {
  theta: { x: 0, y: 1, z: 0 },
  phi: { x: 1, y: 0, z: 0 },
};

function inRingPlane(from, which) {
  const n = RING_NORMAL[which];
  if (!n) return from;
  const dot = from.x * n.x + from.y * n.y + from.z * n.z;
  const p = { x: from.x - n.x * dot, y: from.y - n.y * dot, z: from.z - n.z * dot };
  return Math.hypot(p.x, p.y, p.z) < 1e-9 ? from : normalize3(p);
}

function cross3(u, v) {
  return {
    x: u.y * v.z - u.z * v.y,
    y: u.z * v.x - u.x * v.z,
    z: u.x * v.y - u.y * v.x,
  };
}

function quadrant(from, top, which) {
  const seen = inRingPlane(from, which);
  const along = seen.x * top.x + seen.y * top.y + seen.z * top.z;

  const perp = cross3(RING_NORMAL[which] || top, top);
  const across = seen.x * perp.x + seen.y * perp.y + seen.z * perp.z;

  const fromTop = along >= 0;
  const end = fromTop ? top : { x: -top.x, y: -top.y, z: -top.z };
  const side = across >= 0 ? 1 : -1;
  return { end, sign: fromTop ? side : -side };
}

function quarterTurnToward(from, to) {
  const dot = from.x * to.x + from.y * to.y + from.z * to.z;
  const p = {
    x: to.x - from.x * dot,
    y: to.y - from.y * dot,
    z: to.z - from.z * dot,
  };
  return Math.hypot(p.x, p.y, p.z) < 1e-9 ? null : normalize3(p);
}

function dirRay(g, ball, dir, label, from) {
  const seat = seatFromDir(dir);
  const base = from || ball.origin;
  const [ox, oy] = viewProject(ball.view, ball.c, base);
  const [x, y, z] = ballPointAt(ball, seat.a, seat.b);
  g.appendChild(tag('line', { x1: ox, y1: oy, x2: x, y2: y, class: 'ring-normal' }));
  g.appendChild(tag('circle', { cx: x, cy: y, r: 5.5, class: `ring-normal-dot${backish(z)}` }));

  const [lx, ly] = ballPointAt(ball, seat.a, seat.b, ball.r * SPHERE_LABEL_OFFSET);
  const t = tag('text', { x: lx, y: ly + 4, class: `ring-label normal${backish(z)}` });
  t.textContent = label;
  g.appendChild(t);
}

function normalsShown(view) {
  return view.normals !== false;
}

function sphereAngles(g, ball, spec, incoming) {
  for (const which of ['theta', 'phi']) {
    if (!spec[which] || spec[which].axis === undefined) continue;
    const [a, b] = sphereAt(spec.points, spec[which].axis, which);
    const top = unitPoint(a, b);
    const glyph = sphereGlyph(which, ball.sub);
    const quad = quadrant(incoming, top, which);
    angleArc(g, ball, incoming, quad.end, ANGLE_ARC_REACH, glyph, quad.sign);

    if (normalsShown(ball.view)) {
      const normal = quarterTurnToward(incoming, top);
      const tilt = tiltedEnd(spec, which, incoming);
      const seat = tilt && {
        x: ball.origin.x + ball.r * tilt.x,
        y: ball.origin.y + ball.r * tilt.y,
        z: ball.origin.z + ball.r * tilt.z,
      };
      if (normal) dirRay(g, ball, normal, `normal ${glyph}`, seat);
    }
  }
}
