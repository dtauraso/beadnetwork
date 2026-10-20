const bKind = row => row.kind || 'angle';
const bFloat = row => bKind(row) === 'base';

function bDirs(st, tree) {
  const t = bDir(st, tree.ab.row);
  return t ? [{ t }] : [];
}

function bCopy(st, row) {
  const into = row.from.via && row.from.via.row;
  if (!into || bKind(into) !== 'angle') return { phi: 0, theta: 0 };
  const k = bS(st, row.from) / bS(st, into.from);
  return { phi: into.start.phi * k, theta: into.start.theta * k };
}

function bDir(st, row) {
  if (bKind(row) === 'base') return row.base || null;
  if (bKind(row) === 'angle') return row.start;
  const c = bCopy(st, row), s = bS(st, row.from);
  return { phi: bWrap(s, c.phi + row.end.phi), theta: bWrap(s, c.theta + row.end.theta), r: row.start.r };
}

function bReview(st, r) {
  const v = bView(st, r);
  if (!v || bFloat(r)) return;
  bSave(r);
  if (bKind(r) === 'angle') { r.start = v; return; }
  const c = bCopy(st, r), s = bS(st, r.from);
  r.start = { ...r.start, r: v.r };
  r.end = { phi: bWrap(s, v.phi - c.phi), theta: bWrap(s, v.theta - c.theta) };
}

function bSetKind(st, tree, kind) {
  const { row, a, b } = tree.ab, said = `${a.name}→${b.name} as ${kind}`, was = bKind(row);
  const keep = bKeep();
  if (kind === 'base' && b.via === tree.ab) bLetGo(st, b);
  bSave(row);
  row.kind = kind;
  if (kind === 'base') { bDone(); row.base = bView(st, row); return row.base ? '' : `${said}: ${a.name} or ${b.name} has no position to measure from`; }
  row.end ||= { phi: 0, theta: 0 };
  bLink(st, tree.ab);
  bReview(st, row);
  return bSettle(st, keep, [tree], said, was);
}

const bAt = r => `${r.from.name}→${r.to.name}`;

function bRowCells(st, tree, c, h, on) {
  const r = tree.ab.row, kind = h('select', {}, ['angle', 'change', 'base'].map(k => h('option', { value: k, textContent: k, selected: bKind(r) === k })));
  kind.onchange = () => on.kind(kind.value);
  const d = bView(st, r), now = (x, k) => [x, h('span', { className: 'now', textContent: d ? `now ${bN(+bF(d[k]))}` : '' })];
  if (bKind(r) === 'angle') return [['kind', kind], ['start φ', now(c(r.start, 'phi', tree), 'phi')], ['start θ', now(c(r.start, 'theta', tree), 'theta')], ['start r', now(c(r.start, 'r', tree), 'r')]];
  if (bKind(r) === 'change') {
    const k = bCopy(st, r);
    return [['kind', kind], ['start φ', `${bN(k.phi)} copy`], ['start θ', `${bN(k.theta)} copy`], ['start r', now(c(r.start, 'r', tree), 'r')],
      ['end φ', now(c(r.end, 'phi', tree, 'end'), 'phi')], ['end θ', now(c(r.end, 'theta', tree, 'end'), 'theta')]];
  }
  const ts = bDirs(st, tree).map(e => e.t);
  const set = h('button', { textContent: 'set base' });
  set.onclick = () => { r.base = bView(st, r); on.changed(); };
  const three = [['φ', 'phi'], ['θ', 'theta'], ['r', 'r']].map(([n, k]) => [n, ts.length ? ts.map(t => bF(t[k])).join(' / ') : '—']);
  return [['kind', kind], ...three, ['base', [set]]];
}

function bHow(st, o, r) {
  if (bKind(r) === 'angle') return `angle ${r.from.name}→${r.to.name}`;
  const ch = o.via && bKind(o.via.row) === 'change' ? o.via.row : null;
  if (bKind(r) === 'base' && ch) {
    const k = bCopy(st, ch);
    return `base; ${o.name} = ${ch.from.name} + (${ch.from.name}'s φ ${bN(k.phi)} θ ${bN(k.theta)} + ${o.name}'s own φ ${bN(ch.end.phi)} θ ${bN(ch.end.theta)})`;
  }
  return `${bKind(r)} ${r.from.name}→${r.to.name}`;
}

function bSecondBase(st, n, h, table) {
  const bases = n.out.filter(t => bFloat(t.ab.row)).flatMap(t => bDirs(st, t).map(e => ({ r: t.ab.row, ...e })));
  const seen = bases.length ? [...n.out.flatMap(t => bDirs(st, t).map(e => [t.ab.b, e.t, t.ab.row])),
    ...n.in.flatMap(t => bDirs(st, t).map(e => [t.ab.a, bBack(st, t.ab.row, e.t), t.ab.row]))] : [];
  return bases.flatMap(({ r, t: z }) => [
    h('h2', { textContent: `${n.name}'s second base — the ${bAt(r)} ${bKind(r)}, measured once at φ ${bF(z.phi)}, θ ${bF(z.theta)} and stored as 0` }),
    table(['toward', 'φ', 'θ', 'r', 'how'], seen.map(([o, t, x]) => h('tr', {}, [o.name, bN(t.phi - z.phi), bN(t.theta - z.theta), bN(t.r), bHow(st, o, x)]
      .map(v => h('td', { textContent: v })))))]);
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
    ...(flip ? [`${a}→${b}'s stored φ is past ${n / 2}, so φ = ${n} − φ = ${bF(d.phi)}, θ = θ + ${n / 2} = ${bF(d.theta)}`] : []),
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
  const ns = [...st.nodes].filter(n => n.out.length || n.in.length);
  const items = ns.map(n => {
    const at = n.known ? [['I_r', 'r'], ['I_φ', 'phi'], ['I_θ', 'theta']].map(([k, x]) => [k, bN(n.idx[x])]) : [['position', 'none']];
    const c = bNodePos(st, n);
    const attrs = [['s', `${bS(st, n)} (${12 * bS(st, n)} ticks a turn)`], ...at, ...(c ? [['C', `${['x', 'y', 'z'].map((k, i) => `${k} ${bN(+bF(c[i]))}`).join(' ')}`]] : []),
      ...n.in.map(t => [t.ab.a.name, bFromAngle(st, t, picked), t.ab.row]).map(([a, d, row]) => [`from ${a}`, d ? `φ ${bN(+bF(d.phi))} θ ${bN(+bF(d.theta))} r ${bN(+bF(d.r))}` :`${a} or ${n.name} has no position`, ...bDelta(d, row.delta, h), ...bAngleMath(st, row, d, h)])];
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
      h('ul', { className: 'attrs' }, [...three('row', e.t), ...three('placed', bView(st, r))]
        .map(([k, v]) => h('li', {}, [h('span', { className: 'k', textContent: k }), v])))]));
  });
  return [h('h2', { textContent: 'read only: each row, and the direction from its first node to its second node where they are placed' }),
    items.length ? h('ul', { className: 'abs' }, items) : h('div', { className: 'hint', textContent: 'no rows' })];
}
