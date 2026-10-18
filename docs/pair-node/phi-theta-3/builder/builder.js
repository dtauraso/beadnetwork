(function () {
  const KEY = 'pair-builder-v5';
  const preset = bPreset;
  let err = null, pickFrom = null;
  let st;
  try { st = JSON.parse(localStorage.getItem(KEY)); if (st) bWire(st); else st = preset(); } catch { st = preset(); }
  const book = bSeqBook(), save = () => { bSeqKeep(book, st); try { localStorage.setItem(KEY, JSON.stringify(st, bPack)); } catch { } };
  const open = s => { st = s; reach = null; pickFrom = null; setup(); update(); };

  const el = id => document.getElementById(id);
  const css = name => getComputedStyle(document.documentElement).getPropertyValue(name).trim();
  const h = (tag, props = {}, kids = []) => {
    const e = Object.assign(document.createElement(tag), props);
    kids.forEach(k => e.append(k));
    return e;
  };
  const num = (value, onset, step = 1) => {
    const i = h('input', { type: 'number', value, step });
    i.addEventListener('change', () => { const v = +i.value; if (Number.isFinite(v)) { err = null; onset(v); update(); } });
    return i;
  };
  const check = (at, text, value) => { err = text ? { at, text, value } : null; setup(); };
  const flag = (at, input) => { if (err && err.at === at) { input.value = err.value; input.classList.add('invalid'); } return input; };
  const note = (at, cols) => (err && err.at.split(':')[0] === at ? [h('tr', {}, [h('td', { className: 'missing', colSpan: cols, textContent: err.text })])] : []);

  const cv = el('view'), ctx = cv.getContext('2d');
  const view = { yaw: -0.6, pitch: 0.5 }, pan = { x: 0, y: 0, zoom: 1 };
  let focus = [0, 0, 0], S = 200, reach = null;
  const draw = atobCanvas(ctx, cv.width, cv.height, S, view, () => focus, css);

  function setup() {
    el('rho').value = st.rho; el('rho').onchange = () => { st.rho = +el('rho').value || st.rho; update(); };
    [['Tphi', 'Tphi', true], ['Ttheta', 'Ttheta', true], ['cr', 'cr', false]].forEach(([id, k, whole]) => {
      el(id).value = st[k];
      el(id).onchange = () => { const v = +el(id).value; if (v > 0) st[k] = whole ? Math.round(v) : v; update(); };
    });

    const all = [...st.nodes];
    el('nodes').replaceChildren(...all.flatMap(n => {
      const known = h('input', { type: 'checkbox', checked: n.known });
      known.onchange = () => { n.known = known.checked; setup(); update(); };
      const xyz = ['r', 'phi', 'theta'].map(k => { const f = num(n.idx[k], v => { bLetGo(st, n); n.idx[k] = v; }); f.disabled = !n.known; return h('td', {}, [f]); });
      const del = h('button', { textContent: 'remove' });
      del.onclick = () => { bDropNode(st, n); setup(); update(); };
      const s = flag(`${n.name}:s`, num(bS(st, n), v => check(`${n.name}:s`, bSetS(st, n, v), v)));
      return [h('tr', {}, [h('td', { textContent: n.name }), h('td', {}, [s]), h('td', {}, [known]), ...xyz, h('td', {}, [del])]), ...note(n.name, 7)];
    }));

    const table = (head, rows) => (rows.length
      ? h('table', {}, [h('thead', {}, [h('tr', {}, head.map(t => h('th', { textContent: t })))]), h('tbody', {}, rows)])
      : h('div', { className: 'hint', textContent: 'no rows' }));
    const c = (o, k, tree, part = 'start') => {
      const at = `${tree.ab.a.name}→${tree.ab.b.name}:${part}${k}`;
      return flag(at, num(o[k], v => check(at, bSetStart(st, tree, k, v, part), v)));
    };
    const settings = [...st.trees].map(tree => {
      const del = h('button', { textContent: 'remove' });
      del.onclick = () => { bDropTree(st, tree); setup(); update(); };
      const ab = `${tree.ab.a.name}→${tree.ab.b.name}`;
      const on = { kind: v => { check(`${ab}:kind`, bSetKind(st, tree, v), v); update(); }, changed: () => { setup(); update(); } };
      const attrs = bRowCells(st, tree, c, h, on).map(([k, v]) => h('li', {}, [h('span', { className: 'k', textContent: k }), ...[].concat(v)]));
      const miss = err && err.at.split(':')[0] === ab ? [h('div', { className: 'missing', textContent: err.text })] : [];
      return h('li', {}, [h('div', { className: 'row' }, [h('span', { className: 't', textContent: ab }), del]), h('ul', { className: 'attrs' }, attrs), ...miss]);
    });
    const froms = all.filter(a => a.out.length < all.length - 1);
    if (!pickFrom || pickFrom.gone || pickFrom.out.length >= all.length - 1) pickFrom = froms[0];
    const taken = new Set(pickFrom ? pickFrom.out.map(t => t.ab.b) : []);
    const tos = pickFrom ? all.filter(b => b !== pickFrom && !taken.has(b)) : [];
    const pickA = h('select', {}, froms.map((a, i) => h('option', { value: i, textContent: a.name, selected: a === pickFrom })));
    pickA.onchange = () => { pickFrom = froms[+pickA.value]; setup(); };
    const pickB = h('select', {}, tos.map((b, i) => h('option', { value: i, textContent: b.name })));
    const add = h('button', { textContent: 'add a→b setting', disabled: !tos.length });
    add.onclick = () => {
      const row = { from: pickFrom, to: tos[+pickB.value], kind: 'angle', start: { phi: 0, theta: 0, r: 2 }, end: { phi: 0, theta: 0 } };
      bLink(st, bAddTree(st, row).ab);
      setup(); update();
    };
    el('settings').replaceChildren(bSeqPicker(book, h, open),
      bSteps(h, [settings.length ? h('ul', { className: 'abs' }, settings) : h('div', { className: 'hint', textContent: 'no rows' })]),
      h('div', { className: 'row' }, tos.length ? [pickA, ' → ', pickB, add] : []), ...bLands(st, h),
      h('h2', { textContent: 'a→b settings — one item per a→b, set at a, with its attributes: an angle (φ, θ, r), a change (a copy of the angle into a, plus its own end φ, end θ), a copy (a stored direction plus the change of the angle row from a since it was set — no φ, θ, r of its own), or a base (measured once from where b lands, then stored)' }));
    el('tables').replaceChildren(
      ...all.map(n => {
        const seen = [...n.out.flatMap(tree => bDirs(st, tree).map(({ t }) => [tree.ab.b.name, t])),
          ...n.in.flatMap(tree => bDirs(st, tree).map(({ t }) => [tree.ab.a.name, bBack(st, tree.ab.row, t)]))];
        return h('div', {}, [
          h('h2', { textContent: `${n.name}'s table — ${n.name}'s φ, θ, r toward each node it connects to, in ${n.name}'s ticks (s ${bS(st, n)}: ${12 * bS(st, n)} a turn)` }),
          table(['toward', 'φ', 'θ', 'r'], seen.map(([o, t]) => h('tr', {}, [o, bN(t.phi), bN(t.theta), bN(t.r)].map(x => h('td', { textContent: x }))))),
          ...bSecondBase(st, n, h, table),
        ]);
      }));

    el('item').replaceChildren(...all.map((n, i) => h('option', { value: i, textContent: n.name, selected: n === st.item })));
    el('item').onchange = () => { st.item = all[+el('item').value]; update(); };

    el('spheres').replaceChildren(...bSphereBoxes(st, h, () => { setup(); update(); }));

    el('methods').replaceChildren(...B_METHODS.map(m => {
      const box = h('input', { type: 'checkbox', checked: !!st.on[m.id] });
      box.onchange = () => { st.on[m.id] = box.checked; update(); };
      return h('label', { className: 'method' }, [box, h('span', { className: `m-${m.color.slice(2)}`, textContent: m.name })]);
    }));
  }

  function runs() {
    if (!st.item || st.item.gone) st.item = bFirst(st.nodes);
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
    const pts = [[0, 0, 0], ...[...st.nodes].filter(n => n.known).map(n => bPlace(st, n.idx))];
    rs.forEach(({ steps }) => steps.slice(0, st.step).forEach(s => s.draw.forEach(o => pts.push(...(o.arrow || [o.dot || o.square])))));
    const cells = st.cells !== false ? bLinks(st) : [];
    cells.forEach(({ P, t, s }) => pts.push(...bCellCorners(st, P, t, s)));
    reach ??= Math.max(0.25, ...pts.map(bLen));
    focus = [0, 0, 0];
    S = pan.zoom * (cv.width / 2 - 60) / reach;
    Object.assign(draw, atobCanvas(ctx, cv.width, cv.height, S, view, () => focus, css));
    paint(rs, cells);
  }

  function paint(rs, cells) {
    ctx.clearRect(0, 0, cv.width, cv.height);
    ctx.save(); ctx.translate(pan.x, pan.y);
    bDrawAxes(draw, ctx, css, [0, 0, 0], st.cr, true);
    el('cellnote').replaceChildren(...[...bDrawCells(st, draw, ctx, css, cells), ...bGlobalLines(st)].map(t => h('div', { textContent: t })));
    bDrawGuides(st, draw, ctx, css, bCellPoints(st, rs));
    bDrawSpheres(st, draw, ctx);
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
    [...st.nodes].filter(n => n.known).forEach(n => draw.dot(bPlace(st, n.idx), css('--ink'), n.name));
    ctx.restore();
  }

  let drag = null;
  cv.addEventListener('pointerdown', e => { drag = { x: e.clientX, y: e.clientY }; cv.setPointerCapture(e.pointerId); });
  cv.addEventListener('pointermove', e => {
    if (!drag) return;
    const dx = e.clientX - drag.x, dy = e.clientY - drag.y, k = cv.width / cv.getBoundingClientRect().width;
    if (e.shiftKey) { pan.x += dx * k; pan.y += dy * k; } else { view.yaw += dx * 0.01; view.pitch += dy * 0.01; }
    drag = { x: e.clientX, y: e.clientY };
    render(runs());
  });
  cv.addEventListener('pointerup', () => { drag = null; });
  const zoomAt = (f, cx, cy) => {
    const box = cv.getBoundingClientRect(), k = cv.width / box.width;
    const mx = (cx - box.left) * k - cv.width / 2, my = (cy - box.top) * k - cv.height / 2;
    pan.x = mx - f * (mx - pan.x); pan.y = my - f * (my - pan.y); pan.zoom *= f;
    render(runs());
  };
  cv.addEventListener('wheel', e => {
    e.preventDefault();
    if (e.ctrlKey) return zoomAt(Math.exp(-e.deltaY * 0.01), e.clientX, e.clientY);
    const k = cv.width / cv.getBoundingClientRect().width;
    pan.x -= e.deltaX * k; pan.y -= e.deltaY * k;
    render(runs());
  }, { passive: false });
  let pinch = 1;
  cv.addEventListener('gesturestart', e => { e.preventDefault(); pinch = 1; });
  cv.addEventListener('gesturechange', e => { e.preventDefault(); zoomAt(e.scale / pinch, e.clientX, e.clientY); pinch = e.scale; });

  el('prev').onclick = () => { st.step--; update(); };
  el('next').onclick = () => { st.step++; update(); };
  el('add-node').onclick = () => {
    st.nodes.add(bMake({ name: bName(st.made++), known: false, idx: { r: 0, phi: 0, theta: 0 } }));
    setup(); update();
  };
  el('preset').onclick = () => open(preset());

  setup();
  update();
})();
