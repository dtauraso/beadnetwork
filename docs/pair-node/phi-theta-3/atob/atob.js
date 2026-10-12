(function () {
  const O = [0, 0, 0];
  const PAIRS = [
    { node: 'a', other: 'c', id: 'ab', C: [1, 0, 0], color: '--local', cls: 'l', side: 1 },
    { node: 'c', other: 'a', id: 'cb', C: [0, 0, 1], color: '--local2', cls: 'l2', side: -1 },
  ];
  const rad = d => d * Math.PI / 180, deg = a => a * 180 / Math.PI;
  const add = (p, q) => p.map((v, i) => v + q[i]);
  const sub = (p, q) => p.map((v, i) => v - q[i]);
  const toCart = (R, ph, th) => [R * Math.sin(ph) * Math.cos(th), R * Math.cos(ph), R * Math.sin(ph) * Math.sin(th)];
  const css = name => getComputedStyle(document.documentElement).getPropertyValue(name).trim();

  const view = { yaw: -0.6, pitch: 0.9 };
  const cv = document.getElementById('view');
  const ctx = cv.getContext('2d');
  const W = cv.width, H = cv.height, S = 230;
  let focus = [0.5, 0, 0.5];

  const { proj, curve, arrow, dot, label, triangle } = atobCanvas(ctx, W, H, S, view, () => focus, css);

  function grid(C, r, wire, width, alpha) {
    const on = (ph, th) => add(C, toCart(r, rad(ph), rad(th)));
    for (let ph = 30; ph <= 150; ph += 30) {
      const pts = []; for (let t = 0; t <= 360; t += 6) pts.push(on(ph, t));
      curve(pts, C, wire, width, alpha);
    }
    for (let th = 0; th < 180; th += 30) {
      const pts = []; for (let p = 0; p <= 360; p += 6) pts.push(on(p, th));
      curve(pts, C, wire, width, alpha);
    }
  }

  function sphere(pair, r, phi, theta, phiColor, thColor) {
    const on = (ph, th) => add(pair.C, toCart(r, rad(ph), rad(th)));
    grid(pair.C, r, css('--wire'), 0.8, 0.3);
    const phiPath = []; for (let p = 0; p <= 360; p += 2) phiPath.push(on(p, theta));
    curve(phiPath, pair.C, phiColor, 4);
    const thPath = []; for (let t = 0; t <= 360; t += 2) thPath.push(on(phi, t));
    curve(thPath, pair.C, thColor, 4);
  }

  function shared(dCA, ra, rc) {
    const dAC = dCA.map(v => -v), D = Math.hypot(...dAC);
    if (D < 1e-9) return null;
    const u = dAC.map(v => v / D);
    const h = (D * D + ra * ra - rc * rc) / (2 * D);
    if (ra * ra - h * h < 0) return null;
    return { u, h, rho: Math.sqrt(ra * ra - h * h), m: u.map(v => v * h), mc: u.map(v => v * (h - D)) };
  }

  function sharedCircle(pa, s) {
    const { u, h, rho } = s;
    const cross = (p, q) => [p[1] * q[2] - p[2] * q[1], p[2] * q[0] - p[0] * q[2], p[0] * q[1] - p[1] * q[0]];
    const e1 = cross(u, Math.abs(u[1]) < 0.9 ? [0, 1, 0] : [1, 0, 0]), n1 = Math.hypot(...e1);
    const f1 = e1.map(v => v / n1), f2 = cross(u, f1);
    const local = [];
    for (let t = 0; t <= 360; t += 3) {
      const c = Math.cos(rad(t)), sn = Math.sin(rad(t));
      local.push(u.map((v, i) => h * v + rho * (c * f1[i] + sn * f2[i])));
    }
    const centre = add(pa.C, s.m);
    ctx.setLineDash([12, 8]);
    curve(local.map(d => add(pa.C, d)), centre, css('--circle'), 4);
    ctx.setLineDash([]);
  }

  function sharedSphere(pa, s, Phi, Theta) {
    const centre = add(pa.C, s.m), on = (ph, th) => add(centre, toCart(s.rho, rad(ph), rad(th)));
    grid(centre, s.rho, css('--circle'), 1.2, 0.5);
    const phiPath = []; for (let p = 0; p <= 360; p += 2) phiPath.push(on(p, Theta));
    curve(phiPath, centre, css('--circle'), 2.5);
    const thPath = []; for (let t = 0; t <= 360; t += 2) thPath.push(on(Phi, t));
    curve(thPath, centre, css('--circle'), 2.5);
  }

  function vectors(pair, d, b) {
    const color = css(pair.color), n = pair.node;
    triangle([O, pair.C, b], color);
    arrow(O, pair.C, css('--muted'), 2, true);
    arrow(O, b, css('--global'), 3, true);
    if (pair.via) {
      const o = add(pair.C, pair.via);
      arrow(pair.C, o, color, 3, true);
      label(add(pair.C, pair.via.map(v => v / 2)), `2 d_${n}${pair.other}`, color, -12, -14);
      arrow(o, b, color, 2, true);
    }
    arrow(pair.C, b, color, 5);
    label(add(pair.C, d.map(v => v / 2)), `3 d_${n}b`, color, pair.side * 12, pair.side * 26);
    const [nx, ny] = proj(pair.C);
    ctx.beginPath(); ctx.arc(nx, ny, 0.14 * S, 0, Math.PI * 2);
    ctx.fillStyle = css('--card'); ctx.globalAlpha = 0.85; ctx.fill(); ctx.globalAlpha = 1;
    ctx.lineWidth = 6; ctx.strokeStyle = color; ctx.stroke();
    label(pair.C, `1 node ${n}`, color, pair.side * (0.14 * S + 8), 0.14 * S + 10);
    dot(pair.C, color);
  }

  const el = id => document.getElementById(id);
  const val = id => +el(id).value;
  const toPolar = d => {
    const r = Math.hypot(...d);
    const phi = r ? deg(Math.acos(Math.max(-1, Math.min(1, d[1] / r)))) : 0;
    return [r, phi, (deg(Math.atan2(d[2], d[0])) + 360) % 360];
  };
  const angles = id => [val(`sphi-${id}`), val(`sth-${id}`)];
  let dAB = toCart(1, rad(90), rad(90)), dCB = null, driver = 'ab', held = null;
  const len = v => Math.hypot(...v);

  function positions() {
    const [pa, pc] = PAIRS;
    const pos = id => ['x', 'y', 'z'].map(k => val(`s${id}${k}`));
    pa.C = pos('a'); pc.C = pos('c');
    pc.toA = sub(pa.C, pc.C); pa.toC = pc.toA.map(v => -v);
    focus = pa.C.map((v, i) => (v + pc.C[i]) / 2);
    ['a', 'c'].forEach(id => ['x', 'y', 'z'].forEach(k => { el(`v${id}${k}`).textContent = val(`s${id}${k}`).toFixed(1); }));
    return pc.toA;
  }

  function state() {
    const dCA = PAIRS[1].toA, gap = sub(add(dCA, dAB), dCB), common = len(gap) < 1e-6;
    return { dCA, gap, common, a: toPolar(dAB), c: toPolar(dCB), both: held && common ? toPolar(sub(dAB, held.m)) : null };
  }

  function update(group) {
    const gap = dCB && sub(add(PAIRS[1].toA, dAB), dCB);
    const dCA = positions();
    if (!dCB) dCB = add(dCA, dAB);
    if (group === 'pos-c') dCB = sub(add(dCA, dAB), gap);
    if (group === 'pos-a') dAB = add(sub(dCB, dCA), gap);
    const linked = el('link').checked;
    ['sphi-both', 'sth-both'].forEach(id => { el(id).disabled = !linked; });
    if (group === 'both' && !(linked && held)) { render(); return; }
    if (group === 'link' && linked) dCB = add(dCA, dAB);
    if (group === 'ab') {
      const [phi, th] = angles('ab');
      dAB = toCart(len(dAB) || 1, rad(phi), rad(th));
      if (linked) dCB = add(dCA, dAB);
    } else if (group === 'cb') {
      const [phi, th] = angles('cb');
      dCB = toCart(len(dCB) || 1, rad(phi), rad(th));
      if (linked) dAB = sub(dCB, dCA);
    } else if (group === 'both') {
      const [Phi, Th] = angles('both');
      const onSphere = toCart(held.rho, rad(Phi), rad(Th));
      dAB = add(held.m, onSphere);
      dCB = add(held.mc, onSphere);
    }
    driver = group;
    if (group !== 'both') held = shared(dCA, len(dAB), len(dCB));
    const s = state();
    [['ab', s.a], ['cb', s.c], ['both', s.both]].forEach(([id, p]) => {
      if (id === group || !p) return;
      el(`sphi-${id}`).value = Math.round(p[1]); el(`sth-${id}`).value = Math.round(p[2]);
    });
    render();
  }

  function render() {
    const [pa, pc] = PAIRS;
    ctx.clearRect(0, 0, W, H);
    const muted = css('--muted');
    [[[0.5, 0, 0], 'x'], [[0, 0.5, 0], 'y'], [[0, 0, 0.5], 'z']].forEach(([e, n]) => {
      arrow(O, e, muted, 2);
      const [x, y] = proj(e);
      ctx.fillStyle = muted; ctx.font = '24px -apple-system, sans-serif'; ctx.fillText(n, x + 6, y - 6);
    });

    const s = state(), fresh = shared(s.dCA, len(dAB), len(dCB));
    const runs = [
      { pair: pa, p: s.a, d: dAB, b: add(pa.C, dAB) },
      { pair: pc, p: s.c, d: dCB, b: add(pc.C, dCB) },
    ];
    const linked = el('link').checked;
    pa.via = linked && driver === 'cb' ? pa.toC : null;
    pc.via = linked && (driver === 'ab' || driver === 'link') ? pc.toA : null;
    sphere(pa, s.a[0], s.a[1], s.a[2], css('--phi'), css('--theta'));
    sphere(pc, s.c[0], s.c[1], s.c[2], css('--local2'), css('--local2'));
    if (held && s.both) sharedSphere(pa, held, s.both[1], s.both[2]);
    else if (fresh) grid(add(pa.C, fresh.m), fresh.rho, css('--circle'), 1.2, 0.5);
    if (fresh) sharedCircle(pa, fresh);
    runs.forEach(({ pair, d, b }) => vectors(pair, d, b));
    if (!s.common) arrow(runs[0].b, runs[1].b, css('--muted'), 2, true);
    dot(O, css('--ink'), 'O');
    runs.forEach(({ pair, b }, i) => dot(b, css(pair.color), `4 b_${pair.node}`, true, i === 0 ? -12 : 34));

    [['ab', s.a], ['cb', s.c], ['both', s.both]].forEach(([id, p]) => {
      el(`vphi-${id}`).textContent = p ? `${p[1].toFixed(0)}°` : '–';
      el(`vth-${id}`).textContent = p ? `${p[2].toFixed(0)}°` : '–';
    });
    const ctxt = { driver, linked: el('link').checked, held, both: s.both, dAB, dCB, toA: pc.toA, gap: s.gap, common: s.common, ra: len(dAB), rc: len(dCB) };
    el('calc-both').innerHTML = atobBoth(ctxt);
    runs.forEach(({ pair, p, d, b }) => { el(`calc-${pair.id}`).innerHTML = atobSteps(pair, p, d, b, ctxt); });
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
  document.querySelectorAll('.sliders input').forEach(s => s.addEventListener('input', () => update(s.dataset.group)));
  el('link').addEventListener('change', () => update('link'));
  matchMedia('(prefers-color-scheme: dark)').addEventListener('change', render);
  update('ab');
})();
