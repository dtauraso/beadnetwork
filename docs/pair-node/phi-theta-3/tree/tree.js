const T_KEY = 'pair-builder-v5', T_SEQS = 'pair-builder-seqs-v1';
const T_HEAD = 30, T_LINE = 17, T_CH = 7.2, T_GX = 18, T_GY = 44;
const tNum = v => `${+v.toFixed(3)}`;
const tAng = (o, ks = ['phi', 'theta', 'r']) => (o ? ks.filter(k => k in o).map(k => `${{ phi: 'φ', theta: 'θ', r: 'r' }[k]} ${tNum(o[k])}`).join(' | ') : '—');
const tAB = ab => `${ab.a.name}→${ab.b.name}`;
const tNames = (ts, name) => (ts.length ? ts.map(name).join(', ') : '—');
const tOn = o => tNames(Object.keys(o || {}).filter(k => o[k]), k => k);
const tEsc =s => s.replace(/&/g, '&amp;').replace(/</g, '&lt;');

function tLoad() {
  try {
    const s = JSON.parse(localStorage.getItem(T_KEY));
    if (s) { bWire(s); return { builder: s, from: `the builder's save in this browser (${T_KEY})` }; }
  } catch { }
  return { builder: bPreset(), from: 'the three-node preset — the builder has saved nothing in this browser yet' };
}

function tSeq() {
  try {
    const book = JSON.parse(localStorage.getItem(T_SEQS));
    if (book && book.seqs[book.at]) return { label: book.seqs[book.at].name, lines: [] };
  } catch { }
  return { label: 'sequence 1', lines: [] };
}

function tState(builder) {
  return { label: 'builder', lines: [['s', tNum(builder.s)], ['ρ', tNum(builder.rho)], ['T_φ', tNum(builder.Tphi)], ['T_θ', tNum(builder.Ttheta)], ['c_r', tNum(builder.cr)],
    ['item', builder.item ? builder.item.name : '—', 'ref'], ['made', `${builder.made}`], ['step', `${builder.step}`], ['nodes', `${builder.nodes.size}`], ['trees', `${builder.trees.size}`],
    ['picked', tNames(builder.picked || [], t => tAB(t.ab)), 'ref'], ['methods', tOn(builder.methods)]] };
}

const tWorked = rows => [['', 'not stored', 'note'], ...rows];

function tTree(builder, t) {
  const { row } = t.ab;
  return [['', tAB(t.ab), 'part'], ['from', row.from.name, 'ref'], ['to', row.to.name, 'ref'], ['start', tAng(row.start)], ['Δ', tAng(row.delta, ['phi', 'theta'])],
    ['increment', tAng(row.increment, ['phi', 'theta'])],
    ...tWorked([['~start + Δ', tAng(bDir(builder, row))]])];
}

function tNode(builder, n) {
  const C = bNodePos(builder, n);
  return { label: `node ${n.name}`, lines: [['s', n.s ? tNum(n.s) : `— (the scene's s, ${tNum(builder.s)})`],
    ['idx', tAng(n.idx)],
    ['out', tNames(n.out, t => tAB(t.ab)), 'ref'], ['in', tNames(n.in, t => tAB(t.ab)), 'ref'], ...tWorked([['~C', C ? ['x', 'y', 'z'].map((k, i) => `${k} ${tNum(C[i])}`).join(' | ') : '—']])] };
}

function tGraph(builder) {
  const trees = [...builder.trees], nodes = [...builder.nodes], box = new Map([['builder', tState(builder)], ['seq', tSeq()]]), kids = new Map([['builder', ['seq']], ['seq', []]]);
  if (trees.length) {
    box.set('edges', { label: 'edges', lines: trees.flatMap(t => tTree(builder, t)) });
    kids.set('edges', nodes.filter(n => n.in.length || n.out.length));
    kids.get('seq').push('edges');
  }
  nodes.forEach(n => {
    box.set(n, tNode(builder, n)); kids.set(n, []);
    if (!n.in.length && !n.out.length) kids.get('seq').push(n);
  });
  return { box, kids, levels: [['builder'], ['seq'], trees.length ? ['edges'] : [], nodes].filter(l => l.length) };
}

function tLayout({ box, levels }) {
  const at = new Map();
  levels.forEach(l => l.forEach(k => {
    const { label, lines } = box.get(k), kw = Math.max(...lines.map(([n]) => n.replace(/^~/, '').length)) + 2;
    const chars = Math.max(label.length, ...lines.map(([, v, kind]) => (kind === 'note' || kind === 'part' ? 0 : kw) + v.length));
    at.set(k, { w: chars * T_CH + 24, h: T_HEAD + lines.length * T_LINE + 8, kw: kw * T_CH });
  }));
  const span = l => l.reduce((s, k) => s + at.get(k).w, 0) + (l.length - 1) * T_GX, W = Math.max(...levels.map(span));
  let y = 0;
  levels.forEach(l => {
    let x = (W - span(l)) / 2;
    l.forEach(k => { const n = at.get(k); n.x = x + n.w / 2; n.y = y; x += n.w + T_GX; });
    y += Math.max(...l.map(k => at.get(k).h)) + T_GY;
  });
  return { at, W, H: y - T_GY };
}

function tSvg(builder) {
  const g = tGraph(builder), { at, W, H } = tLayout(g), parts = [];
  g.kids.forEach((cs, k) => cs.forEach(c => {
    const from = [...g.kids].filter(([, ks]) => ks.includes(c)).map(([p]) => p).sort((p, q) => at.get(p).x - at.get(q).x);
    const off = (from.indexOf(k) - (from.length - 1) / 2) * 16, p = at.get(k), n = at.get(c);
    parts.push(`<path class="edge" d="M${p.x} ${p.y + p.h} L${n.x + off} ${n.y}"/>`);
  }));
  at.forEach((n, k) => {
    const { label, lines } = g.box.get(k), left = n.x - n.w / 2;
    parts.push(`<rect class="box" x="${left}" y="${n.y}" width="${n.w}" height="${n.h}" rx="${g.kids.get(k).length ? 8 : 14}"/>`);
    parts.push(`<text class="name" x="${n.x}" y="${n.y + T_HEAD / 2 + 2}">${tEsc(label)}</text>`);
    lines.forEach(([name, v, kind], i) => {
      const calc = name.startsWith('~'), ref = kind === 'ref', y = n.y + T_HEAD + i * T_LINE + T_LINE / 2;
      if (kind === 'note') {
        parts.push(`<line class="split" x1="${left + 8}" x2="${left + n.w - 8}" y1="${y - T_LINE / 2}" y2="${y - T_LINE / 2}"/>`);
        parts.push(`<text class="l data note" x="${left + 12}" y="${y}">${tEsc(v)}</text>`);
        return;
      }
      if (kind === 'part') {
        if (i) parts.push(`<line class="rule" x1="${left}" x2="${left + n.w}" y1="${y - T_LINE / 2}" y2="${y - T_LINE / 2}"/>`);
        parts.push(`<text class="l data part" x="${left + 12}" y="${y}">${tEsc(v)}</text>`);
        return;
      }
      parts.push(`<text class="l data${calc ? ' calc' : ''}" y="${y}"><tspan class="k" x="${left + 12}">${tEsc(calc ? name.slice(1) : name)}</tspan>`
        + `<tspan class="v" x="${left + 12 + n.kw}">${ref && v !== '—' ? v.split(', ').map(r => `<tspan class="ref">${tEsc(r)}</tspan>`).join(', ') : tEsc(v)}</tspan></text>`);
    });
  });
  return { view: `-8 -8 ${W + 16} ${H + 16}`, width: W + 16, body: parts.join('') };
}

function tDraw() {
  const { builder, from } = tLoad(), svg = document.getElementById('tree'), { view, width, body } = tSvg(builder);
  document.getElementById('source').textContent = `from ${from}, read at ${new Date().toLocaleTimeString()}`;
  svg.setAttribute('viewBox', view);
  svg.style.width = `${width}px`;
  svg.innerHTML = body;
}

window.addEventListener('storage', e => { if (e.key === T_KEY || e.key === T_SEQS) tDraw(); });
tDraw();
