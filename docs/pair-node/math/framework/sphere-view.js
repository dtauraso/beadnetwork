function unitsPerPixel(root, S) {
  const box = root.getBoundingClientRect ? root.getBoundingClientRect() : null;
  return box && box.width ? S / box.width : 1;
}

function pointerUnits(root, e, S) {
  const box = root.getBoundingClientRect ? root.getBoundingClientRect() : null;
  if (!box || !box.width) return { x: 0, y: 0 };
  const scale = S / box.width;
  return { x: (e.clientX - box.left) * scale, y: (e.clientY - box.top) * scale };
}

function sphereUnderPointer(root, e, spec, view, S) {
  const { r, screen } = sphereLayout(spec, view, S);
  const p = pointerUnits(root, e, S);
  const inside = screen
    .map((s, k) => (Math.hypot(p.x - s[0], p.y - s[1]) <= r ? k : -1))
    .filter((k) => k >= 0);
  if (inside.length === 1) return { sphere: inside[0], overlap: false };
  return { sphere: -1, overlap: inside.length > 1 };
}

function sphereControls(root, g, spec, view, S) {
  const redraw = () => sphereDraw(g, spec, view, S);
  const clampTilt = (t) => Math.max(-SPHERE_TILT_LIMIT, Math.min(SPHERE_TILT_LIMIT, t));

  let dragging = false, panning = false, seating = -1, lastX = 0, lastY = 0;

  const isPan = (e) => e.shiftKey || e.button === 1 || e.button === 2;

  root.addEventListener('contextmenu', (e) => e.preventDefault());

  const grabSphere = (k) => {
    const was = sphereLayout(spec, view, S).screen;
    if (view.anchorIndex === k) {
      view.seat = antipodeSeat(view.seat || defaultSeat(spec));
      view.anchorIndex = 1 - k;
    }
    view.pivotIndex = view.anchorIndex;

    const now = sphereLayout(spec, view, S).screen;
    view.pan.x += was[view.anchorIndex][0] - now[view.anchorIndex][0];
    view.pan.y += was[view.anchorIndex][1] - now[view.anchorIndex][1];
    seating = k;
  };

  root.addEventListener('pointerdown', (e) => {
    e.preventDefault();
    panning = isPan(e);
    seating = -1;
    if (!panning) {
      const hit = sphereUnderPointer(root, e, spec, view, S);
      if (hit.overlap) return;
      if (hit.sphere >= 0) grabSphere(hit.sphere);
    }
    dragging = true;
    lastX = e.clientX;
    lastY = e.clientY;
    root.setPointerCapture(e.pointerId);
    root.classList.add(panning ? 'panning' : (seating >= 0 ? 'seating' : 'grabbing'));
    if (root.focus) root.focus({ preventScroll: true });
  });

  root.addEventListener('pointermove', (e) => {
    if (!dragging) return;
    const dx = e.clientX - lastX, dy = e.clientY - lastY;
    if (panning) {
      const perPx = unitsPerPixel(root, S);
      view.pan.x += dx * perPx;
      view.pan.y += dy * perPx;
    } else if (seating >= 0) {
      const { c, r } = sphereLayout(spec, view, S);
      const p = pointerUnits(root, e, S);
      view.seat = seatFromPointer(view, c, r, p.x, p.y);
    } else {
      view.yaw += dx * SPHERE_DRAG_RATE;
      view.tilt = clampTilt(view.tilt + dy * SPHERE_DRAG_RATE);
    }
    lastX = e.clientX;
    lastY = e.clientY;
    redraw();
  });

  const stop = (e) => {
    if (!dragging) return;
    dragging = false;
    panning = false;
    seating = -1;
    root.releasePointerCapture(e.pointerId);
    root.classList.remove('grabbing');
    root.classList.remove('panning');
    root.classList.remove('seating');
  };
  root.addEventListener('pointerup', stop);
  root.addEventListener('pointercancel', stop);

  root.addEventListener('wheel', (e) => {
    e.preventDefault();
    if (e.ctrlKey || e.metaKey) {
      view.zoom *= Math.exp(-e.deltaY * 0.0015);
      view.zoom = Math.max(SPHERE_ZOOM_MIN, Math.min(SPHERE_ZOOM_MAX, view.zoom));
    } else {
      const step = e.deltaMode === 1 ? 16 : 1;
      const perPx = unitsPerPixel(root, S);
      view.pan.x -= e.deltaX * step * perPx;
      view.pan.y -= e.deltaY * step * perPx;
    }
    redraw();
  }, { passive: false });

  root.addEventListener('selectstart', (e) => e.preventDefault());

  root.addEventListener('dblclick', (e) => {
    e.preventDefault();
    view.yaw = SPHERE_YAW0;
    view.tilt = SPHERE_TILT0;
    view.zoom = 1;
    view.pan = { x: 0, y: 0 };
    view.pivotIndex = 0;
    view.seat = defaultSeat(spec);
    view.anchorIndex = 0;
    redraw();
  });

  root.addEventListener('keydown', (e) => {
    if (e.key === '1' || e.key === '2') {
      e.preventDefault();
      view.pivotIndex = Number(e.key) - 1;
      redraw();
      return;
    }
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
  const view = {
    yaw: SPHERE_YAW0, tilt: SPHERE_TILT0, zoom: 1,
    pan: { x: 0, y: 0 }, pivotIndex: 0,
    seat: defaultSeat(spec), anchorIndex: 0,
  };

  const root = svg(S, S);
  root.setAttribute('class', 'spin');
  root.setAttribute('tabindex', '0');
  root.setAttribute('aria-label',
    'two spheres, each seated on the other’s surface — drag a sphere to walk it around the other, drag off both to rotate the pair, scroll or shift-drag to pan, pinch to zoom, 1 or 2 to rotate about that sphere, double-click to reset');

  const g = tag('g', {});
  root.appendChild(g);
  sphereDraw(g, spec, view, S);
  sphereControls(root, g, spec, view, S);
  return root;
}

for (const host of document.querySelectorAll('[data-sphere]')) {
  host.appendChild(sphere(JSON.parse(host.dataset.sphere)));
}
