function ringTop(incoming, which) {
  const n = RING_NORMAL[which];
  if (!n) return null;
  const seen = inRingPlane(incoming, which);
  const t = cross3(n, seen);
  return Math.hypot(t.x, t.y, t.z) < 1e-9 ? null : normalize3(t);
}

/* Every direction perpendicular to the centre line is tangent to the other sphere,
   because the centres sit r apart and so each centre lies ON the other's surface.
   That leaves a whole circle of tangent directions, and the tilt says which one:
   0 is the direction in the ring's own plane, and turning it keeps it tangent. */
function tiltedTop(incoming, which, tilt) {
  const base = ringTop(incoming, which);
  if (!base || !tilt) return base;

  const side = normalize3(cross3(incoming, base));
  const c = Math.cos(tilt), s = Math.sin(tilt);
  return normalize3({
    x: base.x * c + side.x * s,
    y: base.y * c + side.y * s,
    z: base.z * c + side.z * s,
  });
}

function endsShown(view, which) {
  const ends = view.ends;
  return !ends || ends[which] !== false;
}

function endTilt(view, which) {
  return (view.endTilt && view.endTilt[which]) || 0;
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
  if (!endsShown(ball.view, which)) return;
  const top = tiltedTop(incoming, which, endTilt(ball.view, which));
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
