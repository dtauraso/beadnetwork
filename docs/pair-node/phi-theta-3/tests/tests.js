const T_EPS = 1e-9;
const tName = e => e.from.name + e.to.name;
const tRows = builder => [...builder.openStep().edges].map(e => [tName(e)]);
const tNodes = builder => builder.shown().map(n => [n.name, n.index(builder)]);
const tSnap = builder => ({ nodes: JSON.parse(JSON.stringify(tNodes(builder))), rows: tRows(builder) });
const tReload = builder => BBuilder.from(JSON.parse(JSON.stringify(builder)));

function tEdits() {
  let builder = bPreset();
  const [ab, bc] = [...builder.openStep().edges], c = builder.openStep().find('c');
  let ab2, bc2;
  return [
    ['the preset: a→b φ 2 θ 0 r 2, b→c φ 4 θ 0 r 2', () => {}],
    ['a→b φ 3', () => ab.setStart(builder, 'phi', 3)],
    ['a→b θ 2', () => ab.setStart(builder, 'theta', 2)],
    ['a→b r 3', () => ab.setStart(builder, 'r', 3)],
    ['a→b φ 4', () => ab.setStart(builder, 'phi', 4)],
    ['a→b θ 1', () => ab.setStart(builder, 'theta', 1)],
    ['add d (s 2) with c→d φ 1 θ 2 r 2, then a→b φ 5', () => {
      const d = builder.openStep().addNode(new BNode({ name: 'd', s: 2, at: { r: 3, phi: 2, theta: 5 } }));
      builder.openStep().addEdge(new BEdge({ from: c, to: d, start: { phi: 1, theta: 2, r: 2 } })).move(builder);
      ab.setStart(builder, 'phi', 5);
    }],
    ['a→b θ 4', () => ab.setStart(builder, 'theta', 4)],
    ['a→b φ 6', () => ab.setStart(builder, 'phi', 6)],
    ['a→b φ 3', () => ab.setStart(builder, 'phi', 3)],
    ['a→b φ 2.5', () => ab.setStart(builder, 'phi', 2.5)],
    ['b→c φ 5', () => bc.setStart(builder, 'phi', 5)],
    ['save and load', () => { builder = tReload(builder); [ab2, bc2] = [...builder.openStep().edges]; }],
    ['after loading, a→b φ 4', () => ab2.setStart(builder, 'phi', 4)],
    ['remove b→c', () => builder.openStep().dropEdge(bc2)],
    ['remove a→b', () => builder.openStep().dropEdge(ab2)],
    ['remove node c', () => builder.dropNode(builder.openStep().find('c'))],
  ].map(([say, run]) => ({ say, run, snap: () => tSnap(builder) }));
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

function tDirsOf(builder) {
  return {
    dirs: [...builder.openStep().edges].map(e => { const d = e.dir(builder); return [tName(e), [[d.phi, d.theta, +d.r.toFixed(3)]]]; }),
    nodes: builder.shown().map(n => { const I = n.index(builder); return [I.r, I.phi].map(v => +v.toFixed(3)); }),
  };
}

function tCheckSaves() {
  let builder = bPreset(), s2;
  return [
    ['the preset', 0, () => builder],
    ['a→b φ 3', 1, () => { [...builder.openStep().edges][0].setStart(builder, 'phi', 3); return builder; }],
    ['save that and load it', 1, () => (s2 = tReload(builder))],
    ['then a→b φ 4', 2, () => { [...s2.openStep().edges][0].setStart(s2, 'phi', 4); return s2; }],
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
  const head = tEl('tr', {}, ['node', 'wanted', 'got'].map(t => tEl('th', { textContent: t })));
  const names = [...new Set([...want.map(n => n[0]), ...(got || []).map(n => n[0])])];
  return tEl('table', {}, [tEl('thead', {}, [head]), tEl('tbody', {}, names.map(name => {
    const w = want.find(n => n[0] === name), g = (got || []).find(n => n[0] === name);
    const bad = tSame(w, g, '', []).length > 0;
    return tEl('tr', { className: bad ? 'bad' : '' }, [name, tIdx(w && w[1]), tIdx(g && g[1])]
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
