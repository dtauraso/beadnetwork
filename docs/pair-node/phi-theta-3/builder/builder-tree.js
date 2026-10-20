const bAB = row => ({ row, a: row.from, b: row.to });
const bMake = n => Object.assign(n, { out: [], in: [] });
const B_GONE = ['turn', 'held', 'went', 'step', 'mine', 'automatic', 'ref', 'pairs'];

function bAddTree(st, row) {
  row.delta ??= { phi: 0, theta: 0 };
  row.by ??= { phi: 1, theta: 1 };
  const t = { ab: bAB(row) }, { a, b } = t.ab;
  a.out = [...a.out, t];
  b.in = [...b.in, t];
  st.trees.add(t);
  return t;
}

function bDropTree(st, t) {
  const { a, b } = t.ab;
  if (b.via === t.ab) bLetGo(st, b);
  a.out = a.out.filter(x => x !== t);
  b.in = b.in.filter(x => x !== t);
  st.trees.delete(t);
}
