function sphereControls(root, g, spec, view, S) {
  const redraw = () => sphereDraw(g, spec, view, S);
  const clampTilt = (t) => Math.max(-SPHERE_TILT_LIMIT, Math.min(SPHERE_TILT_LIMIT, t));

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
    view.tilt = clampTilt(view.tilt + (e.clientY - lastY) * SPHERE_DRAG_RATE);
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

  root.addEventListener('keydown', (e) => {
    const step = e.shiftKey ? 0.25 : 0.08;
    const move = { ArrowLeft: [-step, 0], ArrowRight: [step, 0], ArrowUp: [0, -step], ArrowDown: [0, step] }[e.key];
    if (!move) return;
    e.preventDefault();
    view.yaw += move[0];
    view.tilt = clampTilt(view.tilt + move[1]);
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
    'two spheres, the second seated on a point of the first — drag to rotate, scroll to zoom, double-click to reset');

  const g = tag('g', {});
  root.appendChild(g);
  sphereDraw(g, spec, view, S);
  sphereControls(root, g, spec, view, S);
  return root;
}

for (const host of document.querySelectorAll('[data-sphere]')) {
  host.appendChild(sphere(JSON.parse(host.dataset.sphere)));
}
