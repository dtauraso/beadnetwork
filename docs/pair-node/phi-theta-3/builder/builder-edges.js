function bEdgePicker(builder, h, picked, changed) {
  const trees = [...builder.trees], abName = t => `${t.ab.a.name}→${t.ab.b.name}`;
  picked.splice(0, picked.length, ...picked.filter(t => trees.includes(t)));
  if (!trees.length) return [];
  const left = trees.filter(t => !picked.includes(t));
  const pick = h('select', {}, [h('option', { value: -1, textContent: left.length ? 'pick an edge' : 'all edges picked' }),
    ...left.map((t, i) => h('option', { value: i, textContent: abName(t) }))]);
  pick.onchange = () => { if (left[+pick.value]) picked.push(left[+pick.value]); changed(); };
  const by = (t, k, sign) => { t.ab.row.delta[k] += sign * t.ab.row.by[k]; bMove(builder, t); };
  const items = picked.map(t => {
    const drop = h('button', { textContent: 'unpick' });
    drop.onclick = () => { picked.splice(picked.indexOf(t), 1); changed(); };
    const own = (k, name) => {
      const amount = h('input', { type: 'number', value: t.ab.row.by[k], step: 1 });
      amount.onchange = () => { const v = +amount.value; if (bWhole(v)) t.ab.row.by[k] = Math.round(v); changed(); };
      const step = (sign, text) => { const b = h('button', { textContent: text }); b.onclick = () => { t.ab.row.by[k] += sign; changed(); }; return b; };
      return [` ${name} by`, amount, step(-1, '−'), step(1, '+')];
    };
    return h('li', {}, [h('div', { className: 'row' }, [h('span', { className: 't', textContent: abName(t) }), ...own('phi', 'Δφ'), ...own('theta', 'Δθ'), drop])]);
  });
  const redo = left.map(t => h('li', {}, [h('div', { className: 'row' }, [h('span', { className: 't', textContent: abName(t) })])]));
  const step = (k, sign, text) => {
    const b = h('button', { textContent: text, disabled: !picked.length });
    b.onclick = () => { picked.forEach(t => by(t, k, sign)); changed(); };
    return b;
  };
  return [h('div', { className: 'row' }, [h('label', {}, ['edge ', pick])]),
    items.length ? h('ul', { className: 'abs' }, items) : h('div', { className: 'hint', textContent: 'no edges picked' }),
    h('div', { className: 'row' }, ['each picked edge\'s Δφ, by its own by', step('phi', -1, '−'), step('phi', 1, '+'), ' Δθ', step('theta', -1, '−'), step('theta', 1, '+')]),
    h('div', { className: 'part', textContent: 'chosen for vector recalculation' }),
    redo.length ? h('ul', { className: 'abs' }, redo) : h('div', { className: 'hint', textContent: 'no edges' })];
}

const bShown = builder => [...builder.nodes].filter(n => n.out.length || n.in.length);