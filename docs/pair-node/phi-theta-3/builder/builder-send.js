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

function bWire(builder) {
  const by = new Map(builder.nodes.map(n => [n.name, bMake(n)])), rows = builder.trees || builder.rows, abs = new Map(), kinds = [];
  builder.nodes = new Set(by.values());
  builder.trees = new Set();
  delete builder.rows;
  rows.forEach(r => {
    const key = `${r.from}→${r.to}`;
    r.from = by.get(r.from); r.to = by.get(r.to);
    if (r.kind && r.kind !== 'angle') kinds.push(r);
    B_GONE.forEach(k => delete r[k]);
    if (!r.from || !r.to) return;
    const t = bAddTree(builder, r);
    abs.set(key, t);
  });
  builder.nodes.forEach(n => {
    const from = n.sender && by.get(n.sender), ab = n.via && abs.get(n.via.join('→'))?.ab;
    delete n.ref; delete n.sender; delete n.via;
    if (from && ab && ab.b === n && ab.a === from) { n.sender = from; n.via = ab; }
  });
  const near = x => (Math.abs(x - Math.round(x)) < 1e-9 ? Math.round(x) : x);
  kinds.filter(r => r.from && r.to).forEach(r => {
    r.start = { phi: 0, theta: 0, r: 0 };
    const v = bMeasure(builder, r);
    if (v) r.start = { phi: near(v.phi), theta: near(v.theta), r: near(v.r) };
  });
  builder.trees.forEach(t => bLink(builder, t.ab));
  builder.picked = (builder.picked || []).map(k => abs.get(k)).filter(Boolean);
  builder.item = by.get(builder.item) || bFirst(builder.nodes);
  builder.made ??= 1 + Math.max(-1, ...[...by.keys()].map(bNameAt).filter(Number.isFinite));
}

function bLetGo(builder, n) {
  delete n.sender;
  delete n.via;
}

function bRefer(builder, ab) {
  const n = ab.b;
  if (n.via === ab) return;
  bLetGo(builder, n);
  n.sender = ab.a;
  n.via = ab;
}

function bLink(builder, ab) {
  if (!ab.b.sender) bRefer(builder, ab);
}

function bDropNode(builder, n) {
  [...n.out, ...n.in].forEach(t => bDropTree(builder, t));
  builder.nodes.delete(n);
  n.gone = true;
  if (builder.item === n) builder.item = bFirst(builder.nodes);
}

function bSetAt(builder, n, p) {
  const J = bIndexOf(builder, p);
  n.idx = { ...J, theta: J.theta < 0 ? J.theta + builder.Ttheta : J.theta };
}

function bPut(builder, row) {
  const n = row.to, f = row.from;
  if (!n.known || !f.known) return false;
  const t = bDir(builder, row);
  if (!t) return false;
  bSetAt(builder, n, bAdd(bPlace(builder, f.idx), bVec(builder, bS(builder, f), t)));
  return true;
}

function bRemeasure(builder, n, keep) {
  [...n.in, ...n.out].filter(t => t !== keep).forEach(t => {
    const v = bMeasure(builder, t.ab.row);
    const dl = t.ab.row.delta;
    if (v) t.ab.row.start = { phi: v.phi - dl.phi, theta: v.theta - dl.theta, r: v.r };
  });
}

function bMove(builder, tree) {
  bRefer(builder, tree.ab);
  if (bPut(builder, tree.ab.row)) bRemeasure(builder, tree.ab.b, tree);
}
