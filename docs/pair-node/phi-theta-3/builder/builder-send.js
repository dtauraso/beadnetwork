const bFirst = set => set.values().next().value;
const bName = k => (k < 26 ? 'abcdefghijklmnopqrstuvwxyz'[k] : `n${k}`);
const bNameAt = name => (name.length === 1 ? name.charCodeAt(0) - 97 : +name.slice(1));
const bOpenStep = builder => builder.seqs[builder.open].steps[0];

function bPack(k, v) {
  if (['gone', 'out', 'in'].includes(k)) return undefined;
  if (k === 'edges') return [...v].map(t => t.ab.row);
  if (k === 'picked') return v.map(t => `${t.ab.a.name}→${t.ab.b.name}`);
  if (v instanceof Set) return [...v];
  if (k === 'from' || k === 'to' || k === 'item') return v.name;
  return v;
}

function bWireStep(step) {
  const by = new Map(step.nodes.map(({ name, s, idx }) => [name, bMake({ name, s, idx })])), rows = step.edges.map(({ from, to, start, delta, increment }) => ({ from, to, start, delta, increment }));
  step.nodes = new Set(by.values());
  step.edges = new Set();
  rows.forEach(r => {
    r.from = by.get(r.from); r.to = by.get(r.to);
    if (r.from && r.to) bLink(step, r);
  });
}

function bWire(builder) {
  builder.seqs.forEach(q => q.steps.forEach(bWireStep));
  const step = bOpenStep(builder), by = new Map([...step.nodes].map(n => [n.name, n])), abs = new Map([...step.edges].map(t => [`${t.ab.a.name}→${t.ab.b.name}`, t]));
  builder.picked = (builder.picked || []).map(k => abs.get(k)).filter(Boolean);
  builder.methods ??= {};
  builder.item = by.get(builder.item) || bFirst(step.nodes);
  builder.made ??= 1 + Math.max(-1, ...builder.seqs.flatMap(q => q.steps.flatMap(st => [...st.nodes].map(n => bNameAt(n.name)))).filter(Number.isFinite));
}

function bDropNode(builder, n) {
  [...n.out, ...n.in].forEach(t => bDropTree(builder, t));
  bOpenStep(builder).nodes.delete(n);
  n.gone = true;
  if (builder.item === n) builder.item = bFirst(bOpenStep(builder).nodes);
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
