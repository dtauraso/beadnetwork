const T = {
  root: { label: 'root', kids: ['pair', 'bc'] },
  pair: { label: '(a→b, a→c)', kids: ['ab', 'ac'] },
  bc: { label: 'b→c', kids: ['b', 'c'] },
  ab: { label: 'a→b', kids: ['a', 'b'] },
  ac: { label: 'a→c', kids: ['a', 'c'] },
  a: { label: 'a', kids: [] },
  b: { label: 'b', kids: [] },
  c: { label: 'c', kids: [] },
};
const T_LEVELS = [['root'], ['pair', 'bc'], ['ab', 'ac'], ['a', 'b', 'c']];
const T_HEAD = 30, T_LINE = 17, T_CH = 7.2, T_GX = 18, T_GY = 44;

function tData() {
  const st = bPreset(), [ab, bc, ac] = bRows(st);
  bSetStart(st, ab, 'phi', 3);
  const t = ab.tree, p = t.pairs.find(q => q.ac.row === ac), [a, b, c] = [...st.nodes];
  const sym = { phi: 'φ', theta: 'θ', r: 'r' };
  const pt = (name, o, ks = ['phi', 'theta', 'r']) => `${name} ${ks.filter(k => k in o).map(k => `${sym[k]} ${bN(o[k])}`).join(' ')}`;
  const mat = m => [0, 3, 6].map(i => `  ${m.slice(i, i + 3).map(bF).join('  ')}`);
  const row = r => `row ${r.from.name}→${r.to.name} · ${bKind(r)}`;
  const idx = n => [pt('idx', n.idx, ['r', 'phi', 'theta'])];
  return {
    root: ['ab → the a→b node', 'pairs → (a→b, a→c)', 'bcs → b→c'],
    ab: [row(ab), pt('start', ab.start), pt('end', ab.end), pt('was', t.ab.was), pt('now', t.ab.now)],
    ac: [row(ac), pt('start', ac.start), 'rot', ...mat(p.ac.rot)],
    pair: [pt('base', p.base), pt('ref', p.ref), pt('held', p.held), pt('went', p.went), pt('step', p.step),
      `own step ${p.mine ? 'on' : 'off'}`, `automatic ${bAuto(p) ? 'on' : 'off'}`, '~reads a→b: was, now', '~reads a→c: kind'],
    bc: [row(t.bcs[t.pairs.indexOf(p)].row), pt('start', bc.start), pt('end', bc.end), 'turn', ...mat(bc.turn.m)],
    a: idx(a), b: idx(b), c: idx(c),
  };
}
const T_DATA = tData();

function tUnder(top) {
  const keep = new Set([top]);
  keep.forEach(k => T[k].kids.forEach(c => keep.add(c)));
  return keep;
}

function tLayout(keep) {
  const at = {}, levels = T_LEVELS.map(l => l.filter(k => keep.has(k))).filter(l => l.length);
  levels.forEach(l => l.forEach(k => {
    const lines = T_DATA[k], chars = Math.max(T[k].label.length, ...lines.map(s => s.length));
    at[k] = { w: chars * T_CH + 24, h: T_HEAD + lines.length * T_LINE + 8 };
  }));
  const span = l => l.reduce((s, k) => s + at[k].w, 0) + (l.length - 1) * T_GX, W = Math.max(...levels.map(span));
  let y = 0;
  levels.forEach(l => {
    let x = (W - span(l)) / 2;
    l.forEach(k => { at[k].x = x + at[k].w / 2; at[k].y = y; x += at[k].w + T_GX; });
    y += Math.max(...l.map(k => at[k].h)) + T_GY;
  });
  return { at, W, H: y - T_GY };
}

function tDraw(svg) {
  const top = svg.dataset.top, keep = tUnder(top), mark = svg.classList.contains('subtree'), parts = [];
  const { at, W, H } = tLayout(keep);
  keep.forEach(k => T[k].kids.forEach(c => {
    const from = [...keep].filter(p => T[p].kids.includes(c)).sort((p, q) => at[p].x - at[q].x);
    const off = (from.indexOf(k) - (from.length - 1) / 2) * 16;
    parts.push(`<path class="edge" d="M${at[k].x} ${at[k].y + at[k].h} L${at[c].x + off} ${at[c].y}"/>`);
  }));
  keep.forEach(k => {
    const n = at[k], cls = mark && k === top ? 'box top' : 'box', left = n.x - n.w / 2;
    parts.push(`<rect class="${cls}" x="${left}" y="${n.y}" width="${n.w}" height="${n.h}" rx="${T[k].kids.length ? 8 : 14}"/>`);
    parts.push(`<text class="name" x="${n.x}" y="${n.y + T_HEAD / 2 + 2}">${T[k].label}</text>`);
    T_DATA[k].forEach((s, i) => {
      const muted = s.startsWith('~');
      parts.push(`<text class="l small data${muted ? ' muted' : ''}" x="${left + 12}" y="${n.y + T_HEAD + i * T_LINE + T_LINE / 2}">${muted ? s.slice(1) : s}</text>`);
    });
  });
  svg.setAttribute('viewBox', `-8 -8 ${W + 16} ${H + 16}`);
  if (mark) svg.style.width = `${(W + 16) * 0.8}px`;
  svg.innerHTML = parts.join('');
}

document.querySelectorAll('svg[data-top]').forEach(tDraw);
