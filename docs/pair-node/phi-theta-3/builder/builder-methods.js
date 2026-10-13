const bAdd = (p, q) => p.map((v, i) => v + q[i]);
const bSub = (p, q) => p.map((v, i) => v - q[i]);
const bLen = v => Math.hypot(...v);
const bCart = (R, ph, th) => [R * Math.sin(ph) * Math.cos(th), R * Math.cos(ph), R * Math.sin(ph) * Math.sin(th)];
const bF = n => (Math.abs(n) < 5e-4 ? 0 : n).toFixed(2);
const bPt = p => `(${bF(p[0])}, ${bF(p[1])}, ${bF(p[2])})`;
const B_O = [0, 0, 0];

function bRowVectors(st, item) {
  const ang = t => t * 2 * Math.PI / (12 * st.s);
  return st.rows.filter(r => r.to === item).map(r => ({
    from: r.from, row: r,
    d: bCart(r.start.r * st.rho, ang(r.start.phi), ang(r.start.theta)),
  }));
}

function bVectorStep(rows, item) {
  return {
    title: 'local: each row\'s start as a vector',
    lines: rows.length
      ? rows.map(({ from, row, d }) => `d_${from}${item} = polar_to_cart(${row.start.r}·ρ, ${row.start.phi}·2π/12s, ${row.start.theta}·2π/12s) = ${bPt(d)}`)
      : [`${item}'s table has no rows`],
    note: 'from the row\'s start ticks alone; a vector, not yet a place',
    ok: rows.length > 0, draw: [],
  };
}

function bAgainstSet(st, item, p, label) {
  const set = st.nodes.find(n => n.name === item);
  if (!set || !set.known) return null;
  const miss = bLen(bSub(p, set.pos));
  return `${label} − ${item} as set = ${bF(miss)}${miss < 1e-6 ? ' — lands on it' : ' — misses it'}`;
}

function bMethodAtoB(st, item) {
  const rows = bRowVectors(st, item);
  const steps = [bVectorStep(rows, item)];
  if (!rows.length) return steps;
  const pos = n => st.nodes.find(x => x.name === n);
  const placed = rows.map(r => ({ ...r, C: pos(r.from)?.known ? pos(r.from).pos : null }));
  const missing = placed.filter(r => !r.C).map(r => r.from);
  steps.push({
    title: 'the sender\'s position',
    lines: placed.map(r => r.C ? `C_${r.from} = ${bPt(r.C)}` : `C_${r.from}: missing — ${r.from} has no position`),
    note: 'each row\'s vector starts at its sender, so the sender\'s position is needed',
    ok: !missing.length,
    draw: placed.filter(r => r.C).map(r => ({ dot: r.C, label: `C_${r.from}` })),
  });
  const done = placed.filter(r => r.C).map(r => ({ ...r, b: bAdd(r.C, r.d) }));
  if (!done.length) return steps;
  const lines = done.map(r => `${item}(${r.from}) = C_${r.from} + d_${r.from}${item} = ${bPt(r.b)}`);
  done.slice(1).forEach(r => {
    const gap = bLen(bSub(r.b, done[0].b));
    lines.push(`|${item}(${r.from}) − ${item}(${done[0].from})| = ${bF(gap)}${gap < 1e-6 ? ' — one point' : ' — the rows disagree'}`);
  });
  missing.forEach(n => lines.push(`${item}(${n}): not placed — C_${n} is missing`));
  done.forEach(r => { const c = bAgainstSet(st, item, r.b, `${item}(${r.from})`); if (c) lines.push(c); });
  steps.push({
    title: `global: ${item} from each row`,
    lines, note: `O→n + n→${item}, once per row (the a→b card); nothing is added at O`, ok: true,
    draw: done.flatMap(r => [{ arrow: [r.C, r.b] }, { dot: r.b, label: `${item}(${r.from})` }]),
  });
  return steps;
}

function bMethodPair(st, item) {
  const rows = bRowVectors(st, item);
  const steps = [bVectorStep(rows, item)];
  let tip = B_O;
  const chain = rows.map(r => { const a = tip; tip = bAdd(tip, r.d); return { arrow: [a, tip] }; });
  const v = tip;
  steps.push({
    title: 'the rows added at O',
    lines: [`v = ${rows.map(r => `d_${r.from}${item}`).join(' + ') || '0'} = ${bPt(v)}`],
    note: 'every row\'s vector is laid tip to tail from O; no sender\'s position is used',
    ok: true, draw: [...chain, { dot: v, label: 'v' }],
  });
  const r = bLen(v);
  if (r < 1e-9 && rows.length) {
    steps.push({ title: 'v as an index', lines: ['|v| = 0 with rows present: acos(0/0) is undefined'],
      note: 'the card\'s "no rows: v = 0" covers only an empty table', ok: false, draw: [] });
    return steps;
  }
  const I = r < 1e-9 ? [0, 0, 0] : [r, Math.acos(Math.max(-1, Math.min(1, v[1] / r))), Math.atan2(v[2], v[0])];
  steps.push({
    title: 'v as an index',
    lines: [`I_r·c_r = |v| = ${bF(I[0])}, I_φ·c_φ = acos(v_y/|v|) = ${bF(I[1] * 180 / Math.PI)}°, I_θ·c_θ = atan2(v_z, v_x) = ${bF(I[2] * 180 / Math.PI)}°`],
    note: 'kept exact, so c_r, T_φ, T_θ do not change the result', ok: true, draw: [],
  });
  const Cb = bAdd(B_O, bCart(...I));
  const lines = [`C_${item} = O + polar_to_cart(I_r·c_r, I_φ·c_φ, I_θ·c_θ) = ${bPt(Cb)}`];
  const c = bAgainstSet(st, item, Cb, `C_${item}`);
  if (c) lines.push(c);
  steps.push({ title: `place ${item} from the index`, lines, note: 'the pair card\'s C_b', ok: true,
    draw: [{ arrow: [B_O, Cb] }, { square: Cb, label: `pair C_${item}` }] });
  return steps;
}

const B_METHODS = [
  { id: 'atob', name: 'a→b card: C_n + d_nb, one per row', color: '--local', run: bMethodAtoB },
  { id: 'pair', name: 'pair card: rows added at O', color: '--global', run: bMethodPair },
];
