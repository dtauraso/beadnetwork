const bAB = row => ({ row, a: row.from, b: row.to });
const bAC = copy => ({ row: copy, a: copy.from, c: copy.to });
const bBC = ac => (ac.c.via && ac.c.via !== ac.row ? { row: ac.c.via, b: ac.c.via.from, c: ac.c } : null);

const bTrees = st => [...st.trees];
const bRows = st => bTrees(st).map(t => t.pair.ab.row);
const bOut = (st, n) => bRows(st).filter(r => r.from === n);
const bIn = (st, n) => bRows(st).filter(r => r.to === n);
const bHolder = (st, copy) => bTrees(st).find(t => t.pair.acs.some(ac => ac.row === copy));
const bSib = (st, copy) => bHolder(st, copy)?.pair.ab.row;

function bTouch(t) {
  bSave(t);
  bSave(t.pair);
}

function bFile(st, copy) {
  const home = bTrees(st).find(t => t.pair.ab.a === copy.from && t.pair.ab.row !== copy && bKind(t.pair.ab.row) === 'angle');
  if (!home) return;
  const ac = bAC(copy);
  bTouch(home);
  home.pair.acs = [...home.pair.acs, ac];
  home.bcs = [...home.bcs, bBC(ac)];
}

function bUnfile(st, copy) {
  const t = bHolder(st, copy);
  if (!t) return;
  const i = t.pair.acs.findIndex(ac => ac.row === copy);
  bTouch(t);
  t.pair.acs = t.pair.acs.filter((_, k) => k !== i);
  t.bcs = t.bcs.filter((_, k) => k !== i);
}

function bRehome(st, t) {
  const copies = t.pair.acs.map(ac => ac.row);
  bTouch(t);
  t.pair.acs = [];
  t.bcs = [];
  copies.forEach(copy => bFile(st, copy));
}

function bAdopt(st, row) {
  bOut(st, row.from).filter(r => bKind(r) === 'copy' && !bHolder(st, r)).forEach(r => bFile(st, r));
}

function bRelink(st, n) {
  bTrees(st).filter(t => t.pair.acs.some(ac => ac.c === n)).forEach(t => {
    bTouch(t);
    t.bcs = t.pair.acs.map(bBC);
  });
}

function bAddRow(st, row) {
  row.tree = { pair: { ab: bAB(row), acs: [] }, bcs: [] };
  st.trees.add(row.tree);
  if (bKind(row) === 'copy') bFile(st, row);
  if (bKind(row) === 'angle') bAdopt(st, row);
}

function bDropRow(st, row) {
  if (row.to.via === row) bLetGo(st, row.to);
  st.trees.delete(row.tree);
  bUnfile(st, row);
  bRehome(st, row.tree);
}

function bRekind(st, row, was) {
  const kind = bKind(row);
  if (was === kind) return;
  if (was === 'copy') bUnfile(st, row);
  if (was === 'angle') bRehome(st, row.tree);
  if (kind === 'copy') bFile(st, row);
  if (kind === 'angle') bAdopt(st, row);
}
