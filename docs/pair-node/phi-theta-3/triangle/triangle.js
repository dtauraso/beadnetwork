(function () {
  const R = 0.25, STEP = 1;
  const cv = document.getElementById('view'), ctx = cv.getContext('2d');
  const css = name => getComputedStyle(document.documentElement).getPropertyValue(name).trim();
  const sub = (a, b) => [a[0] - b[0], a[1] - b[1], a[2] - b[2]];
  const add = (a, b) => [a[0] + b[0], a[1] + b[1], a[2] + b[2]];
  const len = v => Math.hypot(v[0], v[1], v[2]);

  function triangle() {
    const obtuse = 120 * Math.PI / 180, side = 10;
    const C = [0, 0, 0], A = [side, 0, 0], B = [side * Math.cos(obtuse), side * Math.sin(obtuse), 0];
    const out = [];
    [[A, B], [B, C], [C, A]].forEach(([p, q]) => {
      const v = sub(q, p), n = Math.round(len(v) / STEP);
      for (let k = 0; k < n; k++) out.push(add(p, v.map(x => x * k / n)));
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
