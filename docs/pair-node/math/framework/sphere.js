const SPHERE_TILT0 = 22 * Math.PI / 180;
const SPHERE_YAW0 = 32 * Math.PI / 180;
const SPHERE_ZOOM_MIN = 0.55;
const SPHERE_ZOOM_MAX = 2.4;
const SPHERE_DRAG_RATE = 0.008;
const SPHERE_TILT_LIMIT = Math.PI / 2 - 0.05;
const SPHERE_SUBS = ['₁', '₂'];
const SPHERE_LABEL_OFFSETS = [1.15, 1.55];
const SPHERE_RADIUS_FRACTION = 0.33;

function unitPoint(aTurn, bTurn) {
  const a = aTurn * 2 * Math.PI;
  const b = bTurn * 2 * Math.PI;
  return {
    x: Math.cos(b) * Math.sin(a),
    y: Math.sin(b),
    z: Math.cos(b) * Math.cos(a),
  };
}

function viewProject(view, c, p) {
  const o = view.pivot || { x: 0, y: 0, z: 0 };
  const px = p.x - o.x, py = p.y - o.y, pz = p.z - o.z;

  const cy = Math.cos(view.yaw), sy = Math.sin(view.yaw);
  const x = px * cy + pz * sy;
  const z = pz * cy - px * sy;
  const ct = Math.cos(view.tilt), st = Math.sin(view.tilt);
  const pan = view.pan || { x: 0, y: 0 };
  return [c + pan.x + x, c + pan.y - (py * ct - z * st), py * st + z * ct];
}

function ballPointAt(ball, aTurn, bTurn, reach) {
  const u = unitPoint(aTurn, bTurn);
  const r = reach === undefined ? ball.r : reach;
  return viewProject(ball.view, ball.c, {
    x: ball.origin.x + r * u.x,
    y: ball.origin.y + r * u.y,
    z: ball.origin.z + r * u.z,
  });
}

function sphereAt(points, i, which) {
  const t = i / points;
  return which === 'theta' ? [t, 0] : [0, t];
}

function sphereGlyph(which, sub) {
  return (which === 'theta' ? 'θ' : 'φ') + sub;
}

function backish(z) {
  return z < 0 ? ' back' : '';
}

function sphereCircle(g, ball, vary) {
  const pts = [];
  for (let k = 0; k <= 120; k++) {
    const [x, y] = ballPointAt(ball, ...vary(k / 120));
    pts.push(`${k ? 'L' : 'M'}${x.toFixed(2)} ${y.toFixed(2)}`);
  }
  g.appendChild(tag('path', { d: pts.join(' '), class: 'ring-line' }));
}

function sphereShell(g, ball, spec) {
  sphereCircle(g, ball, (t) => [t, 0]);
  sphereCircle(g, ball, (t) => [0, t]);
  for (const which of ['theta', 'phi']) {
    for (let i = 0; i < spec.points; i++) {
      const [x, y, z] = ballPointAt(ball, ...sphereAt(spec.points, i, which));
      g.appendChild(tag('circle', { cx: x, cy: y, r: 3, class: `ring-dot${backish(z)}` }));
    }
  }
  const [ox, oy] = viewProject(ball.view, ball.c, ball.origin);
  g.appendChild(tag('circle', { cx: ox, cy: oy, r: 3, class: 'ring-mark' }));
}

function sphereLabel(g, ball, spec, i, which, text, cls) {
  const [, , z] = ballPointAt(ball, ...sphereAt(spec.points, i, which));
  const [lx, ly] = ballPointAt(ball, ...sphereAt(spec.points, i, which), ball.r * ball.labelOff);
  const t = tag('text', { x: lx, y: ly + 4, class: `${cls}${backish(z)}` });
  t.textContent = text;
  g.appendChild(t);
}

function sphereEnds(g, ball, spec, which) {
  const m = spec.points / 2;
  const axis = spec[which].axis;
  const [x1, y1] = ballPointAt(ball, ...sphereAt(spec.points, axis, which));
  const [x2, y2] = ballPointAt(ball, ...sphereAt(spec.points, axis + m, which));
  g.appendChild(tag('line', { x1, y1, x2, y2, class: 'ring-axis' }));

  for (const [i, label] of [[axis, 'top'], [axis + m, 'bottom']]) {
    const [x, y, z] = ballPointAt(ball, ...sphereAt(spec.points, i, which));
    g.appendChild(tag('circle', { cx: x, cy: y, r: 5.5, class: `ring-end${backish(z)}` }));
    sphereLabel(g, ball, spec, i, which, `${label} ${sphereGlyph(which, ball.sub)}`, 'ring-label');
  }
}

function normalIndex(spec, which) {
  const n = spec.points, q = n / 4, top = spec[which].axis;
  const ahead = ((spec[which].arrival + q) % n + n) % n;
  const behind = ((spec[which].arrival - q) % n + n) % n;
  return fold(ahead - top, n) <= fold(behind - top, n) ? ahead : behind;
}

function sphereRay(g, ball, spec, which, i, lineCls, dotCls, label, labelCls) {
  const [ox, oy] = viewProject(ball.view, ball.c, ball.origin);
  const [x, y, z] = ballPointAt(ball, ...sphereAt(spec.points, i, which));
  g.appendChild(tag('line', { x1: ox, y1: oy, x2: x, y2: y, class: lineCls }));
  g.appendChild(tag('circle', { cx: x, cy: y, r: 5.5, class: `${dotCls}${backish(z)}` }));
  sphereLabel(g, ball, spec, i, which, `${label} ${sphereGlyph(which, ball.sub)}`, labelCls);
}

function sphereMarks(g, ball, spec) {
  for (const which of ['theta', 'phi']) {
    if (!spec[which] || spec[which].axis === undefined) continue;
    sphereEnds(g, ball, spec, which);
    if (spec[which].arrival === undefined) continue;
    sphereRay(g, ball, spec, which, spec[which].arrival,
      'ring-arrival', 'ring-arrival-dot', 'arrival', 'ring-label arrival');
    sphereRay(g, ball, spec, which, normalIndex(spec, which),
      'ring-normal', 'ring-normal-dot', 'normal', 'ring-label normal');
  }
}

function sphereOrigins(spec, r) {
  const [a, b] = sphereAt(spec.points, spec.theta.axis, 'theta');
  const u = unitPoint(a, b);
  return [
    { x: 0, y: 0, z: 0 },
    { x: r * u.x, y: r * u.y, z: r * u.z },
  ];
}

function sphereDraw(g, spec, view, S) {
  while (g.firstChild) g.removeChild(g.firstChild);

  const c = S / 2, r = S * SPHERE_RADIUS_FRACTION * view.zoom;
  const origins = sphereOrigins(spec, r);
  view.pivot = origins[view.pivotIndex] || origins[0];
  origins.forEach((origin, k) => {
    const ball = { view, c, r, origin, sub: SPHERE_SUBS[k], labelOff: SPHERE_LABEL_OFFSETS[k] };
    sphereShell(g, ball, spec);
    sphereMarks(g, ball, spec);
  });
}
