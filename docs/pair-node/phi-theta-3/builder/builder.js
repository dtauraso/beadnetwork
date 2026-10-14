(function () {
  const KEY = 'pair-builder-v2';
  const preset = () => ({
    s: 1, rho: 0.5, Tphi: 12, Ttheta: 12, cr: 0.5,
    nodes: [
      { name: 'a', known: true, idx: { r: 2, phi: 3, theta: 0 } },
      { name: 'b', known: false, idx: { r: 0, phi: 0, theta: 0 } },
      { name: 'c', known: true, idx: { r: 4, phi: 1, theta: 3 } },
    ],
    rows: [
      { from: 'a', to: 'b', start: { phi: 3, theta: 3, r: 2 }, end: { phi: 3, theta: 6 } },
      { from: 'c', to: 'b', start: { phi: 5, theta: 0, r: 4 }, end: { phi: 5, theta: 0 } },
    ],
    item: 'b', on: { atob: true, pair: true }, step: 1,
  });
  let st;
  try { st = JSON.parse(localStorage.getItem(KEY)) || preset(); } catch { st = preset(); }
  const save = () => { try { localStorage.setItem(KEY, JSON.stringify(st)); } catch { } };

  const el = id => document.getElementById(id);
  const css = name => getComputedStyle(document.documentElement).getPropertyValue(name).trim();
  const h = (tag, props = {}, kids = []) => {
    const e = Object.assign(document.createElement(tag), props);
    kids.forEach(k => e.append(k));
    return e;
  };
  const num = (value, onset, step = 1) => {
    const i = h('input', { type: 'number', value, step });
    i.addEventListener('change', () => { const v = +i.value; if (Number.isFinite(v)) { onset(v); update(); } });
    return i;
  };

  const cv = el('view'), ctx = cv.getContext('2d');
  const view = { yaw: -0.6, pitch: 0.5 };
  let focus = [0, 0, 0], S = 200;
  const draw = atobCanvas(ctx, cv.width, cv.height, S, view, () => focus, css);

  function setup() {
    el('s').value = st.s; el('s').onchange = () => { st.s = Math.max(1, Math.round(+el('s').value) || 1); update(); };
    el('rho').value = st.rho; el('rho').onchange = () => { st.rho = +el('rho').value || st.rho; update(); };
    [['Tphi', 'Tphi', true], ['Ttheta', 'Ttheta', true], ['cr', 'cr', false]].forEach(([id, k, whole]) => {
      el(id).value = st[k];
      el(id).onchange = () => { const v = +el(id).value; if (v > 0) st[k] = whole ? Math.round(v) : v; update(); };
    });

    el('nodes').replaceChildren(...st.nodes.map(n => {
      const known = h('input', { type: 'checkbox', checked: n.known });
      known.onchange = () => { n.known = known.checked; setup(); update(); };
      const xyz = ['r', 'phi', 'theta'].map(k => { const f = num(n.idx[k], v => { n.idx[k] = v; }); f.disabled = !n.known; return h('td', {}, [f]); });
      const del = h('button', { textContent: 'remove' });
      del.onclick = () => {
        st.nodes = st.nodes.filter(x => x !== n);
        st.rows = st.rows.filter(r => r.from !== n.name && r.to !== n.name);
        setup(); update();
      };
      return h('tr', {}, [h('td', { textContent: n.name }), h('td', {}, [known]), ...xyz, h('td', {}, [del])]);
    }));

    el('tables').replaceChildren(...st.nodes.map(n => {
      const rows = st.rows.filter(r => r.to === n.name);
      const body = rows.map(r => {
        const del = h('button', { textContent: 'remove' });
        del.onclick = () => { st.rows = st.rows.filter(x => x !== r); setup(); update(); };
        const c = (o, k) => h('td', {}, [num(o[k], v => { o[k] = v; })]);
        return h('tr', {}, [h('td', { textContent: r.from }), c(r.start, 'phi'), c(r.start, 'theta'), c(r.start, 'r'),
          c(r.end, 'phi'), c(r.end, 'theta'), h('td', {}, [del])]);
      });
      const free = st.nodes.filter(x => x.name !== n.name && !rows.some(r => r.from === x.name));
      const pick = h('select', {}, free.map(x => h('option', { value: x.name, textContent: x.name })));
      const add = h('button', { textContent: `add row into ${n.name}`, disabled: !free.length });
      add.onclick = () => {
        st.rows.push({ from: pick.value, to: n.name, start: { phi: 0, theta: 0, r: 2 }, end: { phi: 0, theta: 0 } });
        setup(); update();
      };
      return h('div', {}, [
        h('h2', { textContent: `${n.name}'s table — one row per node with an edge into ${n.name}` }),
        rows.length
          ? h('table', {}, [h('thead', {}, [h('tr', {}, ['from', 'start φ', 'start θ', 'start r', 'end φ', 'end θ', ''].map(t => h('th', { textContent: t })))]),
            h('tbody', {}, body)])
          : h('div', { className: 'hint', textContent: 'no rows' }),
        h('div', { className: 'row' }, free.length ? [pick, add] : []),
      ]);
    }));

    el('item').replaceChildren(...st.nodes.map(n => h('option', { value: n.name, textContent: n.name, selected: n.name === st.item })));
    el('item').onchange = () => { st.item = el('item').value; update(); };

    el('methods').replaceChildren(...B_METHODS.map(m => {
      const box = h('input', { type: 'checkbox', checked: !!st.on[m.id] });
      box.onchange = () => { st.on[m.id] = box.checked; update(); };
      return h('label', { className: 'method' }, [box, h('span', { className: `m-${m.color.slice(2)}`, textContent: m.name })]);
    }));
  }

  function runs() {
    if (!st.nodes.some(n => n.name === st.item)) st.item = st.nodes[0]?.name;
    return B_METHODS.filter(m => st.on[m.id] && st.item).map(m => ({ m, steps: m.run(st, st.item) }));
  }

  function update() {
    const rs = runs();
    const max = Math.max(1, ...rs.map(r => r.steps.length));
    st.step = Math.min(Math.max(1, st.step), max);
    el('stepno').textContent = `step ${st.step} of ${max}`;
    el('prev').disabled = st.step <= 1; el('next').disabled = st.step >= max;
    el('lists').replaceChildren(...rs.map(({ m, steps }) => h('div', {}, [
      h('h2', { className: `m-${m.color.slice(2)}`, textContent: m.name }),
      h('ol', { className: 'steps' }, steps.map((s, i) => h('li', {
        className: [i + 1 > st.step ? 'later' : '', i + 1 === st.step ? 'now' : '', s.ok ? '' : 'bad'].join(' '),
      }, [h('span', { className: 't', textContent: `${i + 1}. ${s.title}` }),
        ...s.lines.map(l => h('code', { textContent: l })), h('div', { className: 'note', textContent: s.note })]))),
    ])));
    el('cells').checked = st.cells !== false;
    el('cells').onchange = () => { st.cells = el('cells').checked; update(); };    render(rs);
    save();
  }

  function render(rs) {
    const pts = [[0, 0, 0], ...st.nodes.filter(n => n.known).map(n => bPlace(st, n.idx))];
    rs.forEach(({ steps }) => steps.slice(0, st.step).forEach(s => s.draw.forEach(o => pts.push(...(o.arrow || [o.dot || o.square])))));
    const cells = st.cells !== false ? bCellPoints(st, rs) : [];
    cells.forEach(({ p }) => pts.push(...bCellCorners(st, p)));
    const reach = Math.max(0.25, ...pts.map(bLen));
    focus = [0, 0, 0];
    S = (cv.width / 2 - 60) / reach;
    Object.assign(draw, atobCanvas(ctx, cv.width, cv.height, S, view, () => focus, css));
    paint(rs, cells);
  }

  function paint(rs, cells) {
    ctx.clearRect(0, 0, cv.width, cv.height);
    bDrawAxes(draw, ctx, css, [0, 0, 0], st.cr, true);
    el('cellnote').replaceChildren(...bDrawCells(st, draw, ctx, css, cells).map(t => h('div', { textContent: t })));
    bDrawGuides(st, draw, ctx, css, bCellPoints(st, rs));
    const seen = {};
    rs.forEach(({ m, steps }) => {
      const color = css(m.color);
      steps.slice(0, st.step).forEach((s, i) => {
        const now = i + 1 === st.step;
        s.draw.forEach(o => {
          if (o.arrow) draw.arrow(o.arrow[0], o.arrow[1], color, now ? 4 : 2, !now);
          if (o.dot && o.label.startsWith('C_')) {
            const [x, y] = draw.proj(o.dot);
            ctx.lineWidth = now ? 4 : 2; ctx.strokeStyle = color;
            ctx.beginPath(); ctx.arc(x, y, 18, 0, Math.PI * 2); ctx.stroke();
          } else if (o.dot) {
            const key = o.dot.map(v => v.toFixed(3)).join();
            const k = seen[key] = (seen[key] || 0) + 1;
            draw.dot(o.dot, color, o.label, true, 4 + 30 * k);
          }
          if (o.square) {
            const [x, y] = draw.proj(o.square);
            ctx.lineWidth = 3; ctx.strokeStyle = color; ctx.strokeRect(x - 10, y - 10, 20, 20);
            const key = o.square.map(v => v.toFixed(3)).join();
            const k = seen[key] = (seen[key] || 0) + 1;
            ctx.fillStyle = color; ctx.font = '600 26px -apple-system, sans-serif'; ctx.fillText(o.label, x + 16, y + 4 + 30 * k);
          }
        });
      });
    });
    draw.dot([0, 0, 0], css('--ink'), 'O');
    st.nodes.filter(n => n.known).forEach(n => draw.dot(bPlace(st, n.idx), css('--ink'), n.name));
  }

  let drag = null;
  cv.addEventListener('pointerdown', e => { drag = { x: e.clientX, y: e.clientY }; cv.setPointerCapture(e.pointerId); });
  cv.addEventListener('pointermove', e => {
    if (!drag) return;
    view.yaw += (e.clientX - drag.x) * 0.01; view.pitch += (e.clientY - drag.y) * 0.01;
    drag = { x: e.clientX, y: e.clientY };
    render(runs());
  });
  cv.addEventListener('pointerup', () => { drag = null; });

  el('prev').onclick = () => { st.step--; update(); };
  el('next').onclick = () => { st.step++; update(); };
  el('add-node').onclick = () => {
    const name = 'abcdefghijklmnopqrstuvwxyz'.split('').find(c => !st.nodes.some(n => n.name === c));
    if (!name) return;
    st.nodes.push({ name, known: false, idx: { r: 0, phi: 0, theta: 0 } });
    setup(); update();
  };
  el('preset').onclick = () => { st = preset(); setup(); update(); };

  setup();
  update();
})();
