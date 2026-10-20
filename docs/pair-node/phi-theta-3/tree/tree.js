const T_KEY = 'pair-builder-v5';
const T_HEAD = 30, T_LINE = 17, T_CH = 7.2, T_GX = 18, T_GY = 44;
const tNum = v => (Number.isInteger(v) ? `${v}` : v.toFixed(3));
const tAng = (o, ks = ['phi', 'theta', 'r']) => (o ? ks.filter(k => k in o).map(k => `${{ phi: 'φ', theta: 'θ', r: 'r' }[k]} ${tNum(o[k])}`).join(' · ') : '—');
const tAB = ab => `${ab.a.name}→${ab.b.name}`;
const tNames = (ts, name) => (ts.length ? ts.map(name).join(', ') : '—');
const tOn = o => tNames(Object.keys(o || {}).filter(k => o[k]), k => k);
const tEsc =s => s.replace(/&/g, '&amp;').replace(/</g, '&lt;');

function tLoad() {
  try {
    const s = JSON.parse(localStorage.getItem(T_KEY));
    if (s) { bWire(s); return { st: s, from: `the builder's save in this browser (${T_KEY})` }; }
  } catch { }
  return { st: bPreset(), from: 'the three-node preset — the builder has saved nothing in this browser yet' };
}

function tState(st) {
  return { label: 'st', lines: [`s ${tNum(st.s)}`, `ρ ${tNum(st.rho)}`, `T_φ ${tNum(st.Tphi)}`, `T_θ ${tNum(st.Ttheta)}`, `c_r ${tNum(st.cr)}`,
    `item ${st.item ? st.item.name : '—'}`, `made ${st.made}`, `step ${st.step}`, `nodes ${st.nodes.size}`, `trees ${st.trees.size}`,
    `picked ${tNames(st.picked || [], t => tAB(t.ab))}`, `on ${tOn(st.on)}`, `cells ${st.cells !== false ? 'on' : 'off'}`, `spheres ${tOn(st.spheres)}`] };
}

function tTree(st, t) {
  const { row, b } = t.ab;
  return { label: tAB(t.ab), lines: [`from ${row.from.name}`, `to ${row.to.name}`, `start ${tAng(row.start)}`, `Δ ${tAng(row.delta, ['phi', 'theta'])}`,
    `by ${tAng(row.by, ['phi', 'theta'])}`,
    `~start + Δ ${tAng(bDir(st, row))}`, `places b ${b.via === t.ab ? 'yes — b\'s via' : 'no'}`] };
}

function tNode(st, n) {
  const C = bNodePos(st, n);
  return { label: `node ${n.name}`, lines: [`s ${n.s ? tNum(n.s) : `— (the scene's s, ${tNum(st.s)})`}`, `known ${n.known ? 'yes' : 'no'}`,
    `idx ${n.known ? `I_r ${tNum(n.idx.r)} · I_φ ${tNum(n.idx.phi)} · I_θ ${tNum(n.idx.theta)}` : '—'}`, `~C ${C ? bPt(C) : '—'}`,
    `sender ${n.sender ? n.sender.name : '—'}`, `via ${n.via ? tAB(n.via) : '—'}`,
    `out ${tNames(n.out, t => tAB(t.ab))}`, `in ${tNames(n.in, t => tAB(t.ab))}`] };
}

function tGraph(st) {
  const trees = [...st.trees], nodes = [...st.nodes], box = new Map([['st', tState(st)]]), kids = new Map([['st', []]]);
  trees.forEach(t => { box.set(t, tTree(st, t)); kids.set(t, [t.ab.a, t.ab.b]); kids.get('st').push(t); });
  nodes.forEach(n => {
    box.set(n, tNode(st, n)); kids.set(n, []);
    if (!n.in.length && !n.out.length) kids.get('st').push(n);
  });
  return { box, kids, levels: [['st'], trees, nodes].filter(l => l.length) };
}

function tLayout({ box, levels }) {
  const at = new Map();
  levels.forEach(l => l.forEach(k => {
    const { label, lines } = box.get(k), chars = Math.max(label.length, ...lines.map(s => s.replace(/^~/, '').length));
    at.set(k, { w: chars * T_CH + 24, h: T_HEAD + lines.length * T_LINE + 8 });
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

function tSvg(st) {
  const g = tGraph(st), { at, W, H } = tLayout(g), parts = [];
  g.kids.forEach((cs, k) => cs.forEach(c => {
    const from = [...g.kids].filter(([, ks]) => ks.includes(c)).map(([p]) => p).sort((p, q) => at.get(p).x - at.get(q).x);
    const off = (from.indexOf(k) - (from.length - 1) / 2) * 16, p = at.get(k), n = at.get(c);
    parts.push(`<path class="edge" d="M${p.x} ${p.y + p.h} L${n.x + off} ${n.y}"/>`);
  }));
  at.forEach((n, k) => {
    const { label, lines } = g.box.get(k), left = n.x - n.w / 2;
    parts.push(`<rect class="box" x="${left}" y="${n.y}" width="${n.w}" height="${n.h}" rx="${g.kids.get(k).length ? 8 : 14}"/>`);
    parts.push(`<text class="name" x="${n.x}" y="${n.y + T_HEAD / 2 + 2}">${tEsc(label)}</text>`);
    lines.forEach((s, i) => {
      const calc = s.startsWith('~');
      parts.push(`<text class="l data${calc ? ' calc' : ''}" x="${left + 12}" y="${n.y + T_HEAD + i * T_LINE + T_LINE / 2}">${tEsc(calc ? s.slice(1) : s)}</text>`);
    });
  });
  return { view: `-8 -8 ${W + 16} ${H + 16}`, width: W + 16, body: parts.join('') };
}

function tDraw() {
  const { st, from } = tLoad(), svg = document.getElementById('tree'), { view, width, body } = tSvg(st);
  document.getElementById('source').textContent = `from ${from}, read at ${new Date().toLocaleTimeString()}`;
  svg.setAttribute('viewBox', view);
  svg.style.width = `${width}px`;
  svg.innerHTML = body;
}

window.addEventListener('storage', e => { if (e.key === T_KEY) tDraw(); });
tDraw();
