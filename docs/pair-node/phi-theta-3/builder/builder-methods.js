const bAdd = (p, q) => p.map((v, i) => v + q[i]);
const bSub = (p, q) => p.map((v, i) => v - q[i]);
const bLen = v => Math.hypot(...v);
const bCart = (R, ph, th) => [R * Math.sin(ph) * Math.cos(th), R * Math.cos(ph), R * Math.sin(ph) * Math.sin(th)];
const bF = n => (Math.abs(n) < 5e-4 ? 0 : n).toFixed(2);
const bN = n => (Number.isInteger(n) ? `${n}` : bF(n));
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

const bS = (st, n) => (n && n.s) || st.s;
const bWrap = (s, t) => ((t % (12 * s)) + 12 * s) % (12 * s);
const bBack = (st, row) => {
  const sa = bS(st, row.from), sb = bS(st, row.to), a = bDir(st, row);
  return { phi: bWrap(sb, a.phi * sb / sa + 6 * sb), theta: bWrap(sb, a.theta * sb / sa + 6 * sb), r: a.r };
};
const bWhole = x => Math.abs(x - Math.round(x)) < 1e-6;

const bIdx = I => `(I_r ${bF(I.r)}, I_φ ${bF(I.phi)}, I_θ ${bF(I.theta)})`;
const bPlaceText = I => `O + polar_to_cart(${bF(I.r)}·c_r, ${bF(I.phi)}·c_φ, ${bF(I.theta)}·c_θ)`;

const bNodePos = (st, n) => (n && n.known ? bPlace(st, n.idx) : null);

function bVec(st, s, t) {
  const ang = x => x * 2 * Math.PI / (12 * s);
  return bCart(t.r * st.rho, ang(t.phi), ang(t.theta));
}

function bRowVectors(st, item) {
  return ((bNode(st, item) || {}).rowsIn || []).filter(r => bDir(st, r)).map(r => {
    const t = bDir(st, r);
    return { from: r.from, row: r, t, d: bVec(st, bS(st, r.from), t) };
  });
}

function bVectorStep(rows, item) {
  return {
    title: 'local: each row\'s start as a vector',
    lines: rows.length
      ? rows.map(({ from: { name: from }, row, t, d }) => `d_${from}${item} (${bKind(row)}) = polar_to_cart(${bN(t.r)}·ρ, ${bN(t.phi)}·2π/12s_${from}, ${bN(t.theta)}·2π/12s_${from}) = ${bPt(d)}`)
      : [`${item}'s table has no rows`],
    note: 'from the row\'s start ticks alone; a vector, not yet a place',
    ok: rows.length > 0, draw: [],
  };
}

function bResult(st, item, p, label) {
  const lines = [`${label} as an index: ${bIdx(bIndexOf(st, p))}`];
  const set = bNodePos(st, item);
  item = item.name;
  if (set) {
    const miss = bLen(bSub(p, set));
    lines.push(`|${label} − ${item} as set| = ${bF(miss)}${miss < 1e-6 ? ' — lands on it' : ' — misses it'}`);
  }
  return lines;
}

function bSenders(st, rows) {
  const placed = rows.map(r => ({ ...r, from: r.from.name, I: r.from.known ? r.from.idx : null, C: bNodePos(st, r.from) }));
  const missing = placed.filter(r => !r.C).map(r => r.from);
  const step = {
    title: 'the sender\'s position, from its index',
    lines: placed.map(r => r.C
      ? `C_${r.from} = ${bPlaceText(r.I)} = ${bPt(r.C)}`
      : `C_${r.from}: missing — ${r.from} has no index`),
    note: 'each row\'s vector starts at its sender, so the sender\'s position is needed',
    ok: !missing.length,
    draw: placed.filter(r => r.C).map(r => ({ dot: r.C, label: `C_${r.from}` })),
  };
  return { step, missing, done: placed.filter(r => r.C).map(r => ({ ...r, b: bAdd(r.C, r.d) })) };
}

function bMethodAtoB(st, item) {
  const rows = bRowVectors(st, item);
  const steps = [bVectorStep(rows, item)];
  if (!rows.length) return steps;
  const { step, missing, done } = bSenders(st, rows);
  steps.push(step);
  if (!done.length) return steps;
  const lines = done.map(r => `${item}(${r.from}) = C_${r.from} + d_${r.from}${item} = ${bPt(r.b)}`);
  done.slice(1).forEach(r => {
    const gap = bLen(bSub(r.b, done[0].b));
    lines.push(`|${item}(${r.from}) − ${item}(${done[0].from})| = ${bF(gap)}${gap < 1e-6 ? ' — one point' : ' — the rows disagree'}`);
  });
  missing.forEach(n => lines.push(`${item}(${n}): not placed — C_${n} is missing`));
  const node = bNode(st, item);
  done.forEach(r => lines.push(...bResult(st, node, r.b, `${item}(${r.from})`)));
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
  if (!rows.length) return steps;
  const { step, missing, done } = bSenders(st, rows);
  steps.push(step);
  if (!done.length) return steps;
  steps.push({
    title: `global: C_${item} = C_n + d_n${item}, per pair`,
    lines: [...done.map(r => `pair ${r.from}→${item}: C_${item} = C_${r.from} + d_${r.from}${item} = ${bPt(r.b)}`),
      ...missing.map(n => `pair ${n}→${item}: not placed — C_${n} is missing`)],
    note: `in ${item}, with C_n from the edge n→${item}`, ok: true,
    draw: done.flatMap(r => [{ arrow: [r.C, r.b] }]),
  });
  const indexed = done.map(r => ({ ...r, I: bIndexOf(st, r.b) }));
  steps.push({
    title: `convert: C_${item} − O as an index`,
    lines: ['I_r = |C_b − O| / c_r,  I_φ = acos((C_b − O)_y / |C_b − O|) / c_φ,  I_θ = atan2((C_b − O)_z, (C_b − O)_x) / c_θ',
      ...indexed.map(r => `pair ${r.from}→${item}: I_${item} = ${bIdx(r.I)}`)],
    note: 'kept exact — not rounded to whole index steps; C_b = O gives I_b = 0', ok: true, draw: [],
  });
  const lines = [], draw = [], node = bNode(st, item);
  indexed.forEach(r => {
    const Cb = bPlace(st, r.I), back = bLen(bSub(Cb, r.b));
    lines.push(`pair ${r.from}→${item}: C_${item} = ${bPlaceText(r.I)} = ${bPt(Cb)}`,
      `|placed − C_${r.from} + d_${r.from}${item}| = ${bF(back)}${back < 1e-6 ? ' — the index holds the same point' : ' — the index lost the point'}`,
      ...bResult(st, node, Cb, `C_${item}`).slice(1));
    draw.push({ arrow: [B_O, Cb] }, { square: Cb, label: `pair ${r.from}: C_${item}` });
  });
  steps.push({ title: `place ${item} from the index`, lines, note: 'the pair card\'s C_b', ok: true, draw });
  return steps;
}

const B_METHODS = [
  { id: 'atob', name: 'a→b card: C_n + d_nb, one per row', color: '--local', run: bMethodAtoB },
  { id: 'pair', name: 'pair card: C_b = C_a + d_ab, then index and place', color: '--global', run: bMethodPair },
];
