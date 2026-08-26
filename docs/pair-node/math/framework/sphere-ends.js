function ringTop(incoming, which) {
  const n = RING_NORMAL[which];
  if (!n) return null;
  const seen = inRingPlane(incoming, which);
  const t = cross3(n, seen);
  return Math.hypot(t.x, t.y, t.z) < 1e-9 ? null : normalize3(t);
}

function endAt(ball, dir, sense, reach) {
  const r = reach === undefined ? ball.r : reach;
  return viewProject(ball.view, ball.c, {
    x: ball.origin.x + r * sense * dir.x,
    y: ball.origin.y + r * sense * dir.y,
    z: ball.origin.z + r * sense * dir.z,
  });
}

function sphereEnds(g, ball, spec, which, incoming) {
  const top = ringTop(incoming, which);
  if (!top) return;

  const [x1, y1] = endAt(ball, top, 1);
  const [x2, y2] = endAt(ball, top, -1);
  g.appendChild(tag('line', { x1, y1, x2, y2, class: 'ring-axis' }));

  const glyph = sphereGlyph(which, ball.sub);
  for (const [label, sense] of [['top', 1], ['bottom', -1]]) {
    const [x, y, z] = endAt(ball, top, sense);
    g.appendChild(tag('circle', { cx: x, cy: y, r: 5.5, class: `ring-end${backish(z)}` }));

    const [lx, ly] = endAt(ball, top, sense, ball.r * SPHERE_LABEL_OFFSET);
    const t = tag('text', { x: lx, y: ly + 4, class: `ring-label${backish(z)}` });
    t.textContent = `${label} ${glyph}`;
    g.appendChild(t);
  }
}
