const bFirst = set => set.values().next().value;
const bName = k => (k < 26 ? 'abcdefghijklmnopqrstuvwxyz'[k] : `n${k}`);
const bNameAt = name => (name.length === 1 ? name.charCodeAt(0) - 97 : +name.slice(1));

function bPack(k, v) {
  if (['gone', 'out', 'in'].includes(k)) return undefined;
  if (k === 'trees') return [...v].map(t => t.ab.row);
  if (k === 'picked') return v.map(t => `${t.ab.a.name}→${t.ab.b.name}`);
  if (v instanceof Set) return [...v];
  if (k === 'from' || k === 'to' || k === 'item') return v.name;
  return v;
}

function bWire(builder) {
  const by = new Map(builder.nodes.map(({ name, s, idx }) => [name, bMake({ name, s, idx })])), rows = builder.trees.map(({ from, to, start, delta, increment }) => ({ from, to, start, delta, increment })), abs = new Map();
  builder.nodes = new Set(by.values());
  builder.trees = new Set();
  rows.forEach(r => {
    const key = `${r.from}→${r.to}`;
    r.from = by.get(r.from); r.to = by.get(r.to);
    if (!r.from || !r.to) return;
    const t = bAddTree(builder, r);
    abs.set(key, t);
  });
  builder.picked = (builder.picked || []).map(k => abs.get(k)).filter(Boolean);
  builder.item = by.get(builder.item) || bFirst(builder.nodes);
  builder.made ??= 1 + Math.max(-1, ...[...by.keys()].map(bNameAt).filter(Number.isFinite));
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
  if (bPut(builder, tree.ab.row)) bRemeasure(builder, tree.ab.b, tree);
}
