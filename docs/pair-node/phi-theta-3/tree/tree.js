const T_KEY = 'pair-builder-v5';
const tEl = (tag, props = {}, kids = []) => {
  const e = Object.assign(document.createElement(tag), props);
  kids.forEach(k => e.append(k));
  return e;
};
const tNum = v => (Number.isInteger(v) ? `${v}` : v.toFixed(3));
const tAng = (o, ks = ['phi', 'theta', 'r']) => (o ? ks.filter(k => k in o).map(k => `${{ phi: 'φ', theta: 'θ', r: 'r' }[k]} ${tNum(o[k])}`).join(' · ') : '—');
const tAB = ab => `${ab.a.name}→${(ab.b || ab.c).name}`;
const tId = ab => `t-${ab.a.name}-${(ab.b || ab.c).name}`;
const tRootId = r => `r-${tId(r.tree.ab)}-${tId(r.pair.ac)}`;
const tLink = (id, text) => tEl('a', { href: `#${id}`, textContent: text });

function tItem(label, value, kids = [], { id, cls = '' } = {}) {
  const vs = [].concat(value).map(v => (typeof v === 'string' ? tEl('span', { className: 'v', textContent: v }) : v));
  return tEl('li', id ? { id } : {}, [tEl('div', { className: `it ${cls}` }, [tEl('span', { className: 'k', textContent: label }), ...vs]),
    ...(kids.length ? [tEl('ul', {}, kids)] : [])]);
}
const tCalc = (label, value) => tItem(label, value, [], { cls: 'calc' });
const tLinks = (ts, id, name) => (ts.length ? ts.flatMap((t, i) => [...(i ? [', '] : []), tLink(id(t), name(t))]) : '—');

function tLoad() {
  try {
    const s = JSON.parse(localStorage.getItem(T_KEY));
    if (s) { bWire(s); return { st: s, from: `the builder's save in this browser (${T_KEY})` }; }
  } catch { }
  return { st: bPreset(), from: 'the three-node preset — the builder has saved nothing in this browser yet' };
}

function tState(st) {
  return [tItem('st', '', [
    tItem('s', tNum(st.s)), tItem('ρ', tNum(st.rho)), tItem('T_φ', tNum(st.Tphi)), tItem('T_θ', tNum(st.Ttheta)), tItem('c_r', tNum(st.cr)),
    tItem('item', st.item ? tLink(`n-${st.item.name}`, st.item.name) : '—'), tItem('made', `${st.made}`),
    tItem('step', `${st.step}`), tItem('nodes', `${st.nodes.size}`), tItem('trees', `${st.trees.size}`),
  ], { cls: 'head' })];
}

function tNode(st, n) {
  const C = bNodePos(st, n);
  return tItem('node', n.name, [
    tItem('s', n.s ? tNum(n.s) : `— (the scene's s, ${tNum(st.s)})`),
    tItem('known', n.known ? 'yes' : 'no'),
    tItem('idx', n.known ? `I_r ${tNum(n.idx.r)} · I_φ ${tNum(n.idx.phi)} · I_θ ${tNum(n.idx.theta)}` : '—'),
    tCalc('C', C ? bPt(C) : '—'),
    tItem('sender', n.sender ? tLink(`n-${n.sender.name}`, n.sender.name) : '—'),
    tItem('via', n.via ? tLink(tId(n.via), tAB(n.via)) : '—'),
    tItem('out', tLinks(n.out, t => tId(t.ab), t => tAB(t.ab))),
    tItem('in', tLinks(n.in, t => tId(t.ab), t => tAB(t.ab))),
  ], { id: `n-${n.name}`, cls: 'head' });
}

function tRow(st, t) {
  const { row, a, b } = t.ab, dirs = bDirs(st, t);
  const has = k => (row[k] ? [tItem(k, tAng(row[k], k === 'start' || k === 'base' ? ['phi', 'theta', 'r'] : ['phi', 'theta']))] : []);
  return tItem('ab', tAB(t.ab), [
    tItem('a', tLink(`n-${a.name}`, a.name)), tItem('b', tLink(`n-${b.name}`, b.name)),
    tItem('row', '', [tItem('kind', bKind(row)), ...has('start'), ...has('end'), ...has('delta'), ...has('base')]),
    tCalc('direction', dirs.length ? dirs.map(d => tAng(d.t)).join(' / ') : '—'),
    tCalc('placed', tAng(bView(st, row))),
    tItem('places b', b.via === t.ab ? 'yes — b\'s via' : 'no'),
  ]);
}

function tRoot(st, t, r) {
  const p = r.pair, ac = p.ac, d = bChange(st, p), copy = bKind(ac.row) === 'copy';
  return tItem('root', `(${tAB(t.ab)}, ${tAB(ac)})`, [
    tItem('pair', '', [
      tItem('ab', tLink(tId(t.ab), tAB(t.ab))), tItem('ac', tLink(tId(ac), tAB(ac))),
      tItem('base', tAng(p.base)), tItem('ref', tAng(p.ref, ['phi', 'theta'])),
      tCalc('change', d ? `${tAB(t.ab)} start − ref = ${tAng(d, ['phi', 'theta'])}` : '—'),
      tItem(copy ? 'used by' : 'unused', copy ? `${tAB(ac)}'s copy direction: base + change` : `${tAB(ac)} is ${bKind(ac.row)}, not a copy`, [], { cls: 'note' }),
    ]),
    tItem('back', tLink(tRootId(r.back), `(${tAB(r.back.tree.ab)}, ${tAB(r.back.pair.ac)}) in ${tAB(r.back.tree.ab)}'s tree`)),
  ], { id: tRootId(r) });
}

function tTree(st, t) {
  return tItem('tree', tAB(t.ab), [tRow(st, t),
    tItem('roots', t.roots.length ? '' : `— (${t.ab.a.name} has no other row)`, t.roots.map(r => tRoot(st, t, r)))], { id: tId(t.ab), cls: 'head' });
}

function tDraw() {
  const { st, from } = tLoad();
  document.getElementById('source').textContent = `from ${from}, read at ${new Date().toLocaleTimeString()}`;
  document.getElementById('state').replaceChildren(...tState(st));
  document.getElementById('nodes').replaceChildren(...[...st.nodes].map(n => tNode(st, n)));
  document.getElementById('trees').replaceChildren(...[...st.trees].map(t => tTree(st, t)));
  if (location.hash) document.getElementById(location.hash.slice(1))?.scrollIntoView();
}

window.addEventListener('storage', e => { if (e.key === T_KEY) tDraw(); });
tDraw();
