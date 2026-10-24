(function () {
  const preset = bPreset;
  let err = null, pickFrom = null;
  const mathOpen = new WeakSet();
  let builder = bLoad() || preset();
  const save = () => bKeep(builder);
  const open = s => { builder = s; reach = null; pickFrom = null; setup(); update(); };

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
    el('rho').value = builder.rho; el('rho').onchange = () => { builder.rho = +el('rho').value || builder.rho; update(); };
    [['Tphi', 'Tphi', true], ['Ttheta', 'Ttheta', true], ['cr', 'cr', false]].forEach(([id, k, whole]) => {
      el(id).value = builder[k];
      el(id).onchange = () => { const v = +el(id).value; if (v > 0) builder[k] = whole ? Math.round(v) : v; update(); };
    });

    const all = [...bOpenStep(builder).nodes];
    el('nodes').replaceChildren(...all.flatMap(n => {
      const xyz = ['r', 'phi', 'theta'].map(k => { const f = num(n.idx[k], v => { n.idx[k] = v; bRemeasure(builder, n); }); return h('td', {}, [f]); });
      const del = h('button', { textContent: 'remove' });
      del.onclick = () => { bDropNode(builder, n); setup(); update(); };
      const s = flag(`${n.name}:s`, num(bS(builder, n), v => check(`${n.name}:s`, bSetS(builder, n, v), v)));
      return [h('tr', {}, [h('td', { textContent: n.name }), h('td', {}, [s]), ...xyz, h('td', {}, [del])]), ...note(n.name, 6)];
    }));

    const c = (o, k, tree) => {
      const at = `${tree.ab.a.name}→${tree.ab.b.name}:start${k}`;
      return flag(at, num(o[k], v => check(at, bSetStart(builder, tree, k, v), v)));
    };
    const froms = all.filter(a => a.out.length < all.length - 1);
    if (!pickFrom || pickFrom.gone || pickFrom.out.length >= all.length - 1) pickFrom = froms[0];
    const taken = new Set(pickFrom ? pickFrom.out.map(t => t.ab.b) : []);
    const tos = pickFrom ? all.filter(b => b !== pickFrom && !taken.has(b)) : [];
    const pickA = h('select', {}, froms.map((a, i) => h('option', { value: i, textContent: a.name, selected: a === pickFrom })));
    pickA.onchange = () => { pickFrom = froms[+pickA.value]; setup(); };
    const pickB = h('select', {}, tos.map((b, i) => h('option', { value: i, textContent: b.name })));
    const add = h('button', { textContent: '+ add edge', disabled: !tos.length });
    add.onclick = () => {
      const row = { from: pickFrom, to: tos[+pickB.value], start: { phi: 0, theta: 0, r: 2 } };
      bAddTree(builder, row);
      row.start = bMeasure(builder, row) || row.start;
      setup(); update();
    };
    el('settings').replaceChildren(bSeqPicker(builder, h, open),
      bSteps(h, [{ body: [h('div', { className: 'part', textContent: 'edges' }), ...bEdgeTable(builder, h, c, mathOpen, err, () => { setup(); update(); }), ...bStepNodes(builder, h)],
        head: tos.length ? [pickA, ' → ', pickB, add] : [h('span', { className: 'hint', textContent: 'no edge left to add' })] }]));

    const shown = bShown(builder);
    el('item').replaceChildren(...shown.map((n, i) => h('option', { value: i, textContent: n.name, selected: n === builder.item })));
    el('item').onchange = () => { builder.item = shown[+el('item').value]; update(); };

    el('methods').replaceChildren(...B_METHODS.map(m => {
      const box = h('input', { type: 'checkbox', checked: !!builder.methods[m.id] });
      box.onchange = () => { builder.methods[m.id] = box.checked; update(); };
      return h('label', { className: 'method' }, [box, h('span', { className: `m-${m.color.slice(2)}`, textContent: m.name })]);
    }));
  }

  function runs() {
    const shown = bShown(builder);
    if (shown.length && !shown.includes(builder.item)) builder.item = shown[0];
    if (!builder.item || builder.item.gone) builder.item = bFirst(bOpenStep(builder).nodes);
    return B_METHODS.filter(m => builder.methods[m.id] && shown.includes(builder.item)).map(m => ({ m, steps: m.run(builder, builder.item) }));
  }

  const redraw = () => render(runs());

  function update() {
    const rs = runs();
    const max = Math.max(1, ...rs.map(r => r.steps.length));
    builder.step = Math.min(Math.max(1, builder.step), max);
    el('stepno').textContent = `step ${builder.step} of ${max}`;
    el('prev').disabled = builder.step <= 1; el('next').disabled = builder.step >= max;
    el('lists').replaceChildren(...rs.map(({ m, steps }) => h('div', {}, [
      h('h2', { className: `m-${m.color.slice(2)}`, textContent: m.name }),
      h('ol', { className: 'steps' }, steps.map((s, i) => h('li', {
        className: [i + 1 > builder.step ? 'later' : '', i + 1 === builder.step ? 'now' : '', s.ok ? '' : 'bad'].join(' '),
      }, [h('span', { className: 't', textContent: `${i + 1}. ${s.title}` }),
        ...s.lines.map(l => h('code', { textContent: l })), h('div', { className: 'note', textContent: s.note })]))),
    ])));
    render(rs);
    save();
  }

  function render(rs) {
    const pts = [[0, 0, 0], ...bShown(builder).map(n => bNodePos(builder, n))];
    rs.forEach(({ steps }) => steps.slice(0, builder.step).forEach(s => s.draw.forEach(o => pts.push(...(o.arrow || [o.dot || o.square])))));
    reach ??= Math.max(0.25, ...pts.map(bLen));
    focus = [0, 0, 0];
    S = pan.zoom * (cv.width / 2 - 60) / reach;
    Object.assign(draw, atobCanvas(ctx, cv.width, cv.height, S, view, () => focus, css));
    paint(rs);
  }

  function paint(rs) {
    ctx.clearRect(0, 0, cv.width, cv.height);
    ctx.save(); ctx.translate(pan.x, pan.y);
    bDrawAxes(draw, ctx, css, [0, 0, 0], builder.cr, true);
    el('framenote').replaceChildren(...bGlobalLines(builder).map(t => h('div', { textContent: t })));
    bDrawGuides(builder, draw, ctx, css, bCellPoints(builder, rs));
    const seen = {};
    rs.forEach(({ m, steps }) => {
      const color = css(m.color);
      steps.slice(0, builder.step).forEach((s, i) => {
        const now = i + 1 === builder.step;
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
    bShown(builder).forEach(n => draw.dot(bNodePos(builder, n), css('--ink'), n.name));
    ctx.restore();
  }

  let drag = null;
  cv.addEventListener('pointerdown', e => { drag = { x: e.clientX, y: e.clientY }; cv.setPointerCapture(e.pointerId); });
  cv.addEventListener('pointermove', e => {
    if (!drag) return;
    const dx = e.clientX - drag.x, dy = e.clientY - drag.y, k = cv.width / cv.getBoundingClientRect().width;
    if (e.shiftKey) { pan.x += dx * k; pan.y += dy * k; } else { view.yaw += dx * 0.01; view.pitch += dy * 0.01; }
    drag = { x: e.clientX, y: e.clientY };
    redraw();
  });
  cv.addEventListener('pointerup', () => { drag = null; });
  const zoomAt = (f, cx, cy) => {
    const box = cv.getBoundingClientRect(), k = cv.width / box.width;
    const mx = (cx - box.left) * k - cv.width / 2, my = (cy - box.top) * k - cv.height / 2;
    pan.x = mx - f * (mx - pan.x); pan.y = my - f * (my - pan.y); pan.zoom *= f;
    redraw();
  };
  cv.addEventListener('wheel', e => {
    e.preventDefault();
    if (e.ctrlKey) return zoomAt(Math.exp(-e.deltaY * 0.01), e.clientX, e.clientY);
    const k = cv.width / cv.getBoundingClientRect().width;
    pan.x -= e.deltaX * k; pan.y -= e.deltaY * k;
    redraw();
  }, { passive: false });
  let pinch = 1;
  cv.addEventListener('gesturestart', e => { e.preventDefault(); pinch = 1; });
  cv.addEventListener('gesturechange', e => { e.preventDefault(); zoomAt(e.scale / pinch, e.clientX, e.clientY); pinch = e.scale; });

  el('prev').onclick = () => { builder.step--; update(); };
  el('next').onclick = () => { builder.step++; update(); };
  el('add-node').onclick = () => {
    bOpenStep(builder).nodes.add(bMake({ name: bName(builder.made++), idx: { r: 0, phi: 0, theta: 0 } }));
    setup(); update();
  };
  el('preset').onclick = () => {
    const p = preset();
    builder.seqs[builder.open].steps = p.seqs[0].steps;
    open(Object.assign(p, { seqs: builder.seqs, open: builder.open }));
  };

  setup();
  update();
})();
