let bPickedShown = true, bRedoShown = true, bEdgesShown = true, bNodesShown = true;
const bAngleOpen = new Set();

function bFold(h, open, text, n, flip, changed) {
  const b = h('button', { className: 'part fold', textContent: `${open ? '▾' : '▸'} ${text} (${n})` });
  b.onclick = () => { flip(); changed(); };
  return b;
}

function bEdgePicker(builder, h, picked, changed) {
  const trees = [...bOpenStep(builder).edges], abName = t => `${t.ab.a.name}→${t.ab.b.name}`;
  picked.splice(0, picked.length, ...picked.filter(t => trees.includes(t)));
  if (!trees.length) return [];
  const left = trees.filter(t => !picked.includes(t));
  const pick = h('select', {}, [h('option', { value: -1, textContent: left.length ? 'pick an edge' : 'all edges picked' }),
    ...left.map((t, i) => h('option', { value: i, textContent: abName(t) }))]);
  pick.onchange = () => { if (left[+pick.value]) picked.push(left[+pick.value]); changed(); };
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
  const step = (ks, sign, text) => {
    const b = h('button', { textContent: text, title: `${sign > 0 ? 'add' : 'subtract'} each picked edge's step`, disabled: !picked.length });
    b.onclick = () => { picked.forEach(t => { const r = t.ab.row; ks.forEach(k => { r.delta[k] = bWrap(bS(builder, t.ab.a), r.delta[k] + sign * r.increment[k]); }); }); bMoveAll(builder, picked); changed(); };
    return b;
  };
  const spin = (ks, name) => h('span', { className: 'pair' }, [name, h('span', { className: 'spin' }, [step(ks, 1, '+'), step(ks, -1, '−')])]);
  const all = items.length ? [h('span', { className: 'all' }, ['move all', spin(['phi'], 'Δφ'), spin(['theta'], 'Δθ'), spin(['phi', 'theta'], 'both')])] : [];
  const folds = h('div', { className: 'folds' }, [bFold(h, bPickedShown, 'picked', items.length, () => { bPickedShown = !bPickedShown; }, changed),
    bFold(h, bRedoShown, 'chosen for vector recalculation', redo.length, () => { bRedoShown = !bRedoShown; }, changed)]);
  const pickedList = !bPickedShown ? [] : items.length ? [h('table', { className: 'cells' }, [h('tbody', {}, items)])] : [h('div', { className: 'hint', textContent: 'no edges picked' })];
  const redoList = !bRedoShown ? [] : redo.length ? [h('div', { className: 'row' }, redo)] : [h('div', { className: 'hint', textContent: 'no edges chosen for recalculation' })];
  return [h('div', { className: 'row' }, [pick, ...all]), folds, h('div', { className: 'lists' }, [...pickedList, ...redoList])];
}
