const bKind = row => row.kind || 'angle';
const bFloat = row => bKind(row) === 'base' || bKind(row) === 'copy';
const bAuto = row => row.automatic !== false;
function bSib(st, row) {
  const it = row.from.angles.values(), first = it.next().value;
  return first === row ? it.next().value : first;
}

function bTurned(st, row) {
  const sib = bSib(st, row);
  return sib && row.ref ? { sib, phi: sib.start.phi - row.ref.phi, theta: sib.start.theta - row.ref.theta } : null;
}

function bChange(st, row) {
  const d = bTurned(st, row);
  return d && !bAuto(row) ? { sib: d.sib, phi: row.held.phi, theta: row.held.theta } : d;
}

function bApply(st, row) {
  const d = bTurned(st, row);
  if (d) row.held = { phi: d.phi, theta: d.theta };
}

function bCopy(st, row) {
  const into = row.from.via;
  if (!into || bKind(into) !== 'angle') return { phi: 0, theta: 0 };
  const k = bS(st, row.from) / bS(st, into.from);
  return { phi: into.start.phi * k, theta: into.start.theta * k };
}

function bDir(st, row) {
  if (bKind(row) === 'base') return row.base || null;
  if (bKind(row) === 'angle') return row.start;
  if (bKind(row) === 'copy') {
    const d = bChange(st, row);
    return row.base && d ? { phi: row.base.phi + d.phi, theta: row.base.theta + d.theta, r: row.base.r } : null;
  }
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

function bSetKind(st, row, kind) {
  const said = `${row.from.name}→${row.to.name} as ${kind}`, was = bKind(row);
  const keep = bKeep();
  if (['base', 'copy'].includes(kind) && row.to.via === row) bLetGo(row.to);
  bSave(row); bSave(row.from);
  row.kind = kind;
  if (kind === 'angle') row.from.angles.add(row); else row.from.angles.delete(row);
  if (kind === 'base') { bDone(); row.base = bView(st, row); return row.base ? '' : `${said}: ${row.from.name} or ${row.to.name} has no position to measure from`; }
  if (kind === 'copy') { bDone(); return bSetCopy(st, row, said); }
  row.end ||= { phi: 0, theta: 0 };
  bLink(st, row);
  bReview(st, row);
  return bSettle(st, keep, [row], said, was);
}

function bSetCopy(st, row, said) {
  const sib = bSib(st, row);
  row.kind = 'copy';
  row.base = bView(st, row);
  row.ref = sib && { phi: sib.start.phi, theta: sib.start.theta };
  row.held = { phi: 0, theta: 0 };
  if (!sib) return `${said}: ${row.from.name} has no angle row to copy the change of`;
  return row.base ? '' : `${said}: ${row.from.name} or ${row.to.name} has no position to measure from`;
}

function bRowCells(st, r, c, h, on) {
  const kind = h('select', {}, ['angle', 'change', 'copy', 'base'].map(k => h('option', { value: k, textContent: k, selected: bKind(r) === k })));
  kind.onchange = () => on.kind(kind.value);
  const txt = x => h('td', { textContent: x });
  if (bKind(r) === 'angle') return [h('td', {}, [kind]), c(r.start, 'phi', r), c(r.start, 'theta', r), c(r.start, 'r', r), txt('—'), txt('—')];
  if (bKind(r) === 'change') {
    const k = bCopy(st, r);
    return [h('td', {}, [kind]), txt(`${bN(k.phi)} copy`), txt(`${bN(k.theta)} copy`), c(r.start, 'r', r), c(r.end, 'phi', r, 'end'), c(r.end, 'theta', r, 'end')];
  }
  const copy = bKind(r) === 'copy', d = copy && bChange(st, r);
  const set = h('button', { textContent: copy ? 'set copy' : 'set base' });
  set.onclick = () => { if (copy) bSetCopy(st, r, ''); else r.base = bView(st, r); on.changed(); };
  const t = bDir(st, r), what = d ? `${d.sib.from.name}→${d.sib.to.name} ${!bAuto(r) ? "turn applied" : "change"} φ ${bN(d.phi)} θ ${bN(d.theta)} ` : '';
  const auto = h('input', { type: 'checkbox', checked: bAuto(r) });
  const turn = c0 => {
    const c1 = bChange(st, r), s = c1 && c1.sib.start;
    if (c0 && c1) bCarry(st, r, bRot(st, c1.sib, { phi: s.phi - c1.phi + c0.phi, theta: s.theta - c1.theta + c0.theta }, s));
    on.changed();
  };
  auto.onchange = () => { const c0 = bChange(st, r); r.automatic = auto.checked; bApply(st, r); turn(c0); };
  const apply = h('button', { textContent: 'apply turn' });
  apply.onclick = () => { const c0 = bChange(st, r); bApply(st, r); turn(c0); };
  const extra = copy && d ? [h('label', {}, [auto, ' automatic ']), ...(!bAuto(r) ? [apply] : [])] : [];
  return [h('td', {}, [kind]), ...(t ? [t.phi, t.theta, t.r].map(x => txt(bF(x))) : ['—', '—', '—'].map(txt)), h('td', { colSpan: 2 }, [what, ...extra, set])];
}

function bHow(st, o, r) {
  if (bKind(r) === 'angle') return `angle ${r.from.name}→${r.to.name}`;
  const d = bChange(st, r);
  if (bKind(r) === 'copy' && d) return `copy; stored φ ${bF(r.base.phi)} + ${d.sib.from.name}→${d.sib.to.name}'s change φ ${bN(d.phi)} θ ${bN(d.theta)}`;
  const ch = o.via && bKind(o.via) === 'change' ? o.via : null;
  if (bKind(r) === 'base' && ch) {
    const k = bCopy(st, ch);
    return `base; ${o.name} = ${ch.from.name} + (${ch.from.name}'s φ ${bN(k.phi)} θ ${bN(k.theta)} + ${o.name}'s own φ ${bN(ch.end.phi)} θ ${bN(ch.end.theta)})`;
  }
  return `${bKind(r)} ${r.from.name}→${r.to.name}`;
}

function bSecondBase(st, n, h, table) {
  const bases = [...n.rowsOut].filter(r => bFloat(r) && bDir(st, r));
  const seen = bases.length ? [...[...n.rowsOut].map(r => [r.to, bDir(st, r), r]), ...[...n.rowsIn].map(r => [r.from, bDir(st, r) && bBack(st, r), r])]
    .filter(x => x[1]) : [];
  return bases.flatMap(base => {
    const z = bDir(st, base);
    return [h('h2', { textContent: `${n.name}'s second base — the ${n.name}→${base.to.name} base, measured once at φ ${bF(z.phi)}, θ ${bF(z.theta)} and stored as 0` }),
      table(['toward', 'φ', 'θ', 'r', 'how'], seen.map(([o, t, r]) => h('tr', {}, [o.name, bN(t.phi - z.phi), bN(t.theta - z.theta), bN(t.r), bHow(st, o, r)]
        .map(x => h('td', { textContent: x })))))];
  });
}

function bLands(st, h, table) {
  const three = x => (x ? [x.phi, x.theta, x.r].map(bF) : ['—', '—', '—']);
  const rows = [...st.rows].map(r => h('tr', {}, [`${r.from.name}→${r.to.name}`, ...three(bDir(st, r)), ...three(bView(st, r))]
    .map(x => h('td', { textContent: x }))));
  return [h('h2', { textContent: 'read only: each row, and the direction from its first node to its second node where they are placed' }),
    table(['a→b', 'row φ', 'θ', 'r', 'placed φ', 'θ', 'r'], rows)];
}
