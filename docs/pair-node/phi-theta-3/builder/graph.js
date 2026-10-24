const bAB = row => ({ row, a: row.from, b: row.to });
const bMake = n => Object.assign(Object.create(B_NODE), n, { out: [], in: [] });
const bFirst = set => set.values().next().value;
const bName = k => (k < 26 ? 'abcdefghijklmnopqrstuvwxyz'[k] : `n${k}`);
const bNameAt = name => (name.length === 1 ? name.charCodeAt(0) - 97 : +name.slice(1));
const bOpenStep = builder => builder.seqs[builder.open].steps[0];
const bShown = builder => [...bOpenStep(builder).nodes];

function bLink(step, row) {
  row.delta ??= { phi: 0, theta: 0 };
  row.increment ??= { phi: 1, theta: 1 };
  const t = { ab: bAB(row) }, { a, b } = t.ab;
  a.addOut(t);
  b.addIn(t);
  step.edges.add(t);
  return t;
}

const bAddTree = (builder, row) => bLink(bOpenStep(builder), row);

function bDropTree(builder, t) {
  const { a, b } = t.ab;
  a.drop(t);
  b.drop(t);
  bOpenStep(builder).edges.delete(t);
}

function bDropNode(builder, n) {
  [...n.out, ...n.in].forEach(t => bDropTree(builder, t));
  bOpenStep(builder).nodes.delete(n);
  n.gone = true;
  if (builder.item === n) builder.item = bFirst(bOpenStep(builder).nodes);
}
