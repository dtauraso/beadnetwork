function endsShown(view, which) {
  const ends = view.ends;
  return !ends || ends[which] !== false;
}

function endAt(ball, dir, sense, reach) {
  const r = reach === undefined ? ball.r : reach;
  return viewProject(ball.view, ball.c, {
    x: ball.origin.x + r * sense * dir.x,
    y: ball.origin.y + r * sense * dir.y,
    z: ball.origin.z + r * sense * dir.z,
  });
}

function ringEnd(spec, which, sense) {
  const m = spec.points / 2;
  const i = sense > 0 ? spec[which].axis : spec[which].axis + m;
  const [a, b] = sphereAt(spec.points, i, which);
  return unitPoint(a, b);
}

function sphereEnds(g, ball, spec, which) {
  if (!endsShown(ball.view, which)) return;

  const top = ringEnd(spec, which, 1);
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
