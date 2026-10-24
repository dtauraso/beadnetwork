let bPickedShown = true, bRedoShown = true, bEdgesShown = true, bNodesShown = true;
const bAngleOpen = new Set();
const bLock = { phi: 0, theta: 0 };
let bLocking = false;

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
  const move = signs => { picked.forEach(t => { const r = t.ab.row; Object.entries(signs).forEach(([k, sign]) => { r.delta[k] = bWrap(bS(builder, t.ab.a), r.delta[k] + sign * r.increment[k]); }); }); bMoveAll(builder, picked); changed(); };
  const button = (text, title, onclick, lit) => {
    const b = h('button', { textContent: text, title, disabled: !picked.length, className: lit ? 'lit' : '' });
    b.onclick = onclick;
    return b;
  };
  const step = (k, sign) => bLock[k] === sign
    ? button(sign > 0 ? '+' : '−', 'locked: click to unlock', () => { bLock[k] = 0; changed(); }, true)
    : button(sign > 0 ? '+' : '−', bLocking ? 'lock this direction' : `${sign > 0 ? 'add' : 'subtract'} each picked edge's step`,
      () => { if (bLocking) { bLock[k] = sign; changed(); } else move({ [k]: sign }); });
  const locked = bLock.phi || bLock.theta, way = sign => ({ phi: sign * (bLock.phi || 1), theta: sign * (bLock.theta || 1) });
  const both = [button(locked ? '▲' : '+', locked ? 'step each angle its locked way (an unlocked angle steps +)' : "add each picked edge's steps", () => move(way(1))),
    button(locked ? '▼' : '−', locked ? 'step each angle against its locked way' : "subtract each picked edge's steps", () => move(way(-1)))];
  const spin = (name, kids) => h('span', { className: 'pair' }, [name, h('span', { className: 'spin' }, kids)]);
  const lock = button('lock', 'then click a + or − to lock it; click a lit one to unlock', () => { bLocking = !bLocking; changed(); }, bLocking);
  const all = items.length ? [h('span', { className: 'all' }, ['move all', lock, spin('Δφ', [step('phi', 1), step('phi', -1)]), spin('Δθ', [step('theta', 1), step('theta', -1)]), spin('both', both)])] : [];
  const folds = h('div', { className: 'folds' }, [bFold(h, bPickedShown, 'picked', items.length, () => { bPickedShown = !bPickedShown; }, changed),
    bFold(h, bRedoShown, 'chosen for vector recalculation', redo.length, () => { bRedoShown = !bRedoShown; }, changed)]);
  const pickedList = !bPickedShown ? [] : items.length ? [h('table', { className: 'cells' }, [h('tbody', {}, items)])] : [h('div', { className: 'hint', textContent: 'no edges picked' })];
  const redoList = !bRedoShown ? [] : redo.length ? [h('div', { className: 'row' }, redo)] : [h('div', { className: 'hint', textContent: 'no edges chosen for recalculation' })];
  return [h('div', { className: 'row' }, [pick, ...all]), folds, h('div', { className: 'lists' }, [...pickedList, ...redoList])];
}
