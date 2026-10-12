function acuteToward(from, dir) {
  const dot = from.x * dir.x + from.y * dir.y + from.z * dir.z;
  return dot >= 0 ? dir : { x: -dir.x, y: -dir.y, z: -dir.z };
}

function tiltArc(g, ball, from, to, glyph) {
  const dot = Math.max(-1, Math.min(1, from.x * to.x + from.y * to.y + from.z * to.z));
  if (dot < 0) return;
  const ang = Math.acos(dot);
  if (ang < 1e-6) return;

  const at = (t) => {
    const seat = seatFromDir(slerp(from, to, ang, t));
    return ballPointAt(ball, seat.a, seat.b, ball.r * ANGLE_ARC_REACH);
  };

  const pts = [];
  for (let k = 0; k <= ANGLE_ARC_STEPS; k++) {
    const [x, y] = at(k / ANGLE_ARC_STEPS);
    pts.push(`${k ? 'L' : 'M'}${x.toFixed(2)} ${y.toFixed(2)}`);
  }
  g.appendChild(tag('path', { d: pts.join(' '), class: 'ring-tilt-arc' }));

  const [lx, ly, lz] = at(0.5);
  const t = tag('text', { x: lx, y: ly - 5, class: `ring-tilt-arc-label${backish(lz)}` });
  t.textContent = `${(ang / Math.PI).toFixed(2)}π ${glyph}`;
  g.appendChild(t);
}

function sphereTiltArcs(g, ball, spec, which, incoming) {
  if (!tiltShown(ball.view, which) || !endsShown(ball.view, which)) return;
  const tilt = tiltedEnd(spec, which, incoming);
  if (!tilt) return;

  const glyph = sphereGlyph(which, ball.sub);
  for (const sense of [1, -1]) {
    const end = ringEnd(spec, which, sense);
    tiltArc(g, ball, end, acuteToward(end, tilt), glyph);
  }
}
