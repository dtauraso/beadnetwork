function bPut(builder, row) {
  const n = row.to, f = row.from;
  const t = bDir(builder, row);
  if (!t) return false;
  bSetAt(builder, n, bAdd(bPlace(builder, f.idx), bVec(builder, bS(builder, f), t)));
  return true;
}

function bRemeasure(builder, n, keep = []) {
  [...n.in, ...n.out].filter(t => !keep.includes(t)).forEach(t => {
    const v = bMeasure(builder, t.ab.row);
    const dl = t.ab.row.delta;
    const s = bS(builder, t.ab.a);
    if (v) t.ab.row.start = { phi: bWrap(s, v.phi - dl.phi), theta: bWrap(s, v.theta - dl.theta), r: v.r };
  });
}

function bMove(builder, tree) {
  if (bPut(builder, tree.ab.row)) bRemeasure(builder, tree.ab.b, [tree]);
}

function bMoveAll(builder, trees) {
  trees.forEach(() => trees.forEach(t => bPut(builder, t.ab.row)));
  trees.forEach(t => bRemeasure(builder, t.ab.b, trees));
}

function bSetStart(builder, tree, k, value) {
  const row = tree.ab.row;
  row.start[k] = k === 'r' ? value : bWrap(bS(builder, row.from), value);
  bMove(builder, tree);
  return '';
}

function bSetS(builder, node, value) {
  const was = bS(builder, node), k = value / was;
  if (!bWhole(value) || value < 1) return `${node.name}'s s ${bN(value)}: not a whole number of at least 1, so it stays at ${was}`;
  node.s = value;
  node.out.forEach(t => {
    const r = t.ab.row;
    r.start = { ...r.start, phi: r.start.phi * k, theta: r.start.theta * k };
  });
  return '';
}
