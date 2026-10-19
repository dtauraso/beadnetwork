function bEdgePicker(st, h, picked, changed) {
  const trees = [...st.trees], abName = t => `${t.ab.a.name}→${t.ab.b.name}`;
  picked.splice(0, picked.length, ...picked.filter(t => trees.includes(t)));
  if (!trees.length) return [];
  const left = trees.filter(t => !picked.includes(t));
  const pick = h('select', {}, [h('option', { value: -1, textContent: left.length ? 'pick an edge' : 'all edges picked' }),
    ...left.map((t, i) => h('option', { value: i, textContent: abName(t) }))]);
  pick.onchange = () => { if (left[+pick.value]) picked.push(left[+pick.value]); changed(); };
  const items = picked.map(t => {
    const drop = h('button', { textContent: 'unpick' });
    drop.onclick = () => { picked.splice(picked.indexOf(t), 1); changed(); };
    return h('li', {}, [h('div', { className: 'row' }, [h('span', { className: 't', textContent: abName(t) }), drop])]);
  });
  const redo = left.map(t => h('li', {}, [h('div', { className: 'row' }, [h('span', { className: 't', textContent: abName(t) })])]));
  const step = (k, by, text) => {
    const b = h('button', { textContent: text, disabled: !picked.length });
    b.onclick = () => { picked.forEach(t => { t.ab.row.delta[k] += by; }); changed(); };
    return b;
  };
  return [h('div', { className: 'row' }, [h('label', {}, ['edge ', pick])]),
    items.length ? h('ul', { className: 'abs' }, items) : h('div', { className: 'hint', textContent: 'no edges picked' }),
    h('div', { className: 'row' }, ['picked heads\' from Δφ', step('phi', -1, '−'), step('phi', 1, '+'), ' Δθ', step('theta', -1, '−'), step('theta', 1, '+')]),
    h('div', { className: 'part', textContent: 'chosen for vector recalculation' }),
    redo.length ? h('ul', { className: 'abs' }, redo) : h('div', { className: 'hint', textContent: 'no edges' })];
}

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