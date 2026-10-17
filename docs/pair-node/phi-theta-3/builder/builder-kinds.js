const bKind = row => row.kind || 'angle';
const bFloat = row => bKind(row) === 'base' || bKind(row) === 'copy';
const bAuto = p => p.automatic !== false;
function bTurned(st, p) {
  const s = p.ab.row.start;
  return s && p.ref ? { sib: p.ab.row, phi: s.phi - p.ref.phi, theta: s.theta - p.ref.theta } : null;
}

function bChange(st, p) {
  const d = bTurned(st, p);
  if (d && p.mine) return { sib: d.sib, phi: p.went.phi, theta: p.went.theta };
  return d && !bAuto(p) ? { sib: d.sib, phi: p.held.phi, theta: p.held.theta } : d;
}

function bStep(st, p, was, now) {
  if (!p.mine) return now;
  const sg = k => Math.sign(now[k] - was[k]) * p.step[k];
  bSave(p);
  p.went = { phi: p.went.phi + sg('phi'), theta: p.went.theta + sg('theta') };
  return { phi: was.phi + sg('phi'), theta: was.theta + sg('theta'), r: now.r };
}

function bSetMine(st, p, on) {
  const d = bTurned(st, p), c = bChange(st, p);
  if (!d || !c) return;
  bSave(p);
  if (on) { p.went = { phi: c.phi, theta: c.theta }; p.step ||= { phi: 0, theta: 0 }; }
  else p.ref = { phi: d.sib.start.phi - p.went.phi, theta: d.sib.start.theta - p.went.theta };
  p.mine = on;
}

function bApply(st, p) {
  const d = bTurned(st, p);
  bSave(p);
  if (d) p.held = { phi: d.phi, theta: d.theta };
}

function bDirs(st, tree) {
  const row = tree.ab.row;
  if (bKind(row) !== 'copy') { const t = bDir(st, row); return t ? [{ t }] : []; }
  return tree.roots.map(({ back: { pair: p } }) => {
    const d = bChange(st, p);
    return p.base && d ? { t: { phi: p.base.phi + d.phi, theta: p.base.theta + d.theta, r: p.base.r }, pair: p } : null;
  }).filter(Boolean);
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
  if (bKind(row) === 'copy') return null;
  const c = bCopy(st, row), s = bS(st, row.from);
  return { phi: bWrap(s, c.phi + row.end.phi), theta: bWrap(s, c.theta + row.end.theta), r: row.start.r };
}

function bReview(st, r) {
  const v = bView(st, r);
  if (!v || bFloat(r)) return;
  bSave(r);
  delete r.turn;
  if (bKind(r) === 'angle') { r.start = v; return; }
  const c = bCopy(st, r), s = bS(st, r.from);
  r.start = { ...r.start, r: v.r };
  r.end = { phi: bWrap(s, v.phi - c.phi), theta: bWrap(s, v.theta - c.theta) };
}

function bSetKind(st, tree, kind) {
  const { row, a, b } = tree.ab, said = `${a.name}→${b.name} as ${kind}`, was = bKind(row);
  const keep = bKeep();
  if (['base', 'copy'].includes(kind) && b.via === tree.ab) bLetGo(st, b);
  bSave(row);
  row.kind = kind;
  if (kind === 'base') { bDone(); row.base = bView(st, row); return row.base ? '' : `${said}: ${a.name} or ${b.name} has no position to measure from`; }
  if (kind === 'copy') { bDone(); return bSetCopy(st, tree, said); }
  row.end ||= { phi: 0, theta: 0 };
  bLink(st, tree.ab);
  bReview(st, row);
  return bSettle(st, keep, [tree], said, was);
}

function bSetCopy(st, tree, said) {
  const row = tree.ab.row, pairs = tree.roots.map(({ back }) => back.pair);
  row.kind = 'copy';
  pairs.forEach(p => bSeat(st, p));
  if (!pairs.length) return `${said}: ${row.from.name} has no other row to pair with`;
  return pairs.some(p => p.base) ? '' : `${said}: ${row.from.name} or ${row.to.name} has no position to measure from`;
}

const bAt = (r, p) => `${r.from.name}→${r.to.name}${p ? ` in (${p.ab.a.name}→${p.ab.b.name}, ${p.ac.a.name}→${p.ac.c.name})` : ''}`;

function bPairCells(st, root, h, on) {
  const p = root.pair, d = bChange(st, p);
  if (!d) return [];
  const auto = h('input', { type: 'checkbox', checked: bAuto(p) });
  const turn = c0 => {
    const c1 = bChange(st, p), s = c1 && c1.sib.start;
    if (c0 && c1) bCarry(st, root, bRot(st, c1.sib, { phi: s.phi - c1.phi + c0.phi, theta: s.theta - c1.theta + c0.theta }, s));
    on.changed();
  };
  auto.onchange = () => { const c0 = bChange(st, p); p.automatic = auto.checked; bApply(st, p); turn(c0); };
  const apply = h('button', { textContent: 'apply turn' });
  apply.onclick = () => { const c0 = bChange(st, p); bApply(st, p); turn(c0); };
  const mine = h('button', { textContent: p.mine ? 'own step: on' : 'own step: off' });
  mine.onclick = () => { bSetMine(st, p, !p.mine); on.changed(); };
  const step = k => {
    const f = h('input', { type: 'number', value: p.mine ? p.step[k] : '', disabled: !p.mine, style: 'width: 4em' });
    f.onchange = () => { if (bWhole(+f.value)) p.step[k] = +f.value; on.changed(); };
    return f;
  };
  const ab = `${d.sib.from.name}→${d.sib.to.name}`;
  return [h('div', {}, [`(${ab}, ${p.ac.a.name}→${p.ac.c.name}): ${ab} ${!bAuto(p) ? 'turn applied' : 'change'} φ ${bN(d.phi)} θ ${bN(d.theta)} `,
    h('label', {}, [auto, ' automatic ']), ...(!bAuto(p) ? [apply] : []), mine, ' φ ', step('phi'), ' θ ', step('theta')])];
}

function bRowCells(st, tree, c, h, on) {
  const r = tree.ab.row, kind = h('select', {}, ['angle', 'change', 'copy', 'base'].map(k => h('option', { value: k, textContent: k, selected: bKind(r) === k })));
  kind.onchange = () => on.kind(kind.value);
  const txt = x => h('td', { textContent: x });
  if (bKind(r) === 'angle') return [h('td', {}, [kind]), c(r.start, 'phi', tree), c(r.start, 'theta', tree), c(r.start, 'r', tree), txt('—'), txt('—')];
  if (bKind(r) === 'change') {
    const k = bCopy(st, r);
    return [h('td', {}, [kind]), txt(`${bN(k.phi)} copy`), txt(`${bN(k.theta)} copy`), c(r.start, 'r', tree), c(r.end, 'phi', tree, 'end'), c(r.end, 'theta', tree, 'end')];
  }
  const copy = bKind(r) === 'copy', ts = bDirs(st, tree).map(e => e.t);
  const set = h('button', { textContent: copy ? 'set copy' : 'set base' });
  set.onclick = () => { if (copy) bSetCopy(st, tree, ''); else r.base = bView(st, r); on.changed(); };
  const three = ['phi', 'theta', 'r'].map(k => txt(ts.length ? ts.map(t => bF(t[k])).join(' / ') : '—'));
  const pairs = copy ? tree.roots.flatMap(({ back }) => bPairCells(st, back, h, on)) : [];
  return [h('td', {}, [kind]), ...three, h('td', { colSpan: 2 }, [...pairs, set])];
}

function bHow(st, o, r, p) {
  if (bKind(r) === 'angle') return `angle ${r.from.name}→${r.to.name}`;
  const d = p && bChange(st, p);
  if (bKind(r) === 'copy' && d) return `copy in (${d.sib.from.name}→${d.sib.to.name}, ${r.from.name}→${r.to.name}); stored φ ${bF(p.base.phi)} + ${d.sib.from.name}→${d.sib.to.name}'s change φ ${bN(d.phi)} θ ${bN(d.theta)}`;
  const ch = o.via && bKind(o.via.row) === 'change' ? o.via.row : null;
  if (bKind(r) === 'base' && ch) {
    const k = bCopy(st, ch);
    return `base; ${o.name} = ${ch.from.name} + (${ch.from.name}'s φ ${bN(k.phi)} θ ${bN(k.theta)} + ${o.name}'s own φ ${bN(ch.end.phi)} θ ${bN(ch.end.theta)})`;
  }
  return `${bKind(r)} ${r.from.name}→${r.to.name}`;
}

function bSecondBase(st, n, h, table) {
  const bases = n.out.filter(t => bFloat(t.ab.row)).flatMap(t => bDirs(st, t).map(e => ({ r: t.ab.row, ...e })));
  const seen = bases.length ? [...n.out.flatMap(t => bDirs(st, t).map(e => [t.ab.b, e.t, t.ab.row, e.pair])),
    ...n.in.flatMap(t => bDirs(st, t).map(e => [t.ab.a, bBack(st, t.ab.row, e.t), t.ab.row, e.pair]))] : [];
  return bases.flatMap(({ r, t: z, pair }) => [
    h('h2', { textContent: `${n.name}'s second base — the ${bAt(r, pair)} ${bKind(r)}, measured once at φ ${bF(z.phi)}, θ ${bF(z.theta)} and stored as 0` }),
    table(['toward', 'φ', 'θ', 'r', 'how'], seen.map(([o, t, x, p]) => h('tr', {}, [o.name, bN(t.phi - z.phi), bN(t.theta - z.theta), bN(t.r), bHow(st, o, x, p)]
      .map(v => h('td', { textContent: v })))))]);
}

function bLands(st, h, table) {
  const three = x => (x ? [x.phi, x.theta, x.r].map(bF) : ['—', '—', '—']);
  const rows = [...st.trees].flatMap(t => {
    const dirs = bDirs(st, t), r = t.ab.row;
    return (dirs.length ? dirs : [{}]).map(e => h('tr', {}, [bAt(r, e.pair), ...three(e.t), ...three(bView(st, r))].map(x => h('td', { textContent: x }))));
  });
  return [h('h2', { textContent: 'read only: each row, and the direction from its first node to its second node where they are placed' }),
    table(['a→b', 'row φ', 'θ', 'r', 'placed φ', 'θ', 'r'], rows)];
}
