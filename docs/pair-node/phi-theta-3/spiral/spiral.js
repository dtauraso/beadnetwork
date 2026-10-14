(function () {
  const KEY = 'pair-spiral-v2', R = 0.5, N = 30, GAP = 2 * R * 1.08;
  const cv = document.getElementById('view'), ctx = cv.getContext('2d'), status = document.getElementById('status');
  const css = name => getComputedStyle(document.documentElement).getPropertyValue(name).trim();
  const sub = (a, b) => [a[0] - b[0], a[1] - b[1], a[2] - b[2]];
  const add = (a, b) => [a[0] + b[0], a[1] + b[1], a[2] + b[2]];
  const len = v => Math.hypot(v[0], v[1], v[2]);

  function spiral() {
    const rho = 1.6, rise = 0.25, turn = 2 * Math.asin(Math.sqrt(GAP * GAP - rise * rise) / (2 * rho));
    return Array.from({ length: N }, (_, i) => [rho * Math.cos(i * turn), (i - (N - 1) / 2) * rise, rho * Math.sin(i * turn)]);
  }

  let balls;
  try { balls = JSON.parse(localStorage.getItem(KEY)); } catch { }
  if (!Array.isArray(balls) || balls.length !== N) balls = spiral();
  const save = () => { try { localStorage.setItem(KEY, JSON.stringify(balls)); } catch { } };
  let pivot = null, mover = null, blocked = null;

  const view = { yaw: -0.6, pitch: 0.75 };
  let zoom = 1, draw;
  const reach = Math.max(...spiral().map(len)) + R;
  const layout = () => { draw = atobCanvas(ctx, cv.width, cv.height, zoom * (cv.width / 2 - 30) / reach, view, () => [0, 0, 0], css); };

  function toView([x, y, z]) {
    const cy = Math.cos(view.yaw), sy = Math.sin(view.yaw), cp = Math.cos(view.pitch), sp = Math.sin(view.pitch);
    const x1 = x * cy + z * sy, z1 = -x * sy + z * cy;
    return [x1, y * cp - z1 * sp, y * sp + z1 * cp];
  }
  function fromView([a, b, c]) {
    const cy = Math.cos(view.yaw), sy = Math.sin(view.yaw), cp = Math.cos(view.pitch), sp = Math.sin(view.pitch);
    const z1 = -b * sp + c * cp;
    return [a * cy - z1 * sy, b * cp + c * sp, a * sy + z1 * cy];
  }
  function turn(v, ax, ay) {
    let [a, b, c] = toView(v);
    [a, c] = [a * Math.cos(ax) + c * Math.sin(ax), -a * Math.sin(ax) + c * Math.cos(ax)];
    [b, c] = [b * Math.cos(ay) + c * Math.sin(ay), -b * Math.sin(ay) + c * Math.cos(ay)];
    return fromView([a, b, c]);
  }
  const hits = (i, p) => balls.findIndex((q, j) => j !== i && len(sub(p, q)) < 2 * R - 1e-9);

  function rotate(i, about, ax, ay) {
    const P = balls[about], d = len(sub(balls[i], P));
    const n = Math.max(1, Math.ceil(d * Math.hypot(ax, ay) / (0.05 * R)));
    blocked = null;
    for (let k = 0; k < n; k++) {
      const next = add(P, turn(sub(balls[i], P), ax / n, ay / n));
      const j = hits(i, next);
      if (j >= 0) { blocked = j; return; }
      balls[i] = next;
    }
  }

  function paint() {
    layout();
    ctx.clearRect(0, 0, cv.width, cv.height);
    const s = zoom * (cv.width / 2 - 30) / reach, order = balls.map((p, i) => ({ i, q: draw.proj(p) })).sort((a, b) => a.q[2] - b.q[2]);
    order.forEach(({ i, q: [x, y] }) => {
      const hue = 200 + 140 * i / (N - 1), r = R * s;
      const g = ctx.createRadialGradient(x - r * 0.35, y - r * 0.35, r * 0.1, x, y, r);
      g.addColorStop(0, `hsl(${hue} 80% 85%)`); g.addColorStop(0.6, `hsl(${hue} 60% 55%)`); g.addColorStop(1, `hsl(${hue} 55% 28%)`);
      ctx.beginPath(); ctx.arc(x, y, r, 0, 2 * Math.PI); ctx.fillStyle = g; ctx.fill();
      const ring = i === pivot ? '--pivot' : i === mover ? '--mover' : i === blocked ? '--bad' : null;
      if (ring) { ctx.lineWidth = 5; ctx.strokeStyle = css(ring); ctx.stroke(); }
    });
    status.className = blocked === null ? '' : 'blocked';
    status.textContent = blocked === null ? (pivot === null ? 'no centre picked' : `centre: sphere ${pivot + 1}`) : `stopped: sphere ${blocked + 1} is in the way`;
  }

  function pick(e) {
    const box = cv.getBoundingClientRect(), k = cv.width / box.width;
    const x = (e.clientX - box.left) * k, y = (e.clientY - box.top) * k, r = R * zoom * (cv.width / 2 - 30) / reach;
    const front = balls.map((p, i) => ({ i, q: draw.proj(p) })).sort((a, b) => b.q[2] - a.q[2]);
    const hit = front.find(({ q }) => Math.hypot(q[0] - x, q[1] - y) <= r);
    return hit ? hit.i : null;
  }

  let drag = null;
  cv.addEventListener('pointerdown', e => {
    const i = pick(e);
    drag = { x: e.clientX, y: e.clientY, i, moved: false };
    cv.setPointerCapture(e.pointerId);
  });
  cv.addEventListener('pointermove', e => {
    if (!drag) return;
    const dx = e.clientX - drag.x, dy = e.clientY - drag.y;
    if (!drag.moved && Math.hypot(dx, dy) < 4) return;
    drag.moved = true;
    if (drag.i !== null && drag.i !== pivot) {
      mover = drag.i;
      rotate(mover, pivot ?? (mover === 0 ? 1 : mover - 1), dx * 0.01, -dy * 0.01);
    } else { view.yaw += dx * 0.01; view.pitch += dy * 0.01; }
    drag.x = e.clientX; drag.y = e.clientY;
    paint();
  });
  cv.addEventListener('pointerup', () => {
    if (drag && !drag.moved && drag.i !== null) { pivot = pivot === drag.i ? null : drag.i; blocked = null; }
    drag = null; mover = null;
    save(); paint();
  });
  const zoomBy = f => { zoom = Math.min(8, Math.max(0.2, zoom * f)); paint(); };
  cv.addEventListener('wheel', e => { if (!e.ctrlKey) return; e.preventDefault(); zoomBy(Math.exp(-e.deltaY * 0.01)); }, { passive: false });
  let pinch = 1;
  cv.addEventListener('gesturestart', e => { e.preventDefault(); pinch = 1; });
  cv.addEventListener('gesturechange', e => { e.preventDefault(); zoomBy(e.scale / pinch); pinch = e.scale; });

  document.getElementById('reset').onclick = () => { balls = spiral(); pivot = mover = blocked = null; save(); paint(); };
  matchMedia('(prefers-color-scheme: dark)').addEventListener('change', paint);
  paint();
})();
