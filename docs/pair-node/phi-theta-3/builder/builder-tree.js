const bAB = row => ({ row, a: row.from, b: row.to });
const bAC = row => ({ row, a: row.from, c: row.to });
const bMake = n => Object.assign(n, { out: [], in: [] });
const B_PAIR = ['base', 'ref'];
const B_GONE = ['turn', 'held', 'went', 'step', 'mine', 'automatic'];

function bSeat(st, p) {
  const s = p.ab.row.start, v = bView(st, p.ac.row);
  bSave(p);
  p.base = s && v && { phi: s.phi, theta: s.theta, r: v.r };
  p.ref = s && { phi: s.phi, theta: s.theta };
}

function bJoin(st, t, o) {
  const p = { ab: t.ab, ac: bAC(o.ab.row) }, q = { ab: o.ab, ac: bAC(t.ab.row) };
  bSeat(st, p);
  bSeat(st, q);
  const r = { tree: t, pair: p }, s = { tree: o, pair: q, back: r };
  r.back = s;
  t.roots = [...t.roots, r];
  o.roots = [...o.roots, s];
}

function bAddTree(st, row) {
  row.delta ??= { phi: 0, theta: 0 };
  const t = { ab: bAB(row), roots: [] }, { a, b } = t.ab;
  a.out.forEach(o => bJoin(st, t, o));
  a.out = [...a.out, t];
  b.in = [...b.in, t];
  st.trees.add(t);
  return t;
}

function bDropTree(st, t) {
  const { a, b } = t.ab;
  if (b.via === t.ab) bLetGo(st, b);
  t.roots.forEach(({ back }) => { back.tree.roots = back.tree.roots.filter(x => x !== back); });
  a.out = a.out.filter(x => x !== t);
  b.in = b.in.filter(x => x !== t);
  st.trees.delete(t);
}
