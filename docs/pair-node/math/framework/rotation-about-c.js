(function () {
  const C = [1, 0, 0], O = [0, 0, 0], r = 1;
  const rad = d => d * Math.PI / 180, deg = a => a * 180 / Math.PI;
  const add = (p, q) => p.map((v, i) => v + q[i]);
  const sub = (p, q) => p.map((v, i) => v - q[i]);
  const toCart = (R, ph, th) => [R * Math.sin(ph) * Math.cos(th), R * Math.cos(ph), R * Math.sin(ph) * Math.sin(th)];
  const on = (ph, th) => add(C, toCart(r, rad(ph), rad(th)));
  const f = n => (Math.abs(n) < 5e-4 ? 0 : n).toFixed(2);
  const pt = p => `(${f(p[0])}, ${f(p[1])}, ${f(p[2])})`;
  const css = name => getComputedStyle(document.documentElement).getPropertyValue(name).trim();

  const sPhi = document.getElementById('sphi'), sTh = document.getElementById('sth');
  const view = { yaw: -0.6, pitch: 0.35 };
  const cv = document.getElementById('view');
  const ctx = cv.getContext('2d');
  const W = cv.width, H = cv.height, S = 230, focus = [0.6, 0, 0];

  function proj(p) {
    const [x, y, z] = sub(p, focus);
    const cy = Math.cos(view.yaw), sy = Math.sin(view.yaw);
    const x1 = x * cy + z * sy, z1 = -x * sy + z * cy;
    const cp = Math.cos(view.pitch), sp = Math.sin(view.pitch);
    const y2 = y * cp - z1 * sp, z2 = y * sp + z1 * cp;
    return [W / 2 + x1 * S, H / 2 - y2 * S, z2];
  }

  function curve(points, color, width, dashed, alpha = 1) {
    const dc = proj(C)[2];
    ctx.strokeStyle = color; ctx.lineWidth = width;
    ctx.setLineDash(dashed ? [10, 8] : []);
    for (let i = 0; i + 1 < points.length; i++) {
      const a = proj(points[i]), b = proj(points[i + 1]);
      ctx.globalAlpha = alpha * ((a[2] + b[2]) / 2 < dc ? 0.3 : 1);
      ctx.beginPath(); ctx.moveTo(a[0], a[1]); ctx.lineTo(b[0], b[1]); ctx.stroke();
    }
    ctx.globalAlpha = 1; ctx.setLineDash([]);
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

  function dot(p, color, label, hollow) {
    const [x, y] = proj(p);
    ctx.beginPath(); ctx.arc(x, y, 9, 0, Math.PI * 2);
    ctx.fillStyle = hollow ? css('--card') : color; ctx.fill();
    ctx.lineWidth = 3; ctx.strokeStyle = color; ctx.stroke();
    if (label) { ctx.fillStyle = color; ctx.font = '600 28px -apple-system, sans-serif'; ctx.fillText(label, x + 14, y - 12); }
  }

  function calcSteps(phi, theta, d, Pn) {
    const g = sub(Pn, O), R = Math.hypot(...g);
    const read = R < 1e-9 ? 'P′ is at O — no angle'
      : `R = |P′ − O| = ${f(R)}, Φ = acos(P′y / R) = ${deg(Math.acos(Math.max(-1, Math.min(1, g[1] / R)))).toFixed(1)}°, ` +
        `Θ = atan2(P′z, P′x) = ${deg(Math.atan2(g[2], g[0])).toFixed(1)}°`;
    return [
      ['l', 'centre', `C = ${pt(C)}`, 'the centre of the selected node'],
      ['l', 'local angles', `r = ${f(r)}, φ = ${phi}°, θ = ${theta}°`, 'from the sliders; each changes only its own angle'],
      ['l', 'local vector', `d = polar_to_cart(r, φ, θ) = (r·sinφ·cosθ, r·cosφ, r·sinφ·sinθ) = ${pt(d)}`, 'starts at C'],
      ['l', 'new place', `P′ = C + d = ${pt(Pn)}`, 'only the end of the vector moves'],
      ['g', 'read from O', read, 'output only; never fed back into 2–4'],
    ].map(([c, name, eq, note], i) =>
      `<li><span class="${c}">${i + 1}. ${name}</span><br><code>${eq}</code><div class="note">${note}</div></li>`).join('');
  }

  function render() {
    const phi = +sPhi.value, theta = +sTh.value;
    ctx.clearRect(0, 0, W, H);
    const muted = css('--muted'), wire = css('--wire');

    [[[0.5, 0, 0], 'x'], [[0, 0.5, 0], 'y'], [[0, 0, 0.5], 'z']].forEach(([e, n]) => {
      arrow(O, e, muted, 2);
      const [x, y] = proj(e);
      ctx.fillStyle = muted; ctx.font = '24px -apple-system, sans-serif'; ctx.fillText(n, x + 6, y - 6);
    });

    for (let ph = 30; ph <= 150; ph += 30) {
      const pts = []; for (let t = 0; t <= 360; t += 6) pts.push(on(ph, t));
      curve(pts, wire, 0.8, false, 0.45);
    }
    for (let th = 0; th < 180; th += 30) {
      const pts = []; for (let p = 0; p <= 360; p += 6) pts.push(on(p, th));
      curve(pts, wire, 0.8, false, 0.45);
    }

    const phiPath = []; for (let p = 0; p <= 360; p += 2) phiPath.push(on(p, theta));
    curve(phiPath, css('--phi'), 6, false);
    const thPath = []; for (let t = 0; t <= 360; t += 2) thPath.push(on(phi, t));
    curve(thPath, css('--theta'), 6, false);

    const d = toCart(r, rad(phi), rad(theta)), Pn = add(C, d);
    arrow(O, Pn, css('--global'), 3, true);
    arrow(C, Pn, css('--local'), 5);
    const mid = proj(add(C, d.map(v => v / 2)));
    ctx.fillStyle = css('--local'); ctx.font = '600 26px -apple-system, sans-serif';
    ctx.fillText('3 d', mid[0] + 12, mid[1] + 26);
    const og = proj(add(O, Pn.map(v => v * 0.55)));
    ctx.fillStyle = css('--global'); ctx.fillText('5', og[0] - 26, og[1] - 8);
    dot(O, css('--ink'), 'O');
    const [nx, ny] = proj(C);
    ctx.beginPath(); ctx.arc(nx, ny, 0.14 * S, 0, Math.PI * 2);
    ctx.fillStyle = css('--card'); ctx.globalAlpha = 0.85; ctx.fill(); ctx.globalAlpha = 1;
    ctx.lineWidth = 6; ctx.strokeStyle = css('--local'); ctx.stroke();
    ctx.fillStyle = css('--local'); ctx.font = '600 28px -apple-system, sans-serif';
    ctx.fillText('1 node C', nx + 0.14 * S + 8, ny + 0.14 * S + 10);
    dot(C, css('--local'));
    dot(Pn, css('--local'), '4 P′', true);

    document.getElementById('vphi').textContent = `${phi}°`;
    document.getElementById('vth').textContent = `${theta}°`;
    document.getElementById('calc').innerHTML = calcSteps(phi, theta, d, Pn);
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
  sPhi.addEventListener('input', render);
  sTh.addEventListener('input', render);
  matchMedia('(prefers-color-scheme: dark)').addEventListener('change', render);
  render();
})();
