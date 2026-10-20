function bEdgePicker(st, h, picked, changed) {
  const trees = [...st.trees], abName = t => `${t.ab.a.name}→${t.ab.b.name}`;
  picked.splice(0, picked.length, ...picked.filter(t => trees.includes(t)));
  if (!trees.length) return [];
  const left = trees.filter(t => !picked.includes(t));
  const pick = h('select', {}, [h('option', { value: -1, textContent: left.length ? 'pick an edge' : 'all edges picked' }),
    ...left.map((t, i) => h('option', { value: i, textContent: abName(t) }))]);
  pick.onchange = () => { if (left[+pick.value]) picked.push(left[+pick.value]); changed(); };
  const by = (t, k, sign) => { t.ab.row.delta[k] += sign * t.ab.row.by[k]; };
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

const bShown = st => [...st.nodes].filter(n => n.out.length || n.in.length);

function bShifted(st, n, picked) {
  const t = picked.find(t => t.ab.b === n), P = t && bNodePos(st, t.ab.a), d = t && bView(st, t.ab.row);
  if (!P || !d) return bNodePos(st, n);
  const dl = t.ab.row.delta;
  return bAdd(P, bVec(st, bS(st, t.ab.a), { phi: d.phi + dl.phi, theta: d.theta + dl.theta, r: d.r }));
}

function bAtShifted(st, picked, f) {
  B_AT = new Map(picked.filter(t => t.ab.row.delta.phi || t.ab.row.delta.theta).map(t => [t.ab.b, bShifted(st, t.ab.b, picked)]).filter(([, p]) => p));
  try { return f(); } finally { B_AT = null; }
}

function bFromAngle(st, t, picked) {
  const row = t.ab.row;
  return picked.includes(t) ? bView(st, row) : bView(st, row, bShifted(st, row.from, picked), bShifted(st, row.to, picked));
}