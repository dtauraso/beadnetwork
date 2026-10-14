const bKind = row => row.kind || 'angle';
const bFloat = row => bKind(row) === 'base' || bKind(row) === 'copy';
const bAuto = row => row.automatic !== false;
const bSib = (st, row) => st.rows.find(r => r !== row && r.from === row.from && bKind(r) === 'angle');

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
  const into = st.rows.find(r => r.to === row.from && bKind(r) === 'angle');
  if (!into) return { phi: 0, theta: 0 };
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
  if (bKind(r) === 'angle') { r.start = v; return; }
  const c = bCopy(st, r), s = bS(st, r.from);
  r.start = { ...r.start, r: v.r };
  r.end = { phi: bWrap(s, v.phi - c.phi), theta: bWrap(s, v.theta - c.theta) };
}

const bNode = (st, name) => st.nodes.find(x => x.name === name);

function bLetGo(st, name) {
  const n = bNode(st, name), from = n && n.ref && bNode(st, n.ref);
  if (from) from.out = (from.out || []).filter(x => x !== name);
  if (n) delete n.ref;
}

function bRefer(st, row) {
  const n = bNode(st, row.to), from = bNode(st, row.from);
  if (!n || !from || bFloat(row) || n.ref === row.from) return;
  bLetGo(st, row.to);
  n.ref = row.from;
  (from.out ||= []).push(row.to);
}

function bDropRow(st, row) {
  if ((bNode(st, row.to) || {}).ref === row.from) bLetGo(st, row.to);
  st.rows = st.rows.filter(x => x !== row);
}

function bDropNode(st, n) {
  (n.out || []).forEach(x => delete bNode(st, x).ref);
  bLetGo(st, n.name);
  st.nodes = st.nodes.filter(x => x !== n);
  st.rows = st.rows.filter(r => r.from !== n.name && r.to !== n.name);
}

function bLink(st, row) {
  if (!(bNode(st, row.to) || {}).ref) bRefer(st, row);
}

function bTake(st, name) {
  const n = bNode(st, name), P = n && n.known && n.ref && bNodePos(st, n.ref);
  const row = P && st.rows.find(r => r.from === n.ref && r.to === name), t = row && bDir(st, row);
  if (!t) return false;
  const J = bIndexOf(st, bAdd(P, bVec(st, row, t)));
  n.idx = { ...J, theta: J.theta < 0 ? J.theta + st.Ttheta : J.theta };
  return true;
}

function bSend(st, name) {
  const inbox = [...(bNode(st, name).out || [])];
  for (let left = st.nodes.length; inbox.length; left--) {
    if (!left) return false;
    const to = inbox.shift();
    if (bTake(st, to)) inbox.push(...(bNode(st, to).out || []));
  }
  return true;
}

function bMove(st, row) {
  bRefer(st, row);
  return !bTake(st, row.to) || bSend(st, row.to);
}

function bSetKind(st, row, kind) {
  const said = `${row.from}→${row.to} as ${kind}`, was = bKind(row);
  const keep = JSON.stringify([st.nodes, st.rows]);
  if (['base', 'copy'].includes(kind) && (bNode(st, row.to) || {}).ref === row.from) bLetGo(st, row.to);
  row.kind = kind;
  if (kind === 'base') { row.base = bView(st, row); return row.base ? '' : `${said}: ${row.from} or ${row.to} has no position to measure from`; }
  if (kind === 'copy') return bSetCopy(st, row, said);
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
  if (!sib) return `${said}: ${row.from} has no angle row to copy the change of`;
  return row.base ? '' : `${said}: ${row.from} or ${row.to} has no position to measure from`;
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
  const t = bDir(st, r), what = d ? `${d.sib.from}→${d.sib.to} ${!bAuto(r) ? "turn applied" : "change"} φ ${bN(d.phi)} θ ${bN(d.theta)} ` : '';
  const auto = h('input', { type: 'checkbox', checked: bAuto(r) });
  auto.onchange = () => { r.automatic = auto.checked; bApply(st, r); on.changed(); };
  const apply = h('button', { textContent: 'apply turn' });
  apply.onclick = () => { bApply(st, r); on.changed(); };
  const extra = copy && d ? [h('label', {}, [auto, ' automatic ']), ...(!bAuto(r) ? [apply] : [])] : [];
  return [h('td', {}, [kind]), ...(t ? [t.phi, t.theta, t.r].map(x => txt(bF(x))) : ['—', '—', '—'].map(txt)), h('td', { colSpan: 2 }, [what, ...extra, set])];
}

function bHow(st, o, r) {
  if (bKind(r) === 'angle') return `angle ${r.from}→${r.to}`;
  const d = bChange(st, r);
  if (bKind(r) === 'copy' && d) return `copy; stored φ ${bF(r.base.phi)} + ${d.sib.from}→${d.sib.to}'s change φ ${bN(d.phi)} θ ${bN(d.theta)}`;
  const ch = st.rows.find(x => x.to === o && bKind(x) === 'change');
  if (bKind(r) === 'base' && ch) {
    const k = bCopy(st, ch);
    return `base; ${o} = ${ch.from} + (${ch.from}'s φ ${bN(k.phi)} θ ${bN(k.theta)} + ${o}'s own φ ${bN(ch.end.phi)} θ ${bN(ch.end.theta)})`;
  }
  return `${bKind(r)} ${r.from}→${r.to}`;
}

function bSecondBase(st, n, h, table) {
  return st.rows.filter(r => r.from === n.name && bFloat(r) && bDir(st, r)).flatMap(base => {
    const z = bDir(st, base);
    const seen = st.rows.flatMap(r => (r.from === n.name ? [[r.to, bDir(st, r), r]] : r.to === n.name ? [[r.from, bDir(st, r) && bBack(st, r), r]] : []))
      .filter(x => x[1]);
    return [h('h2', { textContent: `${n.name}'s second base — the ${n.name}→${base.to} base, measured once at φ ${bF(z.phi)}, θ ${bF(z.theta)} and stored as 0` }),
      table(['toward', 'φ', 'θ', 'r', 'how'], seen.map(([o, t, r]) => h('tr', {}, [o, bN(t.phi - z.phi), bN(t.theta - z.theta), bN(t.r), bHow(st, o, r)]
        .map(x => h('td', { textContent: x })))))];
  });
}

function bLands(st, h, table) {
  const three = x => (x ? [x.phi, x.theta, x.r].map(bF) : ['—', '—', '—']);
  const rows = st.rows.map(r => h('tr', {}, [`${r.from}→${r.to}`, ...three(bDir(st, r)), ...three(bView(st, r))]
    .map(x => h('td', { textContent: x }))));
  return [h('h2', { textContent: 'read only: each row, and the direction from its first node to its second node where they are placed' }),
    table(['a→b', 'row φ', 'θ', 'r', 'placed φ', 'θ', 'r'], rows)];
}
