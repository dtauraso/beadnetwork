const T_EPS = 1e-9;
const tName = r => r.from.name + r.to.name;
const tRows = st => [...st.trees].map(({ ab: { row } }) => [tName(row)]);
const tNodes = st => [...st.nodes].map(n => [n.name, n.idx, n.via ? n.via.a.name + n.via.b.name : null]);
const tSnap = st => ({ nodes: JSON.parse(JSON.stringify(tNodes(st))), rows: tRows(st) });
const tReload = st => { const s = JSON.parse(JSON.stringify(st, bPack)); bWire(s); return s; };
const tRow = (from, to, s) => ({ from, to, start: { phi: s[0], theta: s[1], r: s[2] } });

function tEdits() {
  let st = bPreset();
  const [ab, bc] = [...st.trees], [, , c] = [...st.nodes];
  let ab2, bc2;
  return [
    ['the preset: a→b φ 2 θ 0 r 2, b→c φ 4 θ 0 r 2', () => {}],
    ['a→b φ 3', () => bSetStart(st, ab, 'phi', 3)],
    ['a→b θ 2', () => bSetStart(st, ab, 'theta', 2)],
    ['a→b r 3', () => bSetStart(st, ab, 'r', 3)],
    ['a→b φ 4', () => bSetStart(st, ab, 'phi', 4)],
    ['a→b θ 1', () => bSetStart(st, ab, 'theta', 1)],
    ['add d (s 2) with c→d φ 1 θ 2 r 2, then a→b φ 5', () => {
      const d = bMake({ name: 'd', s: 2, known: true, idx: { r: 3, phi: 2, theta: 5 } });
      st.nodes.add(d);
      bMove(st, bAddTree(st, tRow(c, d, [1, 2, 2])));
      bSetStart(st, ab, 'phi', 5);
    }],
    ['a→b θ 4', () => bSetStart(st, ab, 'theta', 4)],
    ['a→b φ 6', () => bSetStart(st, ab, 'phi', 6)],
    ['a→b φ 3', () => bSetStart(st, ab, 'phi', 3)],
    ['a→b φ 2.5 — not a whole number, so the edit is undone', () => bSetStart(st, ab, 'phi', 2.5)],
    ['b→c φ 5', () => bSetStart(st, bc, 'phi', 5)],
    ['save and load', () => { st = tReload(st); [ab2, bc2] = [...st.trees]; }],
    ['after loading, a→b φ 4', () => bSetStart(st, ab2, 'phi', 4)],
    ['remove b→c', () => bDropTree(st, bc2)],
    ['remove a→b', () => bDropTree(st, ab2)],
    ['remove node c', () => bDropNode(st, [...st.nodes].find(n => n.name === 'c'))],
  ].map(([say, run]) => ({ say, run, snap: () => tSnap(st) }));
}

function tSame(want, got, at, out) {
  if (typeof want === 'number' && typeof got === 'number') {
    if (Math.abs(want - got) > T_EPS) out.push(`${at}: wanted ${want}, got ${got}`);
  } else if (want && got && typeof want === 'object' && typeof got === 'object') {
    const keys = new Set([...Object.keys(want), ...Object.keys(got)]);
    keys.forEach(k => tSame(want[k], got[k], at ? `${at} ${k}` : k, out));
  } else if (want !== got) out.push(`${at}: wanted ${JSON.stringify(want)}, got ${JSON.stringify(got)}`);
  return out;
}

function tCheckEdits() {
  return tEdits().map((e, i) => {
    let got, misses;
    try {
      e.run();
      got = e.snap();
      misses = tSame({ nodes: T_WANT[i].nodes, rows: T_WANT[i].rows }, got, '', []);
    } catch (err) { misses = [`threw: ${err.message}`]; }
    return { say: e.say, want: T_WANT[i], got, misses };
  });
}

function tDirsOf(st) {
  return {
    dirs: [...st.trees].map(t => [tName(t.ab.row), bDirs(st, t).map(({ t: d }) => [d.phi, d.theta, +d.r.toFixed(3)])]),
    nodes: [...st.nodes].map(n => [n.idx.r, n.idx.phi].map(v => +v.toFixed(3))),
  };
}

function tCheckSaves() {
  let st = bPreset(), s2, s3;
  return [
    ['the preset', 0, () => st],
    ['a→b φ 3', 1, () => { bSetStart(st, [...st.trees][0], 'phi', 3); return st; }],
    [`load the save the builder wrote at ${T_FROM}, before the rows moved into the trees, with its rotation fields taken out`, 2, () => { s2 = JSON.parse(JSON.stringify(T_SAVE)); bWire(s2); return s2; }],
    ['then a→b φ 4', 3, () => { bSetStart(s2, [...s2.trees][0], 'phi', 4); return s2; }],
    ['save that in the new format and load it', 3, () => (s3 = tReload(s2))],
    ['then a→b φ 5', 4, () => { bSetStart(s3, [...s3.trees][0], 'phi', 5); return s3; }],
  ].map(([say, i, run]) => {
    let got, misses;
    try { got = tDirsOf(run()); misses = tSame({ dirs: T_DIRS[i].dirs, nodes: T_DIRS[i].nodes }, got, '', []); }
    catch (err) { misses = [`threw: ${err.message}`]; }
    return { say, want: T_DIRS[i], got, misses };
  });
}

const tEl = (tag, props = {}, kids = []) => {
  const e = Object.assign(document.createElement(tag), props);
  kids.forEach(k => e.append(k));
  return e;
};
const tNum = v => (Number.isInteger(v) ? `${v}` : v.toFixed(4));
const tIdx = i => (i ? `r ${tNum(i.r)} · φ ${tNum(i.phi)} · θ ${tNum(i.theta)}` : '—');

function tNodeTable(want, got) {
  const head = tEl('tr', {}, ['node', 'wanted', 'got', 'via wanted', 'via got'].map(t => tEl('th', { textContent: t })));
  const names = [...new Set([...want.map(n => n[0]), ...(got || []).map(n => n[0])])];
  return tEl('table', {}, [tEl('thead', {}, [head]), tEl('tbody', {}, names.map(name => {
    const w = want.find(n => n[0] === name), g = (got || []).find(n => n[0] === name);
    const bad = tSame(w, g, '', []).length > 0;
    return tEl('tr', { className: bad ? 'bad' : '' }, [name, tIdx(w && w[1]), tIdx(g && g[1]), (w && w[2]) || '—', (g && g[2]) || '—']
      .map(t => tEl('td', { textContent: t })));
  }))]);
}

function tRowTable(want, got) {
  const head = tEl('tr', {}, ['row', 'wanted', 'got'].map(t => tEl('th', { textContent: t })));
  const names = [...new Set([...want.map(r => r[0]), ...(got || []).map(r => r[0])])];
  return tEl('table', {}, [tEl('thead', {}, [head]), tEl('tbody', {}, names.map(name => {
    const w = want.find(r => r[0] === name), g = (got || []).find(r => r[0] === name);
    const bad = tSame(w, g, '', []).length > 0;
    return tEl('tr', { className: bad ? 'bad' : '' }, [`${name[0]}→${name.slice(1)}`, w ? 'yes' : '—', g ? 'yes' : '—']
      .map(t => tEl('td', { textContent: t })));
  }))]);
}

function tDirTable(want, got) {
  const show = ds => (ds ? ds.map(d => `φ ${tNum(d[0])} θ ${tNum(d[1])} r ${tNum(d[2])}`).join(' / ') || '—' : '—');
  const head = tEl('tr', {}, ['row', 'directions wanted', 'directions got'].map(t => tEl('th', { textContent: t })));
  return tEl('table', {}, [tEl('thead', {}, [head]), tEl('tbody', {}, want.dirs.map(([name, ds], i) => {
    const g = got && got.dirs[i], bad = tSame(ds, g && g[1], '', []).length > 0;
    return tEl('tr', { className: bad ? 'bad' : '' }, [`${name[0]}→${name.slice(1)}`, show(ds), show(g && g[1])].map(t => tEl('td', { textContent: t })));
  }))]);
}

function tShow(id, checks, whys, tables) {
  const box = document.getElementById(id);
  box.replaceChildren(...checks.map((c, i) => {
    const ok = !c.misses.length;
    return tEl('details', { className: ok ? 'check ok' : 'check bad', open: !ok }, [
      tEl('summary', {}, [tEl('span', { className: 'mark', textContent: ok ? '✓' : '✗' }), ` ${i + 1}. ${c.say}`,
        tEl('span', { className: 'why', textContent: whys[i] })]),
      ...c.misses.slice(0, 8).map(m => tEl('div', { className: 'miss', textContent: m })),
      ...tables(c).map(t => tEl('div', { className: 'table-wrap' }, [t])),
    ]);
  }));
  return checks.filter(c => !c.misses.length).length;
}

const tEdit = tCheckEdits(), tSave = tCheckSaves();
const tOkEdit = tShow('edits', tEdit, T_WHY_EDITS, c => [tNodeTable(c.want.nodes, c.got && c.got.nodes), tRowTable(c.want.rows, c.got && c.got.rows)]);
const tOkSave = tShow('saves', tSave, T_WHY_SAVES, c => [tDirTable(c.want, c.got)]);
const tAll = tOkEdit === tEdit.length && tOkSave === tSave.length;
document.getElementById('summary').className = `summary ${tAll ? 'ok' : 'bad'}`;
document.getElementById('summary').textContent = `${tOkEdit} of ${tEdit.length} edit steps and ${tOkSave} of ${tSave.length} save steps give what ${T_AT} gave`;
