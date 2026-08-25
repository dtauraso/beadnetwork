const SPHERE_TILT0 = 22 * Math.PI / 180;
const SPHERE_YAW0 = 32 * Math.PI / 180;
const SPHERE_ZOOM_MIN = 0.55;
const SPHERE_ZOOM_MAX = 2.4;
const SPHERE_DRAG_RATE = 0.008;
const SPHERE_TILT_LIMIT = Math.PI / 2 - 0.05;

function sphereProject(view, c, r, aTurn, bTurn) {
  const a = aTurn * 2 * Math.PI;
  const b = bTurn * 2 * Math.PI;
  const x0 = r * Math.cos(b) * Math.sin(a);
  const y = r * Math.sin(b);
  const z0 = r * Math.cos(b) * Math.cos(a);

  const cy = Math.cos(view.yaw), sy = Math.sin(view.yaw);
  const x = x0 * cy + z0 * sy;
  const z = z0 * cy - x0 * sy;

  const ct = Math.cos(view.tilt), st = Math.sin(view.tilt);
  return [c + x, c - (y * ct - z * st), y * st + z * ct];
}

function sphereAt(points, i, which) {
  const t = i / points;
  return which === 'theta' ? [t, 0] : [0, t];
}

function sphereGlyph(which) {
  return which === 'theta' ? 'θ' : 'φ';
}

function backish(z) {
  return z < 0 ? ' back' : '';
}

function sphereCircle(g, view, c, r, vary) {
  const pts = [];
  for (let k = 0; k <= 120; k++) {
    const [x, y] = sphereProject(view, c, r, ...vary(k / 120));
    pts.push(`${k ? 'L' : 'M'}${x.toFixed(2)} ${y.toFixed(2)}`);
  }
  g.appendChild(tag('path', { d: pts.join(' '), class: 'ring-line' }));
}

function sphereDots(g, view, spec, c, r) {
  for (const which of ['theta', 'phi']) {
    for (let i = 0; i < spec.points; i++) {
      const [x, y, z] = sphereProject(view, c, r, ...sphereAt(spec.points, i, which));
      g.appendChild(tag('circle', { cx: x, cy: y, r: 3, class: `ring-dot${backish(z)}` }));
    }
  }
}

function sphereLabel(g, view, spec, c, r, i, which, text, cls) {
  const [, , z] = sphereProject(view, c, r, ...sphereAt(spec.points, i, which));
  const [lx, ly] = sphereProject(view, c, r * 1.17, ...sphereAt(spec.points, i, which));
  const t = tag('text', { x: lx, y: ly + 4, class: `${cls}${backish(z)}` });
  t.textContent = text;
  g.appendChild(t);
}

function sphereEnds(g, view, spec, which, c, r) {
  const m = spec.points / 2;
  const axis = spec[which].axis;
  const [x1, y1] = sphereProject(view, c, r, ...sphereAt(spec.points, axis, which));
  const [x2, y2] = sphereProject(view, c, r, ...sphereAt(spec.points, axis + m, which));
  g.appendChild(tag('line', { x1, y1, x2, y2, class: 'ring-axis' }));

  for (const [i, label] of [[axis, 'top'], [axis + m, 'bottom']]) {
    const [x, y, z] = sphereProject(view, c, r, ...sphereAt(spec.points, i, which));
    g.appendChild(tag('circle', { cx: x, cy: y, r: 5.5, class: `ring-end${backish(z)}` }));
    sphereLabel(g, view, spec, c, r, i, which, `${label} ${sphereGlyph(which)}`, 'ring-label');
  }
}

function sphereArrival(g, view, spec, which, c, r) {
  const arrival = spec[which].arrival;
  if (arrival === undefined) return;
  const [ax, ay, az] = sphereProject(view, c, r, ...sphereAt(spec.points, arrival, which));
  g.appendChild(tag('line', { x1: c, y1: c, x2: ax, y2: ay, class: 'ring-arrival' }));
  g.appendChild(tag('circle', { cx: ax, cy: ay, r: 5.5, class: `ring-arrival-dot${backish(az)}` }));
  sphereLabel(g, view, spec, c, r, arrival, which, `arrival ${sphereGlyph(which)}`, 'ring-label arrival');
}

function normalIndex(spec, which) {
  const n = spec.points, q = n / 4, top = spec[which].axis;
  const ahead = ((spec[which].arrival + q) % n + n) % n;
  const behind = ((spec[which].arrival - q) % n + n) % n;
  return fold(ahead - top, n) <= fold(behind - top, n) ? ahead : behind;
}

function sphereNormal(g, view, spec, which, c, r) {
  const arrival = spec[which].arrival;
  if (arrival === undefined) return;
  const i = normalIndex(spec, which);
  const [nx, ny, nz] = sphereProject(view, c, r, ...sphereAt(spec.points, i, which));
  g.appendChild(tag('line', { x1: c, y1: c, x2: nx, y2: ny, class: 'ring-normal' }));
  g.appendChild(tag('circle', { cx: nx, cy: ny, r: 5.5, class: `ring-normal-dot${backish(nz)}` }));
  sphereLabel(g, view, spec, c, r, i, which, `normal ${sphereGlyph(which)}`, 'ring-label normal');
}

function sphereDraw(g, spec, view, S) {
  while (g.firstChild) g.removeChild(g.firstChild);

  const c = S / 2, r = S * 0.33 * view.zoom;
  g.appendChild(tag('circle', { cx: c, cy: c, r: r, class: 'ring-line' }));
  sphereCircle(g, view, c, r, (t) => [t, 0]);
  sphereCircle(g, view, c, r, (t) => [0, t]);
  sphereDots(g, view, spec, c, r);
  g.appendChild(tag('circle', { cx: c, cy: c, r: 3, class: 'ring-mark' }));

  for (const which of ['theta', 'phi']) {
    if (!spec[which] || spec[which].axis === undefined) continue;
    sphereEnds(g, view, spec, which, c, r);
    sphereArrival(g, view, spec, which, c, r);
    sphereNormal(g, view, spec, which, c, r);
  }
}

function sphereControls(host, root, g, spec, view, S) {
  const redraw = () => sphereDraw(g, spec, view, S);

  let dragging = false, lastX = 0, lastY = 0;

  root.addEventListener('pointerdown', (e) => {
    e.preventDefault();
    dragging = true;
    lastX = e.clientX;
    lastY = e.clientY;
    root.setPointerCapture(e.pointerId);
    root.classList.add('grabbing');
    if (root.focus) root.focus({ preventScroll: true });
  });

  root.addEventListener('pointermove', (e) => {
    if (!dragging) return;
    view.yaw += (e.clientX - lastX) * SPHERE_DRAG_RATE;
    view.tilt += (e.clientY - lastY) * SPHERE_DRAG_RATE;
    view.tilt = Math.max(-SPHERE_TILT_LIMIT, Math.min(SPHERE_TILT_LIMIT, view.tilt));
    lastX = e.clientX;
    lastY = e.clientY;
    redraw();
  });

  const stop = (e) => {
    if (!dragging) return;
    dragging = false;
    root.releasePointerCapture(e.pointerId);
    root.classList.remove('grabbing');
  };
  root.addEventListener('pointerup', stop);
  root.addEventListener('pointercancel', stop);

  root.addEventListener('wheel', (e) => {
    e.preventDefault();
    view.zoom *= Math.exp(-e.deltaY * 0.0015);
    view.zoom = Math.max(SPHERE_ZOOM_MIN, Math.min(SPHERE_ZOOM_MAX, view.zoom));
    redraw();
  }, { passive: false });

  root.addEventListener('selectstart', (e) => e.preventDefault());

  root.addEventListener('dblclick', (e) => {
    e.preventDefault();
    view.yaw = SPHERE_YAW0;
    view.tilt = SPHERE_TILT0;
    view.zoom = 1;
    redraw();
  });

  host.addEventListener('keydown', (e) => {
    const step = e.shiftKey ? 0.25 : 0.08;
    const move = { ArrowLeft: [-step, 0], ArrowRight: [step, 0], ArrowUp: [0, -step], ArrowDown: [0, step] }[e.key];
    if (!move) return;
    e.preventDefault();
    view.yaw += move[0];
    view.tilt = Math.max(-SPHERE_TILT_LIMIT, Math.min(SPHERE_TILT_LIMIT, view.tilt + move[1]));
    redraw();
  });
}

function sphere(spec) {
  const S = spec.size || 300;
  const view = { yaw: SPHERE_YAW0, tilt: SPHERE_TILT0, zoom: 1 };

  const root = svg(S, S);
  root.setAttribute('class', 'spin');
  root.setAttribute('tabindex', '0');
  root.setAttribute('aria-label',
    'the phi and theta rings on one sphere — drag to rotate, scroll to zoom, double-click to reset');

  const g = tag('g', {});
  root.appendChild(g);
  sphereDraw(g, spec, view, S);
  sphereControls(root, root, g, spec, view, S);
  return root;
}

for (const host of document.querySelectorAll('[data-sphere]')) {
  host.appendChild(sphere(JSON.parse(host.dataset.sphere)));
}
