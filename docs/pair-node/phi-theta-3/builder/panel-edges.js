let bPickedShown = true;

function bEdgePicker(builder, h, picked, changed) {
  const trees = [...bOpenStep(builder).edges], abName = t => `${t.ab.a.name}→${t.ab.b.name}`;
  picked.splice(0, picked.length, ...picked.filter(t => trees.includes(t)));
  if (!trees.length) return [];
  const left = trees.filter(t => !picked.includes(t));
  const pick = h('select', {}, [h('option', { value: -1, textContent: left.length ? 'pick an edge' : 'all edges picked' }),
    ...left.map((t, i) => h('option', { value: i, textContent: abName(t) }))]);
  pick.onchange = () => { if (left[+pick.value]) picked.push(left[+pick.value]); changed(); };
  const bump = (t, k, sign) => { t.ab.row.delta[k] += sign * t.ab.row.increment[k]; bMove(builder, t); };
  const items = picked.map(t => {
    const drop = h('button', { textContent: 'unpick' });
    drop.onclick = () => { picked.splice(picked.indexOf(t), 1); changed(); };
    const own = (k, name) => {
      const amount = h('input', { type: 'number', value: t.ab.row.increment[k], step: 1 });
      amount.onchange = () => { const v = +amount.value; if (bWhole(v)) t.ab.row.increment[k] = Math.round(v); changed(); };
      return h('td', {}, [h('span', { className: 'pair' }, [name, amount])]);
    };
    return h('tr', {}, [h('td', { className: 't', textContent: abName(t) }), own('phi', 'Δφ'), own('theta', 'Δθ'), h('td', {}, [drop])]);
  });
  const redo = left.map(t => h('span', { className: 't', textContent: abName(t) }));
  const step = (k, sign, text) => {
    const b = h('button', { textContent: text, title: `${sign > 0 ? 'add' : 'subtract'} each picked edge's step`, disabled: !picked.length });
    b.onclick = () => { picked.forEach(t => bump(t, k, sign)); changed(); };
    return b;
  };
  const spin = (k, name) => h('span', { className: 'pair' }, [name, h('span', { className: 'spin' }, [step(k, 1, '+'), step(k, -1, '−')])]);
  const all = items.length ? [h('span', { className: 'all' }, ['move all', spin('phi', 'Δφ'), spin('theta', 'Δθ')])] : [];
  const fold = h('button', { textContent: bPickedShown ? 'hide picked' : `show picked (${items.length})` });
  fold.onclick = () => { bPickedShown = !bPickedShown; changed(); };
  const list = !items.length ? [h('div', { className: 'hint', textContent: 'no edges picked' })]
    : bPickedShown ? [h('table', { className: 'cells' }, [h('tbody', {}, items)])] : [];
  return [h('div', { className: 'row' }, [pick, ...(items.length ? [fold] : []), ...all]), ...list,
    h('div', { className: 'part', textContent: 'chosen for vector recalculation' }),
    redo.length ? h('div', { className: 'row' }, redo) :h('div', { className: 'hint', textContent: 'no edges' })];
}
