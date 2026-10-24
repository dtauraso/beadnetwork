function bEdgeTable(builder, h, c, mathOpen, err, changed) {
  const trees = [...bOpenStep(builder).edges], picked = builder.picked, f = x => bN(+bF(x));
  picked.splice(0, picked.length, ...picked.filter(t => trees.includes(t)));
  if (!trees.length) return [h('div', { className: 'hint', textContent: 'no edges' })];
  const step = (k, sign, text) => {
    const b = h('button', { textContent: text, disabled: !picked.length });
    b.onclick = () => { picked.forEach(t => { t.ab.row.delta[k] += sign * t.ab.row.increment[k]; bMove(builder, t); }); changed(); };
    return b;
  };
  const cols = 14;
  const rows = trees.flatMap(t => {
    const r = t.ab.row, ab = bAt(r), d = bMeasure(builder, r), e = bDir(builder, r);
    const pick = h('input', { type: 'checkbox', checked: picked.includes(t), title: 'picked: the Δ buttons above step it' });
    pick.onchange = () => { if (pick.checked) picked.push(t); else picked.splice(picked.indexOf(t), 1); changed(); };
    const inc = k => {
      const amount = h('input', { type: 'number', value: r.increment[k], step: 1 });
      amount.onchange = () => { const v = +amount.value; if (bWhole(v)) r.increment[k] = Math.round(v); changed(); };
      return amount;
    };
    const math = h('button', { textContent: mathOpen.has(t) ? 'hide' : 'math', disabled: !d, title: d ? 'start + Δ, measured from the nodes' : `${r.from.name} or ${r.to.name} has no position` });
    math.onclick = () => { if (mathOpen.has(t)) mathOpen.delete(t); else mathOpen.add(t); changed(); };
    const del = h('button', { textContent: '×', title: `remove ${ab}` });
    del.onclick = () => { bDropTree(builder, t); changed(); };
    const td = kids => h('td', {}, [].concat(kids));
    const out = [h('tr', {}, [td(h('span', { className: 't', textContent: ab })), td(pick),
      ...['phi', 'theta', 'r'].map(k => td(c(r.start, k, t))), td(f(r.delta.phi)), td(f(r.delta.theta)), td(inc('phi')), td(inc('theta')),
      td(f(e.phi)), td(f(e.theta)), td(f(e.r)), td(math), td(del)])];
    if (err && err.at.split(':')[0] === ab) out.push(h('tr', {}, [h('td', { className: 'missing', colSpan: cols, textContent: err.text })]));
    if (mathOpen.has(t) && d) out.push(h('tr', { className: 'math' }, [h('td', { colSpan: cols }, bAngleLines(builder, r, d).map(l => h('code', { textContent: l })))]));
    return out;
  });
  const th = (text, span = 1) => h('th', { textContent: text, colSpan: span });
  return [h('div', { className: 'row' }, ['picked edges, each by its own increment: Δφ', step('phi', -1, '−'), step('phi', 1, '+'), ' Δθ', step('theta', -1, '−'), step('theta', 1, '+')]),
    h('div', { className: 'wide' }, [h('table', { className: 'edges' }, [h('thead', {}, [
      h('tr', {}, [th('edge'), th('pick'), th('start', 3), th('Δ', 2), th('increment', 2), th('start + Δ', 3), th(''), th('')]),
      h('tr', {}, [th(''), th(''), ...['φ', 'θ', 'r', 'φ', 'θ', 'φ', 'θ', 'φ', 'θ', 'r'].map(x => th(x)), th(''), th('')])]),
    h('tbody', {}, rows)])])];
}
