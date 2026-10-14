(function () {
  const R = 0.5, STEP = 2 * R;
  const cv = document.getElementById('view'), ctx = cv.getContext('2d');
  const css = name => getComputedStyle(document.documentElement).getPropertyValue(name).trim();
  const sub = (a, b) => [a[0] - b[0], a[1] - b[1], a[2] - b[2]];
  const add = (a, b) => [a[0] + b[0], a[1] + b[1], a[2] + b[2]];
  const len = v => Math.hypot(v[0], v[1], v[2]);

  function triangle() {
    const obtuse = 120 * Math.PI / 180, side = 10;
    const C = [0, 0, 0], A = [side, 0, 0], B = [side * Math.cos(obtuse), side * Math.sin(obtuse), 0];
    const corners = [A, B, C], out = [];
    const angle = (p, a, b) => { const u = sub(a, p), w = sub(b, p); return Math.acos((u[0] * w[0] + u[1] * w[1] + u[2] * w[2]) / (len(u) * len(w))); };
    const room = corners.map((p, i) => {
      const half = angle(p, corners[(i + 1) % 3], corners[(i + 2) % 3]) / 2;
      return 2 * Math.sin(half) >= 1 ? STEP : STEP / (2 * Math.sin(half));
    });
    corners.forEach((p, i) => {
      const q = corners[(i + 1) % 3], v = sub(q, p), L = len(v);
      out.push(p);
      const t0 = room[i], t1 = L - room[(i + 1) % 3], n = Math.floor((t1 - t0) / STEP);
      for (let k = 0; k <= n; k++) out.push(add(p, v.map(x => x * (t0 + (t1 - t0) * k / n) / L)));
    });
    const mid = [A, B, C].reduce(add).map(x => x / 3);
    return out.map(p => sub(p, mid));
  }
  const balls = triangle(), N = balls.length;

  const view = { yaw: 0, pitch: 0 };
  let zoom = 1, draw;
  const reach = Math.max(...balls.map(len)) + R;
  const scale = () => zoom * (cv.width / 2 - 30) / reach;

  function paint() {
    draw = atobCanvas(ctx, cv.width, cv.height, scale(), view, () => [0, 0, 0], css);
    ctx.clearRect(0, 0, cv.width, cv.height);
    const r = R * scale(), order = balls.map((p, i) => ({ i, q: draw.proj(p) })).sort((a, b) => a.q[2] - b.q[2]);
    order.forEach(({ i, q: [x, y] }) => {
      const hue = 200 + 140 * i / (N - 1);
      const g = ctx.createRadialGradient(x - r * 0.35, y - r * 0.35, r * 0.1, x, y, r);
      g.addColorStop(0, `hsl(${hue} 80% 85%)`); g.addColorStop(0.6, `hsl(${hue} 60% 55%)`); g.addColorStop(1, `hsl(${hue} 55% 28%)`);
      ctx.beginPath(); ctx.arc(x, y, r, 0, 2 * Math.PI); ctx.fillStyle = g; ctx.fill();
    });
  }

  let drag = null;
  cv.addEventListener('pointerdown', e => { drag = { x: e.clientX, y: e.clientY }; cv.setPointerCapture(e.pointerId); });
  cv.addEventListener('pointermove', e => {
    if (!drag) return;
    view.yaw += (e.clientX - drag.x) * 0.01; view.pitch += (e.clientY - drag.y) * 0.01;
    drag = { x: e.clientX, y: e.clientY };
    paint();
  });
  cv.addEventListener('pointerup', () => { drag = null; });
  const zoomBy = f => { zoom = Math.min(8, Math.max(0.2, zoom * f)); paint(); };
  cv.addEventListener('wheel', e => { if (!e.ctrlKey) return; e.preventDefault(); zoomBy(Math.exp(-e.deltaY * 0.01)); }, { passive: false });
  let pinch = 1;
  cv.addEventListener('gesturestart', e => { e.preventDefault(); pinch = 1; });
  cv.addEventListener('gesturechange', e => { e.preventDefault(); zoomBy(e.scale / pinch); pinch = e.scale; });

  document.getElementById('reset').onclick = () => { view.yaw = view.pitch = 0; zoom = 1; paint(); };
  matchMedia('(prefers-color-scheme: dark)').addEventListener('change', paint);
  paint();
})();
