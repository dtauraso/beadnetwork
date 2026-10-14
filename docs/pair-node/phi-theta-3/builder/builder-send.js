const bMake = n => Object.assign(n, { rowsOut: new Set(), rowsIn: new Set(), angles: new Set() });
const bFirst = set => set.values().next().value;
const bName = k => (k < 26 ? 'abcdefghijklmnopqrstuvwxyz'[k] : `n${k}`);
const bNameAt = name => (name.length === 1 ? name.charCodeAt(0) - 97 : +name.slice(1));

function bPack(k, v) {
  if (['rowsOut', 'rowsIn', 'angles', 'gone'].includes(k)) return undefined;
  if (v instanceof Set) return [...v];
  if (k === 'from' || k === 'to' || k === 'sender' || k === 'item') return v.name;
  if (k === 'via') return [v.from.name, v.to.name];
  return v;
}

function bAddRow(st, row) {
  st.rows.add(row);
  row.from.rowsOut.add(row);
  row.to.rowsIn.add(row);
  if (bKind(row) === 'angle') row.from.angles.add(row);
}

function bWire(st) {
  const by = new Map(st.nodes.map(n => [n.name, bMake(n)])), rows = st.rows, via = new Map();
  st.nodes = new Set(by.values());
  st.rows = new Set();
  rows.forEach(r => {
    const key = `${r.from}→${r.to}`;
    r.from = by.get(r.from); r.to = by.get(r.to);
    if (r.from && r.to) { bAddRow(st, r); via.set(key, r); }
  });
  st.nodes.forEach(n => {
    const from = n.sender && by.get(n.sender), row = n.via && via.get(n.via.join('→'));
    delete n.ref; delete n.sender; delete n.via;
    if (from && row && row.to === n && row.from === from) { n.sender = from; n.via = row; }
  });
  st.rows.forEach(r => { if (!bFloat(r)) bLink(st, r); });
  st.item = by.get(st.item) || bFirst(st.nodes);
  st.made ??= 1 + Math.max(-1, ...[...by.keys()].map(bNameAt).filter(Number.isFinite));
}

const B_DATA = ['idx', 'start', 'end', 'base', 'ref', 'held'], B_SETS = ['rowsOut', 'rowsIn', 'angles'];

let bLog = null;
const bKeep = () => (bLog = new Map());
const bDone = () => { bLog = null; };

function bSave(o) {
  if (!bLog || bLog.has(o)) return;
  const c = { ...o };
  B_DATA.forEach(k => { if (c[k]) c[k] = { ...c[k] }; });
  B_SETS.forEach(k => { if (c[k]) c[k] = new Set(c[k]); });
  bLog.set(o, c);
}

function bRestore(keep) {
  keep.forEach((c, o) => {
    Object.keys(o).forEach(k => { if (!(k in c)) delete o[k]; });
    Object.assign(o, c);
  });
  bDone();
}

function bLetGo(n) {
  bSave(n);
  delete n.sender;
  delete n.via;
}

function bRefer(st, row) {
  const n = row.to;
  if (bFloat(row) || n.via === row) return;
  bLetGo(n);
  n.sender = row.from;
  n.via = row;
}

function bLink(st, row) {
  if (!row.to.sender) bRefer(st, row);
}

function bDropRow(st, row) {
  if (row.to.via === row) bLetGo(row.to);
  row.from.rowsOut.delete(row);
  row.from.angles.delete(row);
  row.to.rowsIn.delete(row);
  st.rows.delete(row);
}

function bDropNode(st, n) {
  [...n.rowsOut, ...n.rowsIn].forEach(r => bDropRow(st, r));
  st.nodes.delete(n);
  n.gone = true;
  if (st.item === n) st.item = bFirst(st.nodes);
}

function bPut(st, row) {
  const n = row.to, f = row.from;
  if (!n.known || !f.known) return false;
  const t = bDir(st, row);
  if (!t) return false;
  const J = bIndexOf(st, bAdd(bPlace(st, f.idx), bVec(st, bS(st, f), t)));
  bSave(n);
  n.idx = { ...J, theta: J.theta < 0 ? J.theta + st.Ttheta : J.theta };
  return true;
}

const bTake = (st, n) => !!n.sender && bPut(st, n.via);

function bMove(st, row) {
  bRefer(st, row);
  bTake(st, row.to);
  row.from.rowsOut.forEach(r => { if (bKind(r) === 'copy' && bAuto(r) && bSib(st, r) === row) bPut(st, r); });
}
