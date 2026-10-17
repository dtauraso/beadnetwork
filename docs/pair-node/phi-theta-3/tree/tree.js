const T = {
  root: { label: 'root', x: 300, y: 22, w: 90, kids: ['pair', 'bc'] },
  pair: { label: '(a→b, a→c)', x: 200, y: 132, w: 140, kids: ['ab', 'ac'] },
  bc: { label: 'b→c', x: 460, y: 132, w: 90, kids: ['b', 'c'] },
  ab: { label: 'a→b', x: 110, y: 242, w: 90, kids: ['a', 'b'] },
  ac: { label: 'a→c', x: 290, y: 242, w: 90, kids: ['a', 'c'] },
  a: { label: 'a', x: 150, y: 362, w: 44, kids: [] },
  b: { label: 'b', x: 300, y: 362, w: 44, kids: [] },
  c: { label: 'c', x: 450, y: 362, w: 44, kids: [] },
};
const T_H = 36;

function tUnder(top) {
  const keep = new Set([top]);
  keep.forEach(k => T[k].kids.forEach(c => keep.add(c)));
  return keep;
}

function tDraw(svg) {
  const top = svg.dataset.top, keep = tUnder(top), mark = svg.classList.contains('subtree'), parts = [];
  keep.forEach(k => T[k].kids.forEach(c => {
    const from = [...keep].filter(p => T[p].kids.includes(c)).sort((p, q) => T[p].x - T[q].x);
    const off = (from.indexOf(k) - (from.length - 1) / 2) * 16;
    parts.push(`<path class="edge" d="M${T[k].x} ${T[k].y + T_H} L${T[c].x + off} ${T[c].y}"/>`);
  }));
  keep.forEach(k => {
    const n = T[k], cls = mark && k === top ? 'box top' : 'box';
    parts.push(`<rect class="${cls}" x="${n.x - n.w / 2}" y="${n.y}" width="${n.w}" height="${T_H}" rx="${n.kids.length ? 8 : 18}"/>`);
    parts.push(`<text x="${n.x}" y="${n.y + T_H / 2}">${n.label}</text>`);
  });
  const ns = [...keep].map(k => T[k]);
  const x0 = Math.min(...ns.map(n => n.x - n.w / 2)) - 8, x1 = Math.max(...ns.map(n => n.x + n.w / 2)) + 8;
  const y0 = Math.min(...ns.map(n => n.y)) - 8, y1 = Math.max(...ns.map(n => n.y)) + T_H + 8;
  svg.setAttribute('viewBox', `${x0} ${y0} ${x1 - x0} ${y1 - y0}`);
  if (mark) svg.style.width = `${(x1 - x0) * 0.7}px`;
  svg.innerHTML = parts.join('');
}

document.querySelectorAll('svg[data-top]').forEach(tDraw);
