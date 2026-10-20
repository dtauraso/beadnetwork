const bFirst = set => set.values().next().value;
const bName = k => (k < 26 ? 'abcdefghijklmnopqrstuvwxyz'[k] : `n${k}`);
const bNameAt = name => (name.length === 1 ? name.charCodeAt(0) - 97 : +name.slice(1));

function bPack(k, v) {
  if (['gone', 'out', 'in'].includes(k)) return undefined;
  if (k === 'trees') return [...v].map(t => t.ab.row);
  if (k === 'picked') return v.map(t => `${t.ab.a.name}→${t.ab.b.name}`);
  if (v instanceof Set) return [...v];
  if (k === 'from' || k === 'to' || k === 'sender' || k === 'item') return v.name;
  if (k === 'via') return [v.a.name, v.b.name];
  return v;
}

function bWire(st) {
  const by = new Map(st.nodes.map(n => [n.name, bMake(n)])), rows = st.trees || st.rows, abs = new Map(), copies = [];
  st.nodes = new Set(by.values());
  st.trees = new Set();
  delete st.rows;
  rows.forEach(r => {
    const key = `${r.from}→${r.to}`;
    r.from = by.get(r.from); r.to = by.get(r.to);
    B_GONE.forEach(k => delete r[k]);
    if (r.kind === 'copy') { r.kind = 'base'; delete r.base; copies.push(r); }
    if (!r.from || !r.to) return;
    const t = bAddTree(st, r);
    abs.set(key, t);
  });
  st.nodes.forEach(n => {
    const from = n.sender && by.get(n.sender), ab = n.via && abs.get(n.via.join('→'))?.ab;
    delete n.ref; delete n.sender; delete n.via;
    if (from && ab && ab.b === n && ab.a === from) { n.sender = from; n.via = ab; }
  });
  st.trees.forEach(t => { if (!bFloat(t.ab.row)) bLink(st, t.ab); });
  copies.forEach(r => { if (r.from && r.to) r.base = bView(st, r); });
  st.picked = (st.picked || []).map(k => abs.get(k)).filter(Boolean);
  st.item = by.get(st.item) || bFirst(st.nodes);
  st.made ??= 1 + Math.max(-1, ...[...by.keys()].map(bNameAt).filter(Number.isFinite));
}

const B_DATA = ['idx', 'start', 'end', 'base', 'delta', 'by'];

let bLog = null;
const bKeep = () => (bLog = new Map());
const bDone = () => { bLog = null; };

function bSave(o) {
  if (!bLog || bLog.has(o)) return;
  const c = { ...o };
  B_DATA.forEach(k => { if (c[k]) c[k] = { ...c[k] }; });
  bLog.set(o, c);
}

function bRestore(keep) {
  keep.forEach((c, o) => {
    Object.keys(o).forEach(k => { if (!(k in c)) delete o[k]; });
    Object.assign(o, c);
  });
  bDone();
}

function bLetGo(st, n) {
  bSave(n);
  delete n.sender;
  delete n.via;
}

function bRefer(st, ab) {
  const n = ab.b;
  if (bFloat(ab.row) || n.via === ab) return;
  bLetGo(st, n);
  n.sender = ab.a;
  n.via = ab;
}

function bLink(st, ab) {
  if (!ab.b.sender) bRefer(st, ab);
}

function bDropNode(st, n) {
  [...n.out, ...n.in].forEach(t => bDropTree(st, t));
  st.nodes.delete(n);
  n.gone = true;
  if (st.item === n) st.item = bFirst(st.nodes);
}

function bSetAt(st, n, p) {
  const J = bIndexOf(st, p);
  bSave(n);
  n.idx = { ...J, theta: J.theta < 0 ? J.theta + st.Ttheta : J.theta };
}

function bPut(st, row) {
  const n = row.to, f = row.from;
  if (!n.known || !f.known) return false;
  const t = bDir(st, row);
  if (!t) return false;
  bSetAt(st, n, bAdd(bPlace(st, f.idx), bVec(st, bS(st, f), t)));
  return true;
}

function bMove(st, tree) {
  bRefer(st, tree.ab);
  bPut(st, tree.ab.row);
}
