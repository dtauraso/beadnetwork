function bSteps(h, steps) {
  const step = (i, kids, head = [], cls = '') => h('li', { className: cls }, [h('div', { className: 'dot', textContent: `${i + 1}` }),
    h('div', { className: 'body' }, [h('div', { className: 'step' }, [h('span', { textContent: `step ${i + 1}${kids.length ? '' : ' — empty'}` }), ...head]), ...kids])]);
  return h('ol', { className: 'seq' }, [...steps.map(({ body, head }, i) => step(i, [].concat(body), head)), step(steps.length, [], [], 'next')]);
}

function bStepNodes(builder, h, changed) {
  const ns = bShown(builder);
  const rows = ns.map(n => {
    const c = bNodePos(builder, n);
    const cells = [n.name, `${bS(builder, n)} (${12 * bS(builder, n)})`, ...['r', 'phi', 'theta'].map(x => bN(n.idx[x])),
      c ? ['x', 'y', 'z'].map((k, i) => `${k} ${bN(+bF(c[i]))}`).join(' ') : '—'];
    return h('tr', {}, cells.map((t, i) => h('td', { className: i ? '' : 't', textContent: t })));
  });
  const fold = h('div', { className: 'folds' }, [bFold(h, bNodesShown, 'nodes', rows.length, () => { bNodesShown = !bNodesShown; }, changed)]);
  if (!bNodesShown) return [fold];
  return [fold,
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
