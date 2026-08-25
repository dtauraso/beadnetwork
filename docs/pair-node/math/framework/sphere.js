const SPHERE_TILT = 22 * Math.PI / 180;
const SPHERE_YAW = 32 * Math.PI / 180;

function onSphere(c, r, aTurn, bTurn) {
  const a = aTurn * 2 * Math.PI;
  const b = bTurn * 2 * Math.PI;
  const x0 = r * Math.cos(b) * Math.sin(a);
  const y = r * Math.sin(b);
  const z0 = r * Math.cos(b) * Math.cos(a);

  const cy = Math.cos(SPHERE_YAW), sy = Math.sin(SPHERE_YAW);
  const x = x0 * cy + z0 * sy;
  const z = z0 * cy - x0 * sy;

  const ct = Math.cos(SPHERE_TILT), st = Math.sin(SPHERE_TILT);
  return [c + x, c - (y * ct - z * st), y * st + z * ct];
}

function greatCircle(c, r, vary) {
  const pts = [];
  for (let k = 0; k <= 120; k++) {
    const [x, y] = onSphere(c, r, ...vary(k / 120));
    pts.push(`${k ? 'L' : 'M'}${x.toFixed(2)} ${y.toFixed(2)}`);
  }
  return tag('path', { d: pts.join(' '), class: 'ring-line' });
}

function sphereAt(spec, i, which) {
  const t = i / spec.points;
  return which === 'theta' ? [t, 0] : [0, t];
}

function sphereGlyph(which) {
  return which === 'theta' ? 'θ' : 'φ';
}

function sphereEnds(g, spec, which, c, r, m, labOff) {
  const axis = spec[which].axis;
  const [x1, y1] = onSphere(c, r, ...sphereAt(spec, axis, which));
  const [x2, y2] = onSphere(c, r, ...sphereAt(spec, axis + m, which));
  g.appendChild(tag('line', { x1, y1, x2, y2, class: 'ring-axis' }));

  for (const [i, label] of [[axis, 'top'], [axis + m, 'bottom']]) {
    const [x, y] = onSphere(c, r, ...sphereAt(spec, i, which));
    g.appendChild(tag('circle', { cx: x, cy: y, r: 5.5, class: 'ring-end' }));
    const [lx, ly] = onSphere(c, r * labOff, ...sphereAt(spec, i, which));
    const t = tag('text', { x: lx, y: ly + 4, class: 'ring-label' });
    t.textContent = `${label} ${sphereGlyph(which)}`;
    g.appendChild(t);
  }
}

function sphereArrival(g, spec, which, c, r, labOff) {
  const arrival = spec[which].arrival;
  if (arrival === undefined) return;
  const [ax, ay] = onSphere(c, r, ...sphereAt(spec, arrival, which));
  g.appendChild(tag('line', { x1: c, y1: c, x2: ax, y2: ay, class: 'ring-arrival' }));
  g.appendChild(tag('circle', { cx: ax, cy: ay, r: 5.5, class: 'ring-arrival-dot' }));
  const [lx, ly] = onSphere(c, r * labOff, ...sphereAt(spec, arrival, which));
  const t = tag('text', { x: lx, y: ly + 4, class: 'ring-label arrival' });
  t.textContent = `arrival ${sphereGlyph(which)}`;
  g.appendChild(t);
}

function sphere(spec) {
  const n = spec.points, m = n / 2, S = spec.size || 300, c = S / 2, r = S * 0.29;
  const labOff = 1.16;
  const g = svg(S, S);

  g.appendChild(tag('circle', { cx: c, cy: c, r: r, class: 'ring-line' }));
  g.appendChild(greatCircle(c, r, (t) => [t, 0]));
  g.appendChild(greatCircle(c, r, (t) => [0, t]));

  for (const which of ['theta', 'phi']) {
    for (let i = 0; i < n; i++) {
      const [x, y] = onSphere(c, r, ...sphereAt(spec, i, which));
      g.appendChild(tag('circle', { cx: x, cy: y, r: 3, class: 'ring-dot' }));
    }
  }

  g.appendChild(tag('circle', { cx: c, cy: c, r: 3, class: 'ring-mark' }));

  for (const which of ['theta', 'phi']) {
    if (!spec[which] || spec[which].axis === undefined) continue;
    sphereEnds(g, spec, which, c, r, m, labOff);
    sphereArrival(g, spec, which, c, r, labOff);
  }
  return g;
}

for (const host of document.querySelectorAll('[data-sphere]')) {
  host.appendChild(sphere(JSON.parse(host.dataset.sphere)));
}
