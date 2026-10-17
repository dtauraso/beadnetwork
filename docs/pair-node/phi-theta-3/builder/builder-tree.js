const bAB = row => ({ row, a: row.from, b: row.to });
const bAC = row => ({ row, a: row.from, c: row.to });
const bBC = ac => (ac.c.via && ac.c.via !== ac.row ? { row: ac.c.via, b: ac.c.via.from, c: ac.c } : null);

const bTrees = st => [...st.trees];
const bRows = st => bTrees(st).map(t => t.ab.row);
const bOut = (st, n) => bRows(st).filter(r => r.from === n);
const bIn = (st, n) => bRows(st).filter(r => r.to === n);
const bPairsOf = (st, row) => bTrees(st).flatMap(t => t.pairs).filter(p => p.ac.row === row);
const B_PAIR = ['base', 'ref', 'held', 'went', 'step', 'mine', 'automatic'];

function bSeat(st, p) {
  const s = p.ab.row.start, v = bView(st, p.ac.row);
  bSave(p);
  p.base = s && v && { phi: s.phi, theta: s.theta, r: v.r };
  p.ref = s && { phi: s.phi, theta: s.theta };
  p.held = { phi: 0, theta: 0 };
  p.mine = false;
  p.went = { phi: 0, theta: 0 };
  p.step = { phi: 0, theta: 0 };
}

function bJoin(st, t, row) {
  const p = { ab: t.ab, ac: bAC(row) };
  bSeat(st, p);
  bSave(t);
  t.pairs = [...t.pairs, p];
  t.bcs = [...t.bcs, bBC(p.ac)];
}

function bSplit(t, row) {
  const keep = t.pairs.map(p => p.ac.row !== row);
  bSave(t);
  t.pairs = t.pairs.filter((_, i) => keep[i]);
  t.bcs = t.bcs.filter((_, i) => keep[i]);
}

function bRelink(st, n) {
  bTrees(st).filter(t => t.pairs.some(p => p.ac.c === n)).forEach(t => {
    bSave(t);
    t.bcs = t.pairs.map(p => bBC(p.ac));
  });
}

function bAddRow(st, row) {
  const others = bOut(st, row.from);
  row.tree = { ab: bAB(row), pairs: [], bcs: [] };
  st.trees.add(row.tree);
  others.forEach(o => { bJoin(st, row.tree, o); bJoin(st, o.tree, row); });
}

function bDropRow(st, row) {
  if (row.to.via === row) bLetGo(st, row.to);
  st.trees.delete(row.tree);
  bOut(st, row.from).forEach(o => bSplit(o.tree, row));
}
