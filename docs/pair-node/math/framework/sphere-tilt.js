const SPHERE_TILT_LABEL = 1.3;

function ringTop(incoming, which) {
  const n = RING_NORMAL[which];
  if (!n) return null;
  const seen = inRingPlane(incoming, which);
  const t = cross3(n, seen);
  return Math.hypot(t.x, t.y, t.z) < 1e-9 ? null : normalize3(t);
}

function tiltShown(view, which) {
  const tilts = view.tilts;
  return !tilts || tilts[which] !== false;
}

function tiltedEnd(spec, which, incoming) {
  const top = ringEnd(spec, which, 1);
  const along = top.x * incoming.x + top.y * incoming.y + top.z * incoming.z;
  const t = {
    x: top.x - incoming.x * along,
    y: top.y - incoming.y * along,
    z: top.z - incoming.z * along,
  };
  return Math.hypot(t.x, t.y, t.z) < 1e-9 ? ringTop(incoming, which) : normalize3(t);
}

function sphereTilt(g, ball, spec, which, incoming) {
  if (!tiltShown(ball.view, which)) return;
  const dir = tiltedEnd(spec, which, incoming);
  if (!dir) return;

  const [x1, y1] = endAt(ball, dir, 1);
  const [x2, y2] = endAt(ball, dir, -1);
  g.appendChild(tag('line', { x1, y1, x2, y2, class: 'ring-tilt' }));

  const glyph = sphereGlyph(which, ball.sub);
  for (const [label, sense] of [['δ', 1], ['tilt bottom', -1]]) {
    const [x, y, z] = endAt(ball, dir, sense);
    g.appendChild(tag('circle', { cx: x, cy: y, r: 4.5, class: `ring-tilt-dot${backish(z)}` }));

    const [lx, ly] = endAt(ball, dir, sense, ball.r * SPHERE_TILT_LABEL);
    const t = tag('text', { x: lx, y: ly + 4, class: `ring-tilt-label${backish(z)}` });
    t.textContent = `${label} ${glyph}`;
    g.appendChild(t);
  }
}
