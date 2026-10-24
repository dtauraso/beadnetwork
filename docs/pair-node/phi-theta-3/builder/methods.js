function bRowVectors(builder, node) {
  return node.in.map(e => { const t = e.dir(builder); return { from: e.from, t, d: builder.vec(e.from.scale(builder), t) }; });
}

function bVectorStep(rows, item) {
  return {
    title: 'local: each row\'s start as a vector',
    lines: rows.length
      ? rows.map(({ from: { name: from }, t, d }) => `d_${from}${item} = polar_to_cart(${bN(t.r)}·ρ, ${bN(t.phi)}·2π/12s_${from}, ${bN(t.theta)}·2π/12s_${from}) = ${bPt(d)}`)
      : [`${item}'s table has no rows`],
    note: 'from the row\'s start ticks alone; a vector, not yet a place',
    ok: rows.length > 0, draw: [],
  };
}

function bResult(builder, item, p, label) {
  const miss = bLen(bSub(p, item.place(builder)));
  return [`${label} as an index: ${bIdx(builder.indexOf(p))}`,
    `|${label} − ${item.name} as set| = ${bF(miss)}${miss < 1e-6 ? ' — lands on it' : ' — misses it'}`];
}

function bFromNodes(builder, rows) {
  const placed = rows.map(r => ({ ...r, from: r.from.name, I: r.from.index(builder), C: r.from.place(builder) }));
  const step = {
    title: 'the from node\'s position, from its index',
    lines: placed.map(r => `C_${r.from} = ${bPlaceText(r.I)} = ${bPt(r.C)}`),
    note: 'each row\'s vector starts at its from node, so the from node\'s position is needed',
    ok: true,
    draw: placed.map(r => ({ dot: r.C, label: `C_${r.from}` })),
  };
  return { step, done: placed.map(r => ({ ...r, b: bAdd(r.C, r.d) })) };
}

function bMethodAtoB(builder, node) {
  const item = node.name, rows = bRowVectors(builder, node);
  const steps = [bVectorStep(rows, item)];
  if (!rows.length) return steps;
  const { step, done } = bFromNodes(builder, rows);
  steps.push(step);
  const lines = done.map(r => `${item}(${r.from}) = C_${r.from} + d_${r.from}${item} = ${bPt(r.b)}`);
  done.slice(1).forEach(r => {
    const gap = bLen(bSub(r.b, done[0].b));
    lines.push(`|${item}(${r.from}) − ${item}(${done[0].from})| = ${bF(gap)}${gap < 1e-6 ? ' — one point' : ' — the rows disagree'}`);
  });
  done.forEach(r => lines.push(...bResult(builder, node, r.b, `${item}(${r.from})`)));
  steps.push({
    title: `global: ${item} from each row`,
    lines, note: `O→n + n→${item}, once per row (the a→b card); nothing is added at O; the index is exact, not rounded`, ok: true,
    draw: done.flatMap(r => [{ arrow: [r.C, r.b] }, { dot: r.b, label: `${item}(${r.from})` }]),
  });
  return steps;
}

function bMethodPair(builder, node) {
  const item = node.name, rows = bRowVectors(builder, node);
  const steps = [bVectorStep(rows, item)];
  if (!rows.length) return steps;
  const { step, done } = bFromNodes(builder, rows);
  steps.push(step);
  steps.push({
    title: `global: C_${item} = C_n + d_n${item}, per pair`,
    lines: done.map(r => `pair ${r.from}→${item}: C_${item} = C_${r.from} + d_${r.from}${item} = ${bPt(r.b)}`),
    note: `in ${item}, with C_n from the edge n→${item}`, ok: true,
    draw: done.flatMap(r => [{ arrow: [r.C, r.b] }]),
  });
  const indexed = done.map(r => ({ ...r, I: builder.indexOf(r.b) }));
  steps.push({
    title: `convert: C_${item} − O as an index`,
    lines: ['I_r = |C_b − O| / c_r,  I_φ = acos((C_b − O)_y / |C_b − O|) / c_φ,  I_θ = atan2((C_b − O)_z, (C_b − O)_x) / c_θ',
      ...indexed.map(r => `pair ${r.from}→${item}: I_${item} = ${bIdx(r.I)}`)],
    note: 'kept exact — not rounded to whole index steps; C_b = O gives I_b = 0', ok: true, draw: [],
  });
  const lines = [], draw = [];
  indexed.forEach(r => {
    const Cb = builder.place(r.I), back = bLen(bSub(Cb, r.b));
    lines.push(`pair ${r.from}→${item}: C_${item} = ${bPlaceText(r.I)} = ${bPt(Cb)}`,
      `|placed − C_${r.from} + d_${r.from}${item}| = ${bF(back)}${back < 1e-6 ? ' — the index holds the same point' : ' — the index lost the point'}`,
      ...bResult(builder, node, Cb, `C_${item}`).slice(1));
    draw.push({ arrow: [B_O, Cb] }, { square: Cb, label: `pair ${r.from}: C_${item}` });
  });
  steps.push({ title: `place ${item} from the index`, lines, note: 'the pair card\'s C_b', ok: true, draw });
  return steps;
}

const B_METHODS = [
  { id: 'atob', name: 'a→b card: C_n + d_nb, one per row', color: '--local', run: bMethodAtoB },
  { id: 'pair', name: 'pair card: C_b = C_a + d_ab, then index and place', color: '--global', run: bMethodPair },
];
