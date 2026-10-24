function bSteps(h, steps) {
  const step = (i, kids, head = [], cls = '') => h('li', { className: cls }, [h('div', { className: 'dot', textContent: `${i + 1}` }),
    h('div', { className: 'body' }, [h('div', { className: 'step' }, [h('span', { textContent: `step ${i + 1}${kids.length ? '' : ' — empty'}` }), ...head]), ...kids])]);
  return h('ol', { className: 'seq' }, [...steps.map(({ body, head }, i) => step(i, [].concat(body), head)), step(steps.length, [], [], 'next')]);
}

function bAngleMath(builder, row, d, h) {
  const P = bNodePos(builder, row.from), T = bNodePos(builder, row.to);
  if (!d || !P || !T) return [];
  const s = bS(builder, row.from), n = 12 * s, v = bSub(T, P), L = bLen(v), a = row.from.name, b = row.to.name;
  const was = bDir(builder, row), flip = !!was && was.phi > n / 2, xyz = p => `(${p.map(bF).join(', ')})`;
  const lines = [`v = C_${b} − C_${a} = ${xyz(T)} − ${xyz(P)} = ${xyz(v)}`, `|v| = √(v_x² + v_y² + v_z²) = ${bF(L)}`,
    `1 tick at ${a} = 2π / (12·s_${a}) = 2π / ${n}`,
    `φ = acos(v_y / |v|) / tick = acos(${bF(v[1])} / ${bF(L)}) / (2π/${n}) = ${bF(flip ? n - d.phi : d.phi)}`,
    `θ = atan2(v_z, v_x) / tick = atan2(${bF(v[2])}, ${bF(v[0])}) / (2π/${n}) = ${bF(flip ? bWrap(s, d.theta - n / 2) : d.theta)}, wrapped to 0 … ${n}`,
    ...(flip ? [`${a}→${b}'s start φ is past ${n / 2}, so φ = ${n} − φ = ${bF(d.phi)}, θ = θ + ${n / 2} = ${bF(d.theta)}`] : []),
    `r = |v| / ρ = ${bF(L)} / ${bF(builder.rho)} = ${bF(d.r)}`];
  return [h('details', {}, [h('summary', { textContent: 'start + Δ, measured from the nodes' }), ...lines.map(t => h('code', { textContent: t }))])];
}

function bDelta(row, h) {
  const f = x => bN(+bF(x)), s = row.start, dl = row.delta;
  return [h('div', { className: 'sub', textContent: `Δφ ${f(dl.phi)} | Δθ ${f(dl.theta)}` }),
    h('div', { className: 'sub', textContent: `start φ ${f(s.phi)} + Δφ ${f(dl.phi)} = ${f(s.phi + dl.phi)} | start θ ${f(s.theta)} + Δθ ${f(dl.theta)} = ${f(s.theta + dl.theta)}` })];
}

function bStepNodes(builder, h) {
  const ns = bShown(builder);
  const rows = ns.flatMap(n => {
    const c = bNodePos(builder, n);
    const cells = [n.name, `${bS(builder, n)} (${12 * bS(builder, n)})`, ...['r', 'phi', 'theta'].map(x => bN(n.idx[x])),
      c ? ['x', 'y', 'z'].map((k, i) => `${k} ${bN(+bF(c[i]))}`).join(' ') : '—'];
    const ins = n.in.map(t => [t.ab.a.name, bMeasure(builder, t.ab.row), t.ab.row]).map(([a, d, row]) => h('tr', { className: 'in' }, [
      h('td', { className: 'k', textContent: `${a}→${n.name} start` }),
      h('td', { colSpan: 5 }, [d ? `φ ${bN(+bF(row.start.phi))} θ ${bN(+bF(row.start.theta))} r ${bN(+bF(row.start.r))}` : `${a} or ${n.name} has no position`, ...bDelta(row, h), ...bAngleMath(builder, row, d, h)])]));
    return [h('tr', {}, cells.map((t, i) => h('td', { className: i ? '' : 't', textContent: t }))), ...ins];
  });
  return [h('div', { className: 'part', textContent: 'nodes' }),
    rows.length ? h('table', { className: 'cells' }, [h('thead', {}, [h('tr', {}, ['node', 's (ticks a turn)', 'I_r', 'I_φ', 'I_θ', 'C'].map(t => h('th', { textContent: t })))]), h('tbody', {}, rows)])
      : h('div', { className: 'hint', textContent: 'no nodes' })];
}

function bLands(builder, h) {
  const three = (at, x) => [['φ', 'phi'], ['θ', 'theta'], ['r', 'r']].map(([n, k]) => [`${at} ${n}`, x ? bF(x[k]) : '—']);
  const items = [...bOpenStep(builder).edges].flatMap(t => {
    const dirs = bDirs(builder, t), r = t.ab.row;
    return (dirs.length ? dirs : [{}]).map(e => h('li', {}, [h('div', { className: 'row' }, [h('span', { className: 't', textContent: bAt(r) })]),
      h('ul', { className: 'attrs' }, [...three('start', r.start), ...three('start + Δ', e.t)]
        .map(([k, v]) => h('li', {}, [h('span', { className: 'k', textContent: k }), v])))]));
  });
  return [h('h2', { textContent: 'read only: each row\'s start, and start + Δ — the direction from its first node to its second node' }),
    items.length ? h('ul', { className: 'abs' }, items) : h('div', { className: 'hint', textContent: 'no rows' })];
}

function bGlobalLines(builder) {
  return bShown(builder).map(n => {
    const p = bNodePos(builder, n);
    const r = bLen(p), phi = r < 1e-9 ? 0 : Math.acos(Math.max(-1, Math.min(1, p[1] / r)));
    let th = Math.atan2(p[2], p[0]);
    if (th < -1e-9) th += 2 * Math.PI;
    return `${n.name} in O's frame: r ${bF(r)}, φ ${bF(phi)} rad, θ ${bF(Math.max(0, th))} rad`;
  });
}
