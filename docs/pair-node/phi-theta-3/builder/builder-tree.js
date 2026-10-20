const bAB = row => ({ row, a: row.from, b: row.to });
const bMake = n => Object.assign(n, { out: [], in: [] });

function bLink(step, row) {
  row.delta ??= { phi: 0, theta: 0 };
  row.increment ??= { phi: 1, theta: 1 };
  const t = { ab: bAB(row) }, { a, b } = t.ab;
  a.out = [...a.out, t];
  b.in = [...b.in, t];
  step.edges.add(t);
  return t;
}

const bAddTree = (builder, row) => bLink(bOpenStep(builder), row);

function bDropTree(builder, t) {
  const { a, b } = t.ab;
  a.out = a.out.filter(x => x !== t);
  b.in = b.in.filter(x => x !== t);
  bOpenStep(builder).edges.delete(t);
}
