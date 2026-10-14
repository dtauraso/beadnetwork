const bKind = row => row.kind || 'angle';

function bCopy(st, row) {
  const into = st.rows.find(r => r.to === row.from && bKind(r) === 'angle');
  if (!into) return { phi: 0, theta: 0 };
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
  if (!v || bKind(r) === 'base') return;
  if (bKind(r) === 'angle') { r.start = v; return; }
  const c = bCopy(st, r), s = bS(st, r.from);
  r.start = { ...r.start, r: v.r };
  r.end = { phi: bWrap(s, v.phi - c.phi), theta: bWrap(s, v.theta - c.theta) };
}

function bCarry(st, row) {
  const used = [], names = [row.from];
  const go = r => {
    const n = st.nodes.find(x => x.name === r.to), P = bNodePos(st, r.from), v = bRowVectors(st, r.to).find(x => x.row === r);
    if (!n || !n.known || !P || !v || names.includes(n.name)) return;
    const J = bIndexOf(st, bAdd(P, v.d));
    n.idx = { ...J, theta: J.theta < 0 ? J.theta + st.Ttheta : J.theta };
    used.push(r); names.push(n.name);
    st.rows.filter(x => x.from === n.name && bKind(x) !== 'base').forEach(go);
  };
  go(row);
  return { used, names: names.slice(1) };
}

function bSetKind(st, row, kind) {
  const said = `${row.from}→${row.to} as ${kind}`, was = bKind(row);
  const keep = JSON.stringify([st.nodes, st.rows]);
  row.kind = kind;
  if (kind === 'base') { row.base = bView(st, row); return row.base ? '' : `${said}: ${row.from} or ${row.to} has no position to measure from`; }
  row.end ||= { phi: 0, theta: 0 };
  bReview(st, row);
  return bSettle(st, keep, [row], said, was);
}

function bRowCells(st, r, c, h, on) {
  const kind = h('select', {}, ['angle', 'change', 'base'].map(k => h('option', { value: k, textContent: k, selected: bKind(r) === k })));
  kind.onchange = () => on.kind(kind.value);
  const txt = x => h('td', { textContent: x });
  if (bKind(r) === 'angle') return [h('td', {}, [kind]), c(r.start, 'phi', r), c(r.start, 'theta', r), c(r.start, 'r', r), txt('—'), txt('—')];
  if (bKind(r) === 'change') {
    const k = bCopy(st, r);
    return [h('td', {}, [kind]), txt(`${bN(k.phi)} copy`), txt(`${bN(k.theta)} copy`), c(r.start, 'r', r), c(r.end, 'phi', r, 'end'), c(r.end, 'theta', r, 'end')];
  }
  const set = h('button', { textContent: 'set base' });
  set.onclick = () => { r.base = bView(st, r); on.changed(); };
  const t = r.base;
  return [h('td', {}, [kind]), ...(t ? [t.phi, t.theta, t.r].map(x => txt(bF(x))) : ['—', '—', '—'].map(txt)), h('td', { colSpan: 2 }, [set])];
}

function bHow(st, o, r) {
  if (bKind(r) === 'angle') return `angle ${r.from}→${r.to}`;
  const ch = st.rows.find(x => x.to === o && bKind(x) === 'change');
  if (bKind(r) === 'base' && ch) {
    const k = bCopy(st, ch);
    return `base; ${o} = ${ch.from} + (${ch.from}'s φ ${bN(k.phi)} θ ${bN(k.theta)} + ${o}'s own φ ${bN(ch.end.phi)} θ ${bN(ch.end.theta)})`;
  }
  return `${bKind(r)} ${r.from}→${r.to}`;
}

function bSecondBase(st, n, h, table) {
  return st.rows.filter(r => r.from === n.name && bKind(r) === 'base' && r.base).flatMap(base => {
    const z = base.base;
    const seen = st.rows.flatMap(r => (r.from === n.name ? [[r.to, bDir(st, r), r]] : r.to === n.name ? [[r.from, bDir(st, r) && bBack(st, r), r]] : []))
      .filter(x => x[1]);
    return [h('h2', { textContent: `${n.name}'s second base — the ${n.name}→${base.to} base, measured once at φ ${bF(z.phi)}, θ ${bF(z.theta)} and stored as 0` }),
      table(['toward', 'φ', 'θ', 'r', 'how'], seen.map(([o, t, r]) => h('tr', {}, [o, bN(t.phi - z.phi), bN(t.theta - z.theta), bN(t.r), bHow(st, o, r)]
        .map(x => h('td', { textContent: x })))))];
  });
}
