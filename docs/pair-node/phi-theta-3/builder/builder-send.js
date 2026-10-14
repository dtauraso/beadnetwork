const bMake = n => Object.assign(n, { out: new Set(), rowsOut: new Set(), rowsIn: new Set(), angles: new Set() });
const bFirst = set => set.values().next().value;
const bName = k => (k < 26 ? 'abcdefghijklmnopqrstuvwxyz'[k] : `n${k}`);
const bNameAt = name => (name.length === 1 ? name.charCodeAt(0) - 97 : +name.slice(1));

function bPack(k, v) {
  if (v instanceof Set) return [...v];
  if (k === 'from' || k === 'to' || k === 'sender' || k === 'item') return v.name;
  if (k === 'via') return [v.from.name, v.to.name];
  if (['out', 'rowsOut', 'rowsIn', 'angles', 'gone'].includes(k)) return undefined;
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
    if (from && row && row.to === n && row.from === from) { n.sender = from; n.via = row; from.out.add(n); }
  });
  st.rows.forEach(r => { if (!bFloat(r)) bLink(st, r); });
  st.item = by.get(st.item) || bFirst(st.nodes);
  st.made ??= 1 + Math.max(-1, ...[...by.keys()].map(bNameAt).filter(Number.isFinite));
}

const B_DATA = ['idx', 'start', 'end', 'base', 'ref', 'held'], B_SETS = ['out', 'rowsOut', 'rowsIn', 'angles'];

function bKeep(st) {
  return [...st.nodes, ...st.rows].map(o => {
    const c = { ...o };
    B_DATA.forEach(k => { if (c[k]) c[k] = { ...c[k] }; });
    B_SETS.forEach(k => { if (c[k]) c[k] = new Set(c[k]); });
    return [o, c];
  });
}

function bRestore(st, keep) {
  keep.forEach(([o, c]) => {
    Object.keys(o).forEach(k => { if (!(k in c)) delete o[k]; });
    Object.assign(o, c);
  });
}

function bLetGo(n) {
  if (n.sender) n.sender.out.delete(n);
  delete n.sender;
  delete n.via;
}

function bRefer(st, row) {
  const n = row.to;
  if (bFloat(row) || n.via === row) return;
  bLetGo(n);
  n.sender = row.from;
  n.via = row;
  row.from.out.add(n);
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

function bTake(st, n) {
  const f = n.sender;
  if (!n.known || !f || !f.known) return false;
  const t = bDir(st, n.via);
  if (!t) return false;
  const J = bIndexOf(st, bAdd(bPlace(st, f.idx), bVec(st, bS(st, f), t)));
  n.idx = { ...J, theta: J.theta < 0 ? J.theta + st.Ttheta : J.theta };
  return true;
}

function bSend(st, n) {
  const inbox = [...n.out];
  for (let i = 0; i < inbox.length; i++) {
    if (i >= st.nodes.size) return false;
    if (bTake(st, inbox[i])) inbox.push(...inbox[i].out);
  }
  return true;
}

function bMove(st, row) {
  bRefer(st, row);
  return !bTake(st, row.to) || bSend(st, row.to);
}
