(function () {
  const preset = bPreset;
  let err = null, pickFrom = null;
  let builder = BBuilder.load() || preset();
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

    const all = [...builder.openStep().nodes];
    el('nodes').replaceChildren(...all.flatMap(n => {
      const xyz = ['r', 'phi', 'theta'].map(k => { const f = num(n.index(builder)[k], v => { n.setIdx(builder, k, v); }); return h('td', {}, [f]); });
      const del = h('button', { textContent: 'remove' });
      del.onclick = () => { builder.dropNode(n); setup(); update(); };
      const s = flag(`${n.name}:s`, num(n.scale(builder), v => check(`${n.name}:s`, n.setS(builder, v), v)));
      return [h('tr', {}, [h('td', { textContent: n.name }), h('td', {}, [s]), ...xyz, h('td', {}, [del])]), ...note(n.name, 6)];
    }));

    const c = (o, k, edge) => {
      const at = `${edge.name}:start${k}`;
      const i = num(o[k], v => check(at, edge.setStart(builder, k, v), v)), turn = 12 * edge.from.scale(builder);
      if (k !== 'r') i.addEventListener('input', () => { const v = +i.value; if (i.value !== '' && (v < 0 || v >= turn)) i.value = bWrap(turn / 12, v); });
      return flag(at, i);
    };
    const edges = [...builder.openStep().edges];
    const wide = { phi: bAngleOpen.has('phi'), theta: bAngleOpen.has('theta') };
    const heads = ['edge', 'start φ', ...(wide.phi ? ['Δφ', 'φ+Δφ'] : []), 'start θ', ...(wide.theta ? ['Δθ', 'θ+Δθ'] : []), 'start r', ''];
    const settings = edges.map(edge => {
      const del = h('button', { textContent: 'remove' });
      del.onclick = () => { builder.openStep().dropEdge(edge); setup(); update(); };
      const ab = edge.name;
      const angle = (k, name) => {
        const open = wide[k];
        const b = h('button', { className: 'grow', textContent: open ? '◂' : '▸', title: `${open ? 'hide' : 'show'} Δ${name} and ${name}+Δ${name}` });
        b.onclick = () => { if (open) bAngleOpen.delete(k); else bAngleOpen.add(k); setup(); update(); };
        const more = !open ? [] : [bN(edge.delta[k]), bN(edge.dir(builder)[k])].map(t => h('td', { textContent: t }));
        return [h('td', {}, [h('span', { className: 'pair' }, [c(edge.start, k, edge), b])]), ...more];
      };
      const cells = [...angle('phi', 'φ'), ...angle('theta', 'θ'), h('td', {}, [c(edge.start, 'r', edge)])];
      const miss = err && err.at.split(':')[0] === ab ? [h('tr', {}, [h('td', { className: 'missing', colSpan: heads.length, textContent: err.text })])] : [];
      return [h('tr', {}, [h('td', { className: 't', textContent: ab }), ...cells, h('td', {}, [del])]), ...miss];
    }).flat();
    const froms = all.filter(a => a.out.length < all.length - 1);
    if (!pickFrom || pickFrom.gone || pickFrom.out.length >= all.length - 1) pickFrom = froms[0];
    const taken = new Set(pickFrom ? pickFrom.out.map(e => e.to) : []);
    const tos = pickFrom ? all.filter(b => b !== pickFrom && !taken.has(b)) : [];
    const pickA = h('select', {}, froms.map((a, i) => h('option', { value: i, textContent: a.name, selected: a === pickFrom })));
    pickA.onchange = () => { pickFrom = froms[+pickA.value]; setup(); };
    const pickB = h('select', {}, tos.map((b, i) => h('option', { value: i, textContent: b.name })));
    const add = h('button', { textContent: '+ add edge', disabled: !tos.length });
    add.onclick = () => {
      builder.openStep().connect(builder, pickFrom, tos[+pickB.value]);
      setup(); update();
    };
    const changed = () => { setup(); update(); };
    const edgeTable = !bEdgesShown ? [] : [settings.length ? h('table', { className: 'cells' }, [h('thead', {}, [h('tr', {}, heads.map(t => h('th', { textContent: t })))]), h('tbody', {}, settings)]) : h('div', { className: 'hint', textContent: 'no edges' })];
    el('settings').replaceChildren(bSeqPicker(builder, h, open),
      bSteps(h, [{ body: [...bEdgePicker(builder, h, changed), h('div', { className: 'folds' }, [bFold(h, bEdgesShown, 'edges', settings.length, () => { bEdgesShown = !bEdgesShown; }, changed)]), ...edgeTable, ...bStepNodes(builder, h, changed)],
        head: tos.length ? [pickA, ' → ', pickB, add] : [h('span', { className: 'hint', textContent: 'no edge left to add' })] }]), ...bLands(builder, h),
      h('h2', { textContent: 'a→b settings — one item per a→b, set at a, with its start (φ, θ, r)' }));

    const shown = builder.shown();
    el('item').replaceChildren(...shown.map((n, i) => h('option', { value: i, textContent: n.name, selected: n === builder.openStep().cards.item })));
    el('item').onchange = () => { builder.openStep().cards.item = shown[+el('item').value]; update(); };

    el('methods').replaceChildren(...B_METHODS.map(m => {
      const box = h('input', { type: 'checkbox', checked: !!builder.openStep().cards.methods[m.id] });
      box.onchange = () => { builder.openStep().cards.methods[m.id] = box.checked; update(); };
      return h('label', { className: 'method' }, [box, h('span', { className: `m-${m.color.slice(2)}`, textContent: m.name })]);
    }));
  }

  function runs() {
    const shown = builder.shown();
    builder.openStep().fixItem();
    return B_METHODS.filter(m => builder.openStep().cards.methods[m.id] && shown.includes(builder.openStep().cards.item)).map(m => ({ m, steps: m.run(builder, builder.openStep().cards.item) }));
  }

  const redraw = () => render(runs());

  function update() {
    const rs = runs();
    const max = Math.max(1, ...rs.map(r => r.steps.length));
    builder.openStep().cards.card = Math.min(Math.max(1, builder.openStep().cards.card), max);
    el('stepno').textContent = `step ${builder.openStep().cards.card} of ${max}`;
    el('prev').disabled = builder.openStep().cards.card <= 1; el('next').disabled = builder.openStep().cards.card >= max;
    el('lists').replaceChildren(...rs.map(({ m, steps }) => h('div', {}, [
      h('h2', { className: `m-${m.color.slice(2)}`, textContent: m.name }),
      h('ol', { className: 'steps' }, steps.map((s, i) => h('li', {
        className: [i + 1 > builder.openStep().cards.card ? 'later' : '', i + 1 === builder.openStep().cards.card ? 'now' : '', s.ok ? '' : 'bad'].join(' '),
      }, [h('span', { className: 't', textContent: `${i + 1}. ${s.title}` }),
        ...s.lines.map(l => h('code', { textContent: l })), h('div', { className: 'note', textContent: s.note })]))),
    ])));
    render(rs);
    save();
  }

  function render(rs) {
    const pts = [[0, 0, 0], ...builder.shown().map(n => n.place(builder))];
    rs.forEach(({ steps }) => steps.slice(0, builder.openStep().cards.card).forEach(s => s.draw.forEach(o => pts.push(...(o.arrow || [o.dot || o.square])))));
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
      steps.slice(0, builder.openStep().cards.card).forEach((s, i) => {
        const now = i + 1 === builder.openStep().cards.card;
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
    builder.shown().forEach(n => draw.dot(n.place(builder), css('--ink'), n.name));
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

  el('prev').onclick = () => { builder.openStep().cards.card--; update(); };
  el('next').onclick = () => { builder.openStep().cards.card++; update(); };
  el('add-node').onclick = () => {
    builder.addNode();
    setup(); update();
  };
  el('preset').onclick = () => {
    const p = preset();
    builder.openSeq().steps = p.openSeq().steps;
    open(Object.assign(p, { seqs: builder.seqs, open: builder.open }));
  };

  setup();
  update();
})();
