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
  const by = new Map(st.nodes.map(n => [n.name, bMake(n)])), rows = st.trees || st.rows, abs = new Map(), kinds = [];
  st.nodes = new Set(by.values());
  st.trees = new Set();
  delete st.rows;
  rows.forEach(r => {
    const key = `${r.from}→${r.to}`;
    r.from = by.get(r.from); r.to = by.get(r.to);
    if (r.kind && r.kind !== 'angle') kinds.push(r);
    B_GONE.forEach(k => delete r[k]);
    if (!r.from || !r.to) return;
    const t = bAddTree(st, r);
    abs.set(key, t);
  });
  st.nodes.forEach(n => {
    const from = n.sender && by.get(n.sender), ab = n.via && abs.get(n.via.join('→'))?.ab;
    delete n.ref; delete n.sender; delete n.via;
    if (from && ab && ab.b === n && ab.a === from) { n.sender = from; n.via = ab; }
  });
  const near = x => (Math.abs(x - Math.round(x)) < 1e-9 ? Math.round(x) : x);
  kinds.filter(r => r.from && r.to).forEach(r => {
    r.start = { phi: 0, theta: 0, r: 0 };
    const v = bMeasure(st, r);
    if (v) r.start = { phi: near(v.phi), theta: near(v.theta), r: near(v.r) };
  });
  st.trees.forEach(t => bLink(st, t.ab));
  st.picked = (st.picked || []).map(k => abs.get(k)).filter(Boolean);
  st.item = by.get(st.item) || bFirst(st.nodes);
  st.made ??= 1 + Math.max(-1, ...[...by.keys()].map(bNameAt).filter(Number.isFinite));
}

function bLetGo(st, n) {
  delete n.sender;
  delete n.via;
}

function bRefer(st, ab) {
  const n = ab.b;
  if (n.via === ab) return;
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

function bRemeasure(st, n, keep) {
  [...n.in, ...n.out].filter(t => t !== keep).forEach(t => {
    const v = bMeasure(st, t.ab.row);
    const dl = t.ab.row.delta;
    if (v) t.ab.row.start = { phi: v.phi - dl.phi, theta: v.theta - dl.theta, r: v.r };
  });
}

function bMove(st, tree) {
  bRefer(st, tree.ab);
  if (bPut(st, tree.ab.row)) bRemeasure(st, tree.ab.b, tree);
}
