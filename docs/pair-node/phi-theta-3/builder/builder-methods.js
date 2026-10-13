const bAdd = (p, q) => p.map((v, i) => v + q[i]);
const bSub = (p, q) => p.map((v, i) => v - q[i]);
const bLen = v => Math.hypot(...v);
const bCart = (R, ph, th) => [R * Math.sin(ph) * Math.cos(th), R * Math.cos(ph), R * Math.sin(ph) * Math.sin(th)];
const bF = n => (Math.abs(n) < 5e-4 ? 0 : n).toFixed(2);
const bPt = p => `(${bF(p[0])}, ${bF(p[1])}, ${bF(p[2])})`;
const B_O = [0, 0, 0];

const bC = st => ({ r: st.cr, phi: 2 * Math.PI / st.Tphi, theta: 2 * Math.PI / st.Ttheta });

function bPlace(st, I) {
  const c = bC(st);
  return bAdd(B_O, bCart(I.r * c.r, I.phi * c.phi, I.theta * c.theta));
}

function bIndexOf(st, p) {
  const c = bC(st), w = bSub(p, B_O), r = bLen(w);
  if (r < 1e-9) return { r: 0, phi: 0, theta: 0 };
  return { r: r / c.r, phi: Math.acos(Math.max(-1, Math.min(1, w[1] / r))) / c.phi, theta: Math.atan2(w[2], w[0]) / c.theta };
}

const bIdx = I => `(I_r ${bF(I.r)}, I_φ ${bF(I.phi)}, I_θ ${bF(I.theta)})`;
const bPlaceText = I => `O + polar_to_cart(${bF(I.r)}·c_r, ${bF(I.phi)}·c_φ, ${bF(I.theta)}·c_θ)`;

function bNodePos(st, name) {
  const n = st.nodes.find(x => x.name === name);
  return n && n.known ? bPlace(st, n.idx) : null;
}

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

function bResult(st, item, p, label) {
  const lines = [`${label} as an index: ${bIdx(bIndexOf(st, p))}`];
  const set = bNodePos(st, item);
  if (set) {
    const miss = bLen(bSub(p, set));
    lines.push(`|${label} − ${item} as set| = ${bF(miss)}${miss < 1e-6 ? ' — lands on it' : ' — misses it'}`);
  }
  return lines;
}

function bMethodAtoB(st, item) {
  const rows = bRowVectors(st, item);
  const steps = [bVectorStep(rows, item)];
  if (!rows.length) return steps;
  const placed = rows.map(r => {
    const n = st.nodes.find(x => x.name === r.from);
    return { ...r, I: n && n.known ? n.idx : null, C: bNodePos(st, r.from) };
  });
  const missing = placed.filter(r => !r.C).map(r => r.from);
  steps.push({
    title: 'the sender\'s position, from its index',
    lines: placed.map(r => r.C
      ? `C_${r.from} = ${bPlaceText(r.I)} = ${bPt(r.C)}`
      : `C_${r.from}: missing — ${r.from} has no index`),
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
  done.forEach(r => lines.push(...bResult(st, item, r.b, `${item}(${r.from})`)));
  steps.push({
    title: `global: ${item} from each row`,
    lines, note: `O→n + n→${item}, once per row (the a→b card); nothing is added at O; the index is exact, not rounded`, ok: true,
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
  if (bLen(v) < 1e-9 && rows.length) {
    steps.push({ title: 'v as an index', lines: ['|v| = 0 with rows present: acos(0/0) is undefined'],
      note: 'the card\'s "no rows: v = 0" covers only an empty table', ok: false, draw: [] });
    return steps;
  }
  const I = bIndexOf(st, bAdd(B_O, v));
  steps.push({
    title: 'v as an index',
    lines: ['I_r = |v| / c_r,  I_φ = acos(v_y / |v|) / c_φ,  I_θ = atan2(v_z, v_x) / c_θ', `I_${item} = ${bIdx(I)}`],
    note: 'kept exact — not rounded to whole index steps', ok: true, draw: [],
  });
  const Cb = bPlace(st, I);
  steps.push({
    title: `place ${item} from the index`,
    lines: [`C_${item} = ${bPlaceText(I)} = ${bPt(Cb)}`, ...bResult(st, item, Cb, `C_${item}`).slice(1)],
    note: 'the pair card\'s C_b', ok: true,
    draw: [{ arrow: [B_O, Cb] }, { square: Cb, label: `pair C_${item}` }],
  });
  return steps;
}

const B_METHODS = [
  { id: 'atob', name: 'a→b card: C_n + d_nb, one per row', color: '--local', run: bMethodAtoB },
  { id: 'pair', name: 'pair card: rows added at O', color: '--global', run: bMethodPair },
];
