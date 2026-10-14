(function () {
  const KEY = 'pair-triangle-v1', R = 0.5, GAP = 2 * R * 1.08;
  const cv = document.getElementById('view'), ctx = cv.getContext('2d'), status = document.getElementById('status');
  const css = name => getComputedStyle(document.documentElement).getPropertyValue(name).trim();
  const sub = (a, b) => [a[0] - b[0], a[1] - b[1], a[2] - b[2]];
  const add = (a, b) => [a[0] + b[0], a[1] + b[1], a[2] + b[2]];
  const len = v => Math.hypot(v[0], v[1], v[2]);

  function triangle() {
    const obtuse = 125 * Math.PI / 180, C = [0, 0, 0], A = [6.5, 0, 0], B = [4.5 * Math.cos(obtuse), 4.5 * Math.sin(obtuse), 0];
    const out = [];
    [[A, B], [B, C], [C, A]].forEach(([p, q]) => {
      const v = sub(q, p), n = Math.floor(len(v) / GAP);
      for (let k = 0; k < n; k++) {
        const s = add(p, v.map(x => x * k / n));
        if (k === 0 || [A, B, C, ...out].every(o => len(sub(o, s)) >= GAP - 1e-9)) out.push(s);
      }
    });
    const mid = out.reduce(add).map(x => x / out.length);
    return out.map(p => sub(p, mid));
  }
  const N = triangle().length;

  let balls;
  try { balls = JSON.parse(localStorage.getItem(KEY)); } catch { }
  if (!Array.isArray(balls) || balls.length !== N) balls = triangle();
  const save = () => { try { localStorage.setItem(KEY, JSON.stringify(balls)); } catch { } };
  let pivot = null, blocked = null;
  const chosen = new Set();

  const view = { yaw: 0, pitch: 0 };
  let zoom = 1, draw;
  const reach = Math.max(...triangle().map(len)) + R;
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
  const hits = (group, p) => balls.findIndex((q, j) => !group.includes(j) && len(sub(p, q)) < 2 * R - 1e-9);

  function rotate(group, about, ax, ay) {
    const P = balls[about], d = Math.max(...group.map(i => len(sub(balls[i], P))));
    const n = Math.max(1, Math.ceil(d * Math.hypot(ax, ay) / (0.05 * R)));
    blocked = null;
    for (let k = 0; k < n; k++) {
      const next = group.map(i => add(P, turn(sub(balls[i], P), ax / n, ay / n)));
      const j = next.map(p => hits(group, p)).find(j => j >= 0);
      if (j !== undefined) { blocked = j; return; }
      group.forEach((i, m) => { balls[i] = next[m]; });
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
      const ring = i === pivot ? '--pivot' : chosen.has(i) ? '--mover' : i === blocked ? '--bad' : null;
      if (ring) { ctx.lineWidth = 5; ctx.strokeStyle = css(ring); ctx.stroke(); }
    });
    const picked = `${chosen.size} selected; ${pivot === null ? 'no centre picked' : `centre: sphere ${pivot + 1}`}`;
    status.className = blocked === null ? '' : 'blocked';
    status.textContent = blocked === null ? picked : `stopped: sphere ${blocked + 1} is in the way`;
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
    drag = { x: e.clientX, y: e.clientY, i, second: e.button === 2 || e.ctrlKey, moved: false };
    blocked = null;
    cv.setPointerCapture(e.pointerId);
  });
  cv.addEventListener('pointermove', e => {
    if (!drag || drag.second) return;
    const dx = e.clientX - drag.x, dy = e.clientY - drag.y;
    if (!drag.moved && Math.hypot(dx, dy) < 4) return;
    drag.moved = true;
    if (pivot !== null && chosen.has(drag.i)) rotate([...chosen], pivot, dx * 0.01, -dy * 0.01);
    else { view.yaw += dx * 0.01; view.pitch += dy * 0.01; }
    drag.x = e.clientX; drag.y = e.clientY;
    paint();
  });
  cv.addEventListener('pointerup', () => {
    const i = drag && !drag.moved ? drag.i : null;
    if (i !== null && drag.second) { pivot = pivot === i ? null : i; chosen.delete(i); blocked = null; }
    else if (i !== null && i !== pivot) { chosen.has(i) ? chosen.delete(i) : chosen.add(i); blocked = null; }
    drag = null;
    save(); paint();
  });
  cv.addEventListener('contextmenu', e => e.preventDefault());
  const zoomBy = f => { zoom = Math.min(8, Math.max(0.2, zoom * f)); paint(); };
  cv.addEventListener('wheel', e => { if (!e.ctrlKey) return; e.preventDefault(); zoomBy(Math.exp(-e.deltaY * 0.01)); }, { passive: false });
  let pinch = 1;
  cv.addEventListener('gesturestart', e => { e.preventDefault(); pinch = 1; });
  cv.addEventListener('gesturechange', e => { e.preventDefault(); zoomBy(e.scale / pinch); pinch = e.scale; });

  document.getElementById('reset').onclick = () => { balls = triangle(); pivot = blocked = null; chosen.clear(); save(); paint(); };
  document.getElementById('clear').onclick = () => { pivot = blocked = null; chosen.clear(); paint(); };
  matchMedia('(prefers-color-scheme: dark)').addEventListener('change', paint);
  paint();
})();
