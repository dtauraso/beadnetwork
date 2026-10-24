function bSteps(h, steps) {
  const step = (i, kids, head = [], cls = '') => h('li', { className: cls }, [h('div', { className: 'dot', textContent: `${i + 1}` }),
    h('div', { className: 'body' }, [h('div', { className: 'step' }, [h('span', { textContent: `step ${i + 1}${kids.length ? '' : ' — empty'}` }), ...head]), ...kids])]);
  return h('ol', { className: 'seq' }, [...steps.map(({ body, head }, i) => step(i, [].concat(body), head)), step(steps.length, [], [], 'next')]);
}

function bAngleLines(builder, row, d) {
  const P = bNodePos(builder, row.from), T = bNodePos(builder, row.to);
  const s = bS(builder, row.from), n = 12 * s, v = bSub(T, P), L = bLen(v), a = row.from.name, b = row.to.name;
  const was = bDir(builder, row), flip = !!was && was.phi > n / 2, xyz = p => `(${p.map(bF).join(', ')})`;
  return [`v = C_${b} − C_${a} = ${xyz(T)} − ${xyz(P)} = ${xyz(v)}`, `|v| = √(v_x² + v_y² + v_z²) = ${bF(L)}`,
    `1 tick at ${a} = 2π / (12·s_${a}) = 2π / ${n}`,
    `φ = acos(v_y / |v|) / tick = acos(${bF(v[1])} / ${bF(L)}) / (2π/${n}) = ${bF(flip ? n - d.phi : d.phi)}`,
    `θ = atan2(v_z, v_x) / tick = atan2(${bF(v[2])}, ${bF(v[0])}) / (2π/${n}) = ${bF(flip ? bWrap(s, d.theta - n / 2) : d.theta)}, wrapped to 0 … ${n}`,
    ...(flip ? [`${a}→${b}'s start φ is past ${n / 2}, so φ = ${n} − φ = ${bF(d.phi)}, θ = θ + ${n / 2} = ${bF(d.theta)}`] : []),
    `r = |v| / ρ = ${bF(L)} / ${bF(builder.rho)} = ${bF(d.r)}`];
}

function bStepNodes(builder, h) {
  const ns = bShown(builder);
  const items = ns.map(n => {
    const at = [['I_r', 'r'], ['I_φ', 'phi'], ['I_θ', 'theta']].map(([k, x]) => [k, bN(n.idx[x])]);
    const c = bNodePos(builder, n);
    const attrs = [['s', `${bS(builder, n)} (${12 * bS(builder, n)} ticks a turn)`], ...at, ...(c ? [['C', `${['x', 'y', 'z'].map((k, i) => `${k} ${bN(+bF(c[i]))}`).join(' ')}`]] : [])];
    return h('li', {}, [h('div', { className: 'row' }, [h('span', { className: 't', textContent: n.name })]),
      h('ul', { className: 'attrs' }, attrs.map(([k, v]) => h('li', {}, [h('span', { className: 'k', textContent: k }), v])))]);
  });
  return [h('div', { className: 'part', textContent: 'nodes' }),
    items.length ? h('ul', { className: 'abs' }, items) : h('div', { className: 'hint', textContent: 'no nodes' })];
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
