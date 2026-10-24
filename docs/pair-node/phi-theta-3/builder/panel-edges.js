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
  const move = (ts, signs) => { ts.forEach(t => Object.entries(signs(t)).forEach(([k, sign]) => t.ab.a.step(builder, t, k, sign))); bMoveAll(builder, ts); changed(); };
  const button = (text, title, onclick, lit, disabled = !picked.length) => {
    const b = h('button', { textContent: text, title, disabled, className: lit ? 'lit' : '' });
    b.onclick = onclick;
    return b;
  };
  const spin = (name, kids) => h('span', { className: 'pair' }, [name, h('span', { className: 'spin' }, kids)]);
  const step = (lock, locking, ts, k, sign, whose) => lock[k] === sign
    ? button(sign > 0 ? '+' : '−', 'locked: click to unlock', () => { lock[k] = 0; changed(); }, true)
    : button(sign > 0 ? '+' : '−', locking ? 'lock this direction' : `${sign > 0 ? 'add' : 'subtract'} ${whose} step`,
      () => { if (locking) { lock[k] = sign; changed(); } else move(ts, () => ({ [k]: sign })); });
  const lockButton = (on, flip, disabled) => {
    const b = button(on ? '●' : '○', disabled ? 'a row is locking: finish there first' : 'then click a + or − to lock it; click a lit one to unlock', () => { flip(); changed(); }, on, disabled);
    b.className += ' tall';
    return b;
  };
  const rowLock = t => (t.ab.row.lock ??= { phi: 0, theta: 0 }), lockAll = builder.lock ??= { phi: 0, theta: 0 };
  const items = picked.map(t => {
    const drop = h('button', { textContent: 'unpick' });
    drop.onclick = () => { picked.splice(picked.indexOf(t), 1); changed(); };
    const lock = rowLock(t), locking = !!t.ab.row.locking;
    const own = (k, name) => {
      const amount = h('input', { type: 'number', value: t.ab.row.increment[k], step: 1 });
      amount.onchange = () => { const v = +amount.value; if (bWhole(v)) t.ab.row.increment[k] = Math.round(v); changed(); };
      return h('td', {}, [h('span', { className: 'pair' }, [name, amount, h('span', { className: 'spin' }, [step(lock, locking, [t], k, 1, "this edge's"), step(lock, locking, [t], k, -1, "this edge's")])])]);
    };
    const lockCell = h('td', { className: 'lockcell' }, [h('span', { className: 'spin' }, [lockButton(locking, () => { t.ab.row.locking = !locking; })])]);
    return h('tr', {}, [h('td', { className: 't', textContent: abName(t) }), own('phi', 'Δφ'), own('theta', 'Δθ'), h('td', {}, [drop]), lockCell]);
  });
  const redo = left.map(t => h('span', { className: 't', textContent: abName(t) }));
  const way = (t, k) => rowLock(t)[k] || lockAll[k] || 1;
  const locked = lockAll.phi || lockAll.theta || picked.some(t => rowLock(t).phi || rowLock(t).theta);
  const both = [button(locked ? '▲' : '+', locked ? "step each angle its locked way: the row's lock, else move all's, else +" : "add each picked edge's steps", () => move(picked, t => ({ phi: way(t, 'phi'), theta: way(t, 'theta') }))),
    button(locked ? '▼' : '−', locked ? 'step each angle against its locked way' : "subtract each picked edge's steps", () => move(picked, t => ({ phi: -way(t, 'phi'), theta: -way(t, 'theta') })))];
  const rowsLocking = picked.some(t => t.ab.row.locking);
  if (rowsLocking) builder.locking = false;
  const allStep = (k, sign) => step(lockAll, !!builder.locking, picked, k, sign, "each picked edge's");
  const right = kids => h('td', { className: 'right' }, kids);
  const all = h('tr', { className: 'all' }, [h('td', { className: 't', textContent: 'all' }), right([spin('Δφ', [allStep('phi', 1), allStep('phi', -1)])]),
    right([spin('Δθ', [allStep('theta', 1), allStep('theta', -1)])]), right([spin('both', both)]),
    h('td', { className: 'lockcell' }, [spin('lock', [lockButton(!!builder.locking, () => { builder.locking = !builder.locking; }, rowsLocking)])])]);
  const folds = h('div', { className: 'folds' }, [bFold(h, bPickedShown, 'picked', items.length, () => { bPickedShown = !bPickedShown; }, changed),
    bFold(h, bRedoShown, 'chosen for vector recalculation', redo.length, () => { bRedoShown = !bRedoShown; }, changed)]);
  const pickedList = items.length ? [h('table', { className: 'cells picked' }, [h('tbody', {}, [all, ...(bPickedShown ? items : [])])])]
    : bPickedShown ? [h('div', { className: 'hint', textContent: 'no edges picked' })] : [];
  const redoList = !bRedoShown ? [] : redo.length ? [h('div', { className: 'row' }, redo)] : [h('div', { className: 'hint', textContent: 'no edges chosen for recalculation' })];
  return [h('div', { className: 'row' }, [pick]), folds, h('div', { className: 'lists' }, [...pickedList, ...redoList])];
}
