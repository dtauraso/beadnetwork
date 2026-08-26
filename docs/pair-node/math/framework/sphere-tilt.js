const SPHERE_TILT_LABEL = 1.3;

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

function endTilt(view, which) {
  return (view.endTilt && view.endTilt[which]) || 0;
}

function tiltShown(view, which) {
  const tilts = view.tilts;
  return !tilts || tilts[which] !== false;
}

function sphereTilt(g, ball, spec, which, incoming) {
  if (!tiltShown(ball.view, which)) return;
  const dir = tiltedTop(incoming, which, endTilt(ball.view, which));
  if (!dir) return;

  const [x1, y1] = endAt(ball, dir, 1);
  const [x2, y2] = endAt(ball, dir, -1);
  g.appendChild(tag('line', { x1, y1, x2, y2, class: 'ring-tilt' }));

  const glyph = sphereGlyph(which, ball.sub);
  for (const [label, sense] of [['tilt top', 1], ['tilt bottom', -1]]) {
    const [x, y, z] = endAt(ball, dir, sense);
    g.appendChild(tag('circle', { cx: x, cy: y, r: 4.5, class: `ring-tilt-dot${backish(z)}` }));

    const [lx, ly] = endAt(ball, dir, sense, ball.r * SPHERE_TILT_LABEL);
    const t = tag('text', { x: lx, y: ly + 4, class: `ring-tilt-label${backish(z)}` });
    t.textContent = `${label} ${glyph}`;
    g.appendChild(t);
  }
}
