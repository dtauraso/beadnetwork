const SPHERE_TILT0 = 22 * Math.PI / 180;
const SPHERE_YAW0 = 32 * Math.PI / 180;
const SPHERE_ZOOM_MIN = 0.55;
const SPHERE_ZOOM_MAX = 2.4;
const SPHERE_DRAG_RATE = 0.008;
const SPHERE_TILT_LIMIT = Math.PI / 2 - 0.05;
const SPHERE_SUBS = ['₁', '₂'];
const SPHERE_RINGS = ['theta', 'phi', 'cross'];
const SPHERE_LABEL_OFFSET = 1.17;
const SPHERE_RADIUS_FRACTION = 0.33;
const SPHERE_GRIP_REACH = 1.4;

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
  if (which === 'theta') return [t, 0];
  if (which === 'phi') return [0, t];
  return [0.25, t];
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

function ringShown(view, which) {
  const rings = view.rings;
  return !rings || rings[which] !== false;
}

function sphereShell(g, ball, spec) {
  const [cx, cy] = viewProject(ball.view, ball.c, ball.origin);
  g.appendChild(tag('circle', { cx, cy, r: ball.r, class: 'ring-line' }));

  for (const which of SPHERE_RINGS) {
    if (!ringShown(ball.view, which)) continue;
    sphereCircle(g, ball, (t) => sphereAt(1, t, which));
    for (let i = 0; i < spec.points; i++) {
      const [x, y, z] = ballPointAt(ball, ...sphereAt(spec.points, i, which));
      g.appendChild(tag('circle', { cx: x, cy: y, r: 3, class: `ring-dot${backish(z)}` }));
    }
  }
  g.appendChild(tag('circle', { cx, cy, r: 3, class: 'ring-mark' }));
}

function sphereLabel(g, ball, spec, i, which, text, cls) {
  const [, , z] = ballPointAt(ball, ...sphereAt(spec.points, i, which));
  const [lx, ly] = ballPointAt(ball, ...sphereAt(spec.points, i, which), ball.r * SPHERE_LABEL_OFFSET);
  const t = tag('text', { x: lx, y: ly + 4, class: `${cls}${backish(z)}` });
  t.textContent = text;
  g.appendChild(t);
}

function sphereMarks(g, ball, spec, incoming) {
  for (const which of ['theta', 'phi']) {
    if (!spec[which] || spec[which].axis === undefined) continue;
    sphereEnds(g, ball, spec, which, incoming);
    sphereTilt(g, ball, spec, which, incoming);
    sphereTiltArcs(g, ball, spec, which, incoming);
  }
}

function defaultSeat(spec) {
  const [a, b] = sphereAt(spec.points, spec.theta.axis, 'theta');
  return { a, b };
}

function antipodeSeat(seat) {
  return { a: seat.a + 0.5, b: -seat.b };
}

function sphereOrigins(spec, r, view) {
  const seat = view.seat || defaultSeat(spec);
  const u = unitPoint(seat.a, seat.b);
  const away = { x: r * u.x, y: r * u.y, z: r * u.z };
  const home = { x: 0, y: 0, z: 0 };
  return view.anchorIndex === 1 ? [away, home] : [home, away];
}

function dirFromPointer(view, c, r, sx, sy) {
  const pan = view.pan || { x: 0, y: 0 };
  const X = sx - (c + pan.x);
  const up = (c + pan.y) - sy;

  const grip = r * SPHERE_GRIP_REACH;
  const away = (X * X + up * up) / (grip * grip);
  const depth = away <= 0.5
    ? grip * Math.sqrt(1 - away)
    : grip * 0.5 / Math.sqrt(away);

  const ct = Math.cos(view.tilt), st = Math.sin(view.tilt);
  const py = up * ct + depth * st;
  const z1 = depth * ct - up * st;

  const cy = Math.cos(view.yaw), sy2 = Math.sin(view.yaw);
  const px = X * cy - z1 * sy2;
  const pz = z1 * cy + X * sy2;

  const len = Math.hypot(px, py, pz) || 1;
  return { x: px / len, y: py / len, z: pz / len };
}

function seatFromDir(d) {
  const b = Math.asin(Math.max(-1, Math.min(1, d.y)));
  return { a: Math.atan2(d.x, d.z) / (2 * Math.PI), b: b / (2 * Math.PI) };
}

function seatDir(seat) {
  return unitPoint(seat.a, seat.b);
}

function turnedBy(from, to, v) {
  const kx = from.y * to.z - from.z * to.y;
  const ky = from.z * to.x - from.x * to.z;
  const kz = from.x * to.y - from.y * to.x;
  const s = Math.hypot(kx, ky, kz);
  const c = from.x * to.x + from.y * to.y + from.z * to.z;
  if (s < 1e-12) return c >= 0 ? v : { x: -v.x, y: -v.y, z: -v.z };

  const ux = kx / s, uy = ky / s, uz = kz / s;
  const ang = Math.atan2(s, c), cs = Math.cos(ang), sn = Math.sin(ang);
  const dotv = ux * v.x + uy * v.y + uz * v.z;
  return {
    x: v.x * cs + (uy * v.z - uz * v.y) * sn + ux * dotv * (1 - cs),
    y: v.y * cs + (uz * v.x - ux * v.z) * sn + uy * dotv * (1 - cs),
    z: v.z * cs + (ux * v.y - uy * v.x) * sn + uz * dotv * (1 - cs),
  };
}

function seatDraggedTo(grab, view, c, r, sx, sy) {
  return seatFromDir(turnedBy(grab.from, dirFromPointer(view, c, r, sx, sy), grab.seat));
}

function sphereLayout(spec, view, S) {
  const c = S / 2, r = S * SPHERE_RADIUS_FRACTION * view.zoom;
  const origins = sphereOrigins(spec, r, view);
  const pivot = origins[view.pivotIndex] || origins[0];
  const screen = origins.map((o) => viewProject({ ...view, pivot }, c, o));
  return { c, r, origins, pivot, screen };
}

function sphereDraw(g, spec, view, S) {
  while (g.firstChild) g.removeChild(g.firstChild);

  const { c, r, origins, pivot } = sphereLayout(spec, view, S);
  view.pivot = pivot;
  origins.forEach((origin, k) => {
    const ball = { view, c, r, origin, sub: SPHERE_SUBS[k] };
    const other = origins[1 - k];
    const incoming = normalize3({
      x: origin.x - other.x, y: origin.y - other.y, z: origin.z - other.z,
    });
    sphereShell(g, ball, spec);
    sphereMarks(g, ball, spec, incoming);
    sphereAngles(g, ball, spec, incoming);
  });
  sphereSpan(g, view, c, origins);
  sphereReach(g, view, c, origins);
}
