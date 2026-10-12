(function () {
  const O = [0, 0, 0];
  const PAIRS = [
    { node: 'a', id: 'ab', C: [1, 0, 0], r: 1, color: '--local', cls: 'l', side: 1 },
    { node: 'c', id: 'cb', C: [0, 0, 1], r: 1, color: '--local2', cls: 'l2', side: -1 },
  ];
  const rad = d => d * Math.PI / 180;
  const add = (p, q) => p.map((v, i) => v + q[i]);
  const sub = (p, q) => p.map((v, i) => v - q[i]);
  const toCart = (R, ph, th) => [R * Math.sin(ph) * Math.cos(th), R * Math.cos(ph), R * Math.sin(ph) * Math.sin(th)];
  const css = name => getComputedStyle(document.documentElement).getPropertyValue(name).trim();

  const view = { yaw: -0.6, pitch: 0.9 };
  const cv = document.getElementById('view');
  const ctx = cv.getContext('2d');
  const W = cv.width, H = cv.height, S = 230, focus = [0.5, 0, 0.5];

  function proj(p) {
    const [x, y, z] = sub(p, focus);
    const cy = Math.cos(view.yaw), sy = Math.sin(view.yaw);
    const x1 = x * cy + z * sy, z1 = -x * sy + z * cy;
    const cp = Math.cos(view.pitch), sp = Math.sin(view.pitch);
    const y2 = y * cp - z1 * sp, z2 = y * sp + z1 * cp;
    return [W / 2 + x1 * S, H / 2 - y2 * S, z2];
  }

  function curve(points, about, color, width, alpha = 1) {
    const dc = proj(about)[2];
    ctx.strokeStyle = color; ctx.lineWidth = width;
    for (let i = 0; i + 1 < points.length; i++) {
      const a = proj(points[i]), b = proj(points[i + 1]);
      ctx.globalAlpha = alpha * ((a[2] + b[2]) / 2 < dc ? 0.3 : 1);
      ctx.beginPath(); ctx.moveTo(a[0], a[1]); ctx.lineTo(b[0], b[1]); ctx.stroke();
    }
    ctx.globalAlpha = 1;
  }

  function arrow(p, q, color, width, dashed) {
    const a = proj(p), b = proj(q);
    const ang = Math.atan2(b[1] - a[1], b[0] - a[0]);
    if (Math.hypot(b[0] - a[0], b[1] - a[1]) < 2) return;
    ctx.strokeStyle = color; ctx.fillStyle = color; ctx.lineWidth = width;
    ctx.setLineDash(dashed ? [10, 8] : []);
    ctx.beginPath(); ctx.moveTo(a[0], a[1]); ctx.lineTo(b[0] - 14 * Math.cos(ang), b[1] - 14 * Math.sin(ang)); ctx.stroke();
    ctx.setLineDash([]);
    ctx.beginPath(); ctx.moveTo(b[0], b[1]);
    ctx.lineTo(b[0] - 20 * Math.cos(ang - 0.4), b[1] - 20 * Math.sin(ang - 0.4));
    ctx.lineTo(b[0] - 20 * Math.cos(ang + 0.4), b[1] - 20 * Math.sin(ang + 0.4));
    ctx.closePath(); ctx.fill();
  }

  function dot(p, color, label, hollow, dy = -12) {
    const [x, y] = proj(p);
    ctx.beginPath(); ctx.arc(x, y, 9, 0, Math.PI * 2);
    ctx.fillStyle = hollow ? css('--card') : color; ctx.fill();
    ctx.lineWidth = 3; ctx.strokeStyle = color; ctx.stroke();
    if (label) { ctx.fillStyle = color; ctx.font = '600 28px -apple-system, sans-serif'; ctx.fillText(label, x + 14, y + dy); }
  }

  function label(p, text, color, dx, dy) {
    const [x, y] = proj(p);
    ctx.fillStyle = color; ctx.font = '600 26px -apple-system, sans-serif'; ctx.textAlign = dx < 0 ? 'right' : 'left'; ctx.fillText(text, x + dx, y + dy); ctx.textAlign = 'left';
  }

  function triangle(points, color) {
    const ps = points.map(proj);
    ctx.beginPath(); ctx.moveTo(ps[0][0], ps[0][1]);
    ps.slice(1).forEach(p => ctx.lineTo(p[0], p[1]));
    ctx.closePath(); ctx.fillStyle = color; ctx.globalAlpha = 0.1; ctx.fill(); ctx.globalAlpha = 1;
  }

  function sphere(pair, phi, theta) {
    const on = (ph, th) => add(pair.C, toCart(pair.r, rad(ph), rad(th)));
    const wire = css('--wire');
    for (let ph = 30; ph <= 150; ph += 30) {
      const pts = []; for (let t = 0; t <= 360; t += 6) pts.push(on(ph, t));
      curve(pts, pair.C, wire, 0.8, 0.3);
    }
    for (let th = 0; th < 180; th += 30) {
      const pts = []; for (let p = 0; p <= 360; p += 6) pts.push(on(p, th));
      curve(pts, pair.C, wire, 0.8, 0.3);
    }
    const phiPath = []; for (let p = 0; p <= 360; p += 2) phiPath.push(on(p, theta));
    curve(phiPath, pair.C, css('--phi'), 4);
    const thPath = []; for (let t = 0; t <= 360; t += 2) thPath.push(on(phi, t));
    curve(thPath, pair.C, css('--theta'), 4);
  }

  function vectors(pair, d, b) {
    const color = css(pair.color), n = pair.node;
    triangle([O, pair.C, b], color);
    arrow(O, pair.C, css('--muted'), 2, true);
    arrow(O, b, css('--global'), 3, true);
    arrow(pair.C, b, color, 5);
    label(add(pair.C, d.map(v => v / 2)), `3 d_${n}b`, color, pair.side * 12, pair.side * 26);
    label(b.map(v => v * 0.55), '5', css('--global'), -26, -8);
    const [nx, ny] = proj(pair.C);
    ctx.beginPath(); ctx.arc(nx, ny, 0.14 * S, 0, Math.PI * 2);
    ctx.fillStyle = css('--card'); ctx.globalAlpha = 0.85; ctx.fill(); ctx.globalAlpha = 1;
    ctx.lineWidth = 6; ctx.strokeStyle = color; ctx.stroke();
    label(pair.C, `2 node ${n}`, color, pair.side * (0.14 * S + 8), 0.14 * S + 10);
    dot(pair.C, color);
  }

  function render() {
    ctx.clearRect(0, 0, W, H);
    const muted = css('--muted');
    [[[0.5, 0, 0], 'x'], [[0, 0.5, 0], 'y'], [[0, 0, 0.5], 'z']].forEach(([e, n]) => {
      arrow(O, e, muted, 2);
      const [x, y] = proj(e);
      ctx.fillStyle = muted; ctx.font = '24px -apple-system, sans-serif'; ctx.fillText(n, x + 6, y - 6);
    });

    const runs = PAIRS.map(pair => {
      const phi = +document.getElementById(`sphi-${pair.id}`).value;
      const theta = +document.getElementById(`sth-${pair.id}`).value;
      const d = toCart(pair.r, rad(phi), rad(theta));
      return { pair, phi, theta, d, b: add(pair.C, d) };
    });
    runs.forEach(({ pair, phi, theta }) => sphere(pair, phi, theta));
    runs.forEach(({ pair, d, b }) => vectors(pair, d, b));
    dot(O, css('--ink'), 'O');
    runs.forEach(({ pair, b }, i) => dot(b, css(pair.color), `4 b_${pair.node}`, true, i === 0 ? -12 : 34));

    runs.forEach(({ pair, phi, theta, d, b }) => {
      document.getElementById(`vphi-${pair.id}`).textContent = `${phi}°`;
      document.getElementById(`vth-${pair.id}`).textContent = `${theta}°`;
      document.getElementById(`calc-${pair.id}`).innerHTML = atobSteps(pair, phi, theta, d, b);
    });
  }

  let drag = null;
  cv.addEventListener('pointerdown', e => { drag = { x: e.clientX, y: e.clientY }; cv.setPointerCapture(e.pointerId); });
  cv.addEventListener('pointermove', e => {
    if (!drag) return;
    view.yaw += (e.clientX - drag.x) * 0.01;
    view.pitch = Math.max(-1.5, Math.min(1.5, view.pitch + (e.clientY - drag.y) * 0.01));
    drag = { x: e.clientX, y: e.clientY };
    render();
  });
  cv.addEventListener('pointerup', () => { drag = null; });
  [['phi', 'sphi'], ['th', 'sth']].forEach(([v, s]) => {
    const both = document.getElementById(`${s}-both`);
    const show = () => { document.getElementById(`v${v}-both`).textContent = `${both.value}°`; };
    show();
    both.addEventListener('input', () => {
      PAIRS.forEach(p => { document.getElementById(`${s}-${p.id}`).value = both.value; });
      show();
    });
  });
  document.querySelectorAll('.sliders input').forEach(s => s.addEventListener('input', render));
  matchMedia('(prefers-color-scheme: dark)').addEventListener('change', render);
  render();
})();
