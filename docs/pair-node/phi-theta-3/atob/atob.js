(function () {
  const O = [0, 0, 0];
  const PAIRS = [
    { node: 'a', id: 'ab', C: [1, 0, 0], color: '--local', cls: 'l', side: 1 },
    { node: 'c', id: 'cb', C: [0, 0, 1], color: '--local2', cls: 'l2', side: -1 },
    { node: 'd', id: 'db', C: [1, 1, 1], color: '--local3', cls: 'l3', side: 1 },
  ];
  const BY = Object.fromEntries(PAIRS.map(p => [p.node, p]));
  const rad = d => d * Math.PI / 180, deg = a => a * 180 / Math.PI;
  const add = (p, q) => p.map((v, i) => v + q[i]);
  const sub = (p, q) => p.map((v, i) => v - q[i]);
  const len = v => Math.hypot(...v);
  const toCart = (R, ph, th) => [R * Math.sin(ph) * Math.cos(th), R * Math.cos(ph), R * Math.sin(ph) * Math.sin(th)];
  const css = name => getComputedStyle(document.documentElement).getPropertyValue(name).trim();

  const view = { yaw: -0.6, pitch: 0.9 };
  const cv = document.getElementById('view');
  const ctx = cv.getContext('2d');
  const W = cv.width, H = cv.height, S = 230;
  let focus = [0.5, 0, 0.5];
  const draw = atobCanvas(ctx, W, H, S, view, () => focus, css);
  const { proj, arrow, dot } = draw;
  const scene = atobScene(draw, ctx, S, css);

  const el = id => document.getElementById(id);
  const val = id => +el(id).value;
  const toPolar = d => {
    const r = len(d);
    const phi = r ? deg(Math.acos(Math.max(-1, Math.min(1, d[1] / r)))) : 0;
    return [r, phi, (deg(Math.atan2(d[2], d[0])) + 360) % 360];
  };
  const angles = id => [val(`sphi-${id}`), val(`sth-${id}`)];
  const to = (x, y) => sub(BY[y].C, BY[x].C);
  const vec = { a: toCart(1, rad(90), rad(90)) };
  let driver = 'start', held = null, moved = null;

  function shared(dCA, ra, rc) {
    const D = len(dCA);
    if (D < 1e-9) return null;
    const u = dCA.map(v => -v / D);
    const h = (D * D + ra * ra - rc * rc) / (2 * D);
    if (ra * ra - h * h < 0) return null;
    return { u, h, rho: Math.sqrt(ra * ra - h * h), m: u.map(v => v * h) };
  }
  const centreFrom = n => n === 'a' ? held.m : add(to(n, 'a'), held.m);

  function positions() {
    const pos = n => ['x', 'y', 'z'].map(k => val(`s${n}${k}`));
    PAIRS.forEach(p => {
      p.C = pos(p.node);
      ['x', 'y', 'z'].forEach(k => { el(`v${p.node}${k}`).textContent = val(`s${p.node}${k}`).toFixed(1); });
    });
    focus = [0, 1, 2].map(i => PAIRS.reduce((s, p) => s + p.C[i], 0) / PAIRS.length);
  }

  function state() {
    const fit = (id, r, from) => {
      const [ph, th] = angles(id);
      return len(sub(toCart(r, rad(ph), rad(th)), from)) < 1e-9 * (1 + r) ? [r, ph, th] : toPolar(from);
    };
    const gap = Object.fromEntries(['c', 'd'].map(n => [n, sub(add(to(n, 'a'), vec.a), vec[n])]));
    const common = Object.values(gap).every(g => len(g) < 1e-6);
    const p = Object.fromEntries(PAIRS.map(q => [q.node, fit(q.id, len(vec[q.node]), vec[q.node])]));
    return { gap, common, p, both: held && common ? fit('both', held.rho, sub(vec.a, held.m)) : null };
  }

  function update(group) {
    const before = Object.fromEntries(PAIRS.map(p => [p.node, p.C]));
    positions();
    if (!vec.c) ['c', 'd'].forEach(n => { vec[n] = add(to(n, 'a'), vec.a); });
    const linked = el('link').checked;
    ['sphi-both', 'sth-both'].forEach(id => { el(id).disabled = !linked; });
    if (group === 'both' && !(linked && held)) { render(); return; }
    const mover = group === 'link' ? 'a' : PAIRS.find(p => p.id === group)?.node;
    if (group.startsWith('pos-')) {
      const n = group.slice(4);
      moved = { n, delta: sub(BY[n].C, before[n]), old: vec[n] };
      vec[n] = sub(vec[n], moved.delta);
    } else if (group === 'both') {
      const [Phi, Th] = angles('both'), onSphere = toCart(held.rho, rad(Phi), rad(Th));
      PAIRS.forEach(p => { vec[p.node] = add(centreFrom(p.node), onSphere); });
    } else if (mover) {
      if (group !== 'link') {
        const [phi, th] = angles(group);
        vec[mover] = toCart(len(vec[mover]) || 1, rad(phi), rad(th));
      }
      if (linked) PAIRS.forEach(p => { if (p.node !== mover) vec[p.node] = add(to(p.node, mover), vec[mover]); });
    }
    driver = group;
    if (group !== 'both') held = shared(to('c', 'a'), len(vec.a), len(vec.c));
    const s = state();
    [...PAIRS.map(p => [p.id, s.p[p.node]]), ['both', s.both]].forEach(([id, p]) => {
      if (id === group || !p) return;
      el(`sphi-${id}`).value = Math.round(p[1]); el(`sth-${id}`).value = Math.round(p[2]);
    });
    render();
  }

  function render() {
    ctx.clearRect(0, 0, W, H);
    const muted = css('--muted');
    [[[0.5, 0, 0], 'x'], [[0, 0.5, 0], 'y'], [[0, 0, 0.5], 'z']].forEach(([e, n]) => {
      arrow(O, e, muted, 2);
      const [x, y] = proj(e);
      ctx.fillStyle = muted; ctx.font = '24px -apple-system, sans-serif'; ctx.fillText(n, x + 6, y - 6);
    });

    const s = state(), linked = el('link').checked;
    const shown = PAIRS.filter(p => el(`show-${p.node}`).checked);
    const mover = driver === 'link' ? 'a' : PAIRS.find(p => p.id === driver)?.node;
    PAIRS.forEach(p => { p.via = linked && mover && p.node !== mover ? { to: mover, v: to(p.node, mover) } : null; });
    shown.forEach(q => {
      const [r, phi, th] = s.p[q.node];
      if (q.node === 'a') scene.sphere(q.C, r, phi, th, css('--phi'), css('--theta'));
      else scene.sphere(q.C, r, phi, th, css(q.color), css(q.color));
    });
    if (held && shown.includes(BY.a) && shown.includes(BY.c)) {
      scene.sharedSphere(BY.a.C, held, s.both);
      scene.sharedCircle(BY.a.C, held);
    }
    const b = n => add(BY[n].C, vec[n]);
    shown.forEach(q => scene.vectors(q, vec[q.node], b(q.node)));
    shown.forEach(q => { if (q.node !== 'a' && len(s.gap[q.node]) >= 1e-6) arrow(b('a'), b(q.node), muted, 2, true); });
    dot(O, css('--ink'), 'O');
    shown.forEach((q, i) => dot(b(q.node), css(q.color), `4 b_${q.node}`, true, -12 + 46 * i));

    [...PAIRS.map(p => [p.id, s.p[p.node]]), ['both', s.both]].forEach(([id, p]) => {
      el(`vphi-${id}`).textContent = p ? `${p[1].toFixed(0)}°` : '–';
      el(`vth-${id}`).textContent = p ? `${p[2].toFixed(0)}°` : '–';
    });
    const ctxt = { driver, mover, linked, held, both: s.both, vec, to, gap: s.gap, common: s.common, moved,
      m: held && Object.fromEntries(PAIRS.map(p => [p.node, centreFrom(p.node)])) };
    el('calc-both').innerHTML = atobBoth(ctxt);
    PAIRS.forEach(q => { el(`calc-${q.id}`).innerHTML = atobSteps(q, s.p[q.node], vec[q.node], b(q.node), ctxt); });
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
  PAIRS.forEach(p => el(`show-${p.node}`).addEventListener('change', render));
  matchMedia('(prefers-color-scheme: dark)').addEventListener('change', render);
  update('start');
})();
