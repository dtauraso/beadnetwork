function bDirs(st, tree) {
  return [{ t: bDir(st, tree.ab.row) }];
}

const bDir = (st, row) => row.start;

const bAt = r => `${r.from.name}→${r.to.name}`;

function bRowCells(st, tree, c, h) {
  const r = tree.ab.row, d = bView(st, r), now = (x, k) => [x, h('span', { className: 'now', textContent: d ? `view ${bN(+bF(d[k]))}` : '' })];
  return [['start φ', now(c(r.start, 'phi', tree), 'phi')], ['start θ', now(c(r.start, 'theta', tree), 'theta')], ['start r', now(c(r.start, 'r', tree), 'r')]];
}

function bSteps(h, steps) {
  const step = (i, kids, head = [], cls = '') => h('li', { className: cls }, [h('div', { className: 'dot', textContent: `${i + 1}` }),
    h('div', { className: 'body' }, [h('div', { className: 'step' }, [h('span', { textContent: `step ${i + 1}${kids.length ? '' : ' — empty'}` }), ...head]), ...kids])]);
  return h('ol', { className: 'seq' }, [...steps.map(({ body, head }, i) => step(i, [].concat(body), head)), step(steps.length, [], [], 'next')]);
}

function bAngleMath(st, row, d, h) {
  const P = bNodePos(st, row.from), T = bNodePos(st, row.to);
  if (!d || !P || !T) return [];
  const s = bS(st, row.from), n = 12 * s, v = bSub(T, P), L = bLen(v), a = row.from.name, b = row.to.name;
  const was = bDir(st, row), flip = !!was && was.phi > n / 2, xyz = p => `(${p.map(bF).join(', ')})`;
  const lines = [`v = C_${b} − C_${a} = ${xyz(T)} − ${xyz(P)} = ${xyz(v)}`, `|v| = √(v_x² + v_y² + v_z²) = ${bF(L)}`,
    `1 tick at ${a} = 2π / (12·s_${a}) = 2π / ${n}`,
    `φ = acos(v_y / |v|) / tick = acos(${bF(v[1])} / ${bF(L)}) / (2π/${n}) = ${bF(flip ? n - d.phi : d.phi)}`,
    `θ = atan2(v_z, v_x) / tick = atan2(${bF(v[2])}, ${bF(v[0])}) / (2π/${n}) = ${bF(flip ? bWrap(s, d.theta - n / 2) : d.theta)}, wrapped to 0 … ${n}`,
    ...(flip ? [`${a}→${b}'s start φ is past ${n / 2}, so φ = ${n} − φ = ${bF(d.phi)}, θ = θ + ${n / 2} = ${bF(d.theta)}`] : []),
    `r = |v| / ρ = ${bF(L)} / ${bF(st.rho)} = ${bF(d.r)}`];
  return [h('details', {}, [h('summary', { textContent: 'formulas' }), ...lines.map(t => h('code', { textContent: t }))])];
}

function bDelta(d, dl, h) {
  if (!d) return [];
  const f = x => bN(+bF(x));
  return [h('div', { className: 'sub', textContent: `Δφ ${f(dl.phi)} | Δθ ${f(dl.theta)}` }),
    h('div', { className: 'sub', textContent: `φ ${f(d.phi)} + Δφ ${f(dl.phi)} = ${f(d.phi + dl.phi)} | θ ${f(d.theta)} + Δθ ${f(dl.theta)} = ${f(d.theta + dl.theta)}` })];
}

function bStepNodes(st, h, picked) {
  const ns = bShown(st);
  const items = ns.map(n => {
    const at = n.known ? [['I_r', 'r'], ['I_φ', 'phi'], ['I_θ', 'theta']].map(([k, x]) => [k, bN(n.idx[x])]) : [['position', 'none']];
    const c = bNodePos(st, n);
    const attrs = [['s', `${bS(st, n)} (${12 * bS(st, n)} ticks a turn)`], ...at, ...(c ? [['C', `${['x', 'y', 'z'].map((k, i) => `${k} ${bN(+bF(c[i]))}`).join(' ')}`]] : []),
      ...n.in.map(t => [t.ab.a.name, bFromAngle(st, t, picked), t.ab.row]).map(([a, d, row]) => [`${a}→${n.name} view`, d ? `φ ${bN(+bF(d.phi))} θ ${bN(+bF(d.theta))} r ${bN(+bF(d.r))}` :`${a} or ${n.name} has no position`, ...bDelta(d, row.delta, h), ...bAngleMath(st, row, d, h)])];
    return h('li', {}, [h('div', { className: 'row' }, [h('span', { className: 't', textContent: n.name })]),
      h('ul', { className: 'attrs' }, attrs.map(([k, v, ...more]) => h('li', {}, [h('span', { className: 'k', textContent: k }), v, ...more])))]);
  });
  return [h('div', { className: 'part', textContent: 'nodes' }),
    items.length ? h('ul', { className: 'abs' }, items) : h('div', { className: 'hint', textContent: 'no nodes' })];
}

function bLands(st, h) {
  const three = (at, x) => [['φ', 'phi'], ['θ', 'theta'], ['r', 'r']].map(([n, k]) => [`${at} ${n}`, x ? bF(x[k]) : '—']);
  const items = [...st.trees].flatMap(t => {
    const dirs = bDirs(st, t), r = t.ab.row;
    return (dirs.length ? dirs : [{}]).map(e => h('li', {}, [h('div', { className: 'row' }, [h('span', { className: 't', textContent: bAt(r) })]),
      h('ul', { className: 'attrs' }, [...three('start', e.t), ...three('view', bView(st, r))]
        .map(([k, v]) => h('li', {}, [h('span', { className: 'k', textContent: k }), v])))]));
  });
  return [h('h2', { textContent: 'read only: each row\'s start, and its view — the direction from its first node to its second node where they are placed' }),
    items.length ? h('ul', { className: 'abs' }, items) : h('div', { className: 'hint', textContent: 'no rows' })];
}
