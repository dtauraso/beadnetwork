const bFirst = set => set.values().next().value;
const bName = k => (k < 26 ? 'abcdefghijklmnopqrstuvwxyz'[k] : `n${k}`);
const bNameAt = name => (name.length === 1 ? name.charCodeAt(0) - 97 : +name.slice(1));

function bPack(k, v) {
  if (['gone', 'out', 'in'].includes(k)) return undefined;
  if (k === 'trees') {
    const vars = p => Object.fromEntries(B_PAIR.filter(x => x in p).map(x => [x, p[x]]));
    return [...v].map(t => ({ ...t.ab.row, pairs: t.roots.map(({ pair }) => ({ with: pair.ac.c.name, ...vars(pair) })) }));
  }
  if (v instanceof Set) return [...v];
  if (k === 'from' || k === 'to' || k === 'sender' || k === 'item') return v.name;
  if (k === 'via') return [v.a.name, v.b.name];
  return v;
}

function bOldPairs(t) {
  const r = t.ab.row, x = t.roots.find(({ back }) => bKind(back.tree.ab.row) === 'angle');
  if (x) B_PAIR.filter(k => k in r).forEach(k => { x.back.pair[k] = r[k]; });
  B_PAIR.forEach(k => delete r[k]);
}

function bWire(st) {
  const by = new Map(st.nodes.map(n => [n.name, bMake(n)])), rows = st.trees || st.rows, abs = new Map(), saved = new Map();
  st.nodes = new Set(by.values());
  st.trees = new Set();
  delete st.rows;
  rows.forEach(r => {
    const key = `${r.from}→${r.to}`, ps = r.pairs;
    r.from = by.get(r.from); r.to = by.get(r.to);
    delete r.pairs;
    B_GONE.forEach(k => delete r[k]);
    if (!r.from || !r.to) return;
    const t = bAddTree(st, r);
    abs.set(key, t.ab);
    saved.set(t, ps || []);
  });
  st.nodes.forEach(n => {
    const from = n.sender && by.get(n.sender), ab = n.via && abs.get(n.via.join('→'));
    delete n.ref; delete n.sender; delete n.via;
    if (from && ab && ab.b === n && ab.a === from) { n.sender = from; n.via = ab; }
  });
  saved.forEach((ps, t) => ps.forEach(({ with: c, ...w }) => {
    const x = t.roots.find(({ pair }) => pair.ac.c.name === c);
    if (x) B_PAIR.filter(k => k in w).forEach(k => { x.pair[k] = w[k]; });
  }));
  st.trees.forEach(t => { if (bKind(t.ab.row) === 'copy' && t.ab.row.ref) bOldPairs(t); B_PAIR.filter(k => k !== 'base').forEach(k => delete t.ab.row[k]); });
  st.trees.forEach(t => { if (!bFloat(t.ab.row)) bLink(st, t.ab); });
  st.item = by.get(st.item) || bFirst(st.nodes);
  st.made ??= 1 + Math.max(-1, ...[...by.keys()].map(bNameAt).filter(Number.isFinite));
}

const B_DATA = ['idx', 'start', 'end', 'base', 'ref', 'delta'];

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
