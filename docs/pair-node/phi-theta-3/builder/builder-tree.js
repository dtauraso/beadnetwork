const bAB = row => ({ row, a: row.from, b: row.to });
const bMake = n => Object.assign(n, { out: [], in: [] });

function bAddTree(builder, row) {
  row.delta ??= { phi: 0, theta: 0 };
  row.by ??= { phi: 1, theta: 1 };
  const t = { ab: bAB(row) }, { a, b } = t.ab;
  a.out = [...a.out, t];
  b.in = [...b.in, t];
  builder.trees.add(t);
  return t;
}

function bDropTree(builder, t) {
  const { a, b } = t.ab;
  a.out = a.out.filter(x => x !== t);
  b.in = b.in.filter(x => x !== t);
  builder.trees.delete(t);
}
