const cNS = 'http://www.w3.org/2000/svg';

function cAdd(parent, tag, attrs, text) {
  const el = document.createElementNS(cNS, tag);
  Object.entries(attrs).forEach(([k, v]) => el.setAttribute(k, v));
  if (text !== undefined) el.textContent = text;
  parent.append(el);
  return el;
}

function cBox(svg, x, y, w, h, cls, lines) {
  cAdd(svg, 'rect', { class: `box ${cls}`, x, y, width: w, height: h, rx: 6 });
  const top = lines.length === 1 ? [h / 2] : lines.length === 2 ? [h / 2 - 9, h / 2 + 9] : [20, 40, 58];
  lines.forEach(([text, c], i) => cAdd(svg, 'text', { x: x + w / 2, y: y + top[i], class: c }, text));
}

const C_FILES = [
  [320, 10, 360, 'builder.js', 'uses every file below + atob-draw.js', 'the page: panels, methods, canvas, save', 'page-file'],
  [320, 98, 360, 'panel-lists.js', 'panel-edges direction place graph text vec', 'step list, nodes table, read-only list'],
  [10, 186, 230, 'preset.js', 'move graph', 'the three-node example'],
  [385, 186, 230, 'panel-edges.js', 'move place graph vec', 'edge picker, increments, move all'],
  [750, 186, 230, 'panel-seqs.js', 'save', 'sequence picker, new, delete, undo'],
  [10, 274, 230, 'move.js', 'direction place text vec', 'place a node, re-measure the rest'],
  [255, 274, 230, 'methods.js', 'direction place text vec', 'a→b card and pair card, step by step'],
  [500, 274, 230, 'draw.js', 'direction place graph vec', 'axes, edge arrows, φ and θ arcs'],
  [750, 274, 230, 'save.js', 'graph', 'load, pack, undo, send to the server'],
  [385, 362, 230, 'direction.js', 'place graph vec', 'start + Δ, measure, lines to draw'],
  [385, 450, 230, 'place.js', 'vec', 'index ↔ place, an edge\'s vector'],
  [140, 538, 230, 'vec.js', 'nothing', 'add, subtract, length, polar, wrap'],
  [385, 538, 230, 'graph.js', 'nothing', 'link and drop edges, the open step'],
  [630, 538, 230, 'text.js', 'nothing', 'numbers and points as text'],
];

function cDrawFiles(svg) {
  C_FILES.forEach(([x, y, w, name, uses, note, cls = '']) =>
    cBox(svg, x, y, w, 70, cls, [[name, 'name'], [uses, 'uses'], [note, 'note']]));
}

const C_LANES = [
  ['you', 'name'], ['panel-edges.js'], ['move.js'], ['builder.js', 'file', 'page-file'], ['save.js'],
  ['docs server', 'file', 'ext', 'docs-open.ts'], ['sequences.js', 'file', 'data'], ['a→b tree tab', 'name'],
];
const cLane = i => 63 + 125 * i;

const C_STEPS = [
  ['call', 0, 1, 92, '1 press + under move all'],
  ['loop', 1, 132, ['2 each picked edge:', 'Δ = wrap(Δ + increment)']],
  ['call', 1, 2, 200, '3 bMoveAll(picked)'],
  ['loop', 2, 236, ['4 bPut: place each picked', 'edge\'s second node', 'bRemeasure: rewrite the', 'unpicked edges\' start']],
  ['call', 1, 3, 326, '5 changed()'],
  ['loop', 3, 360, ['6 setup(): rebuild panels', 'update(): run methods,', 'draw the canvas, save()']],
  ['call', 3, 4, 434, '7 bKeep'],
  ['loop', 4, 466, ['8 pack to JSON,', 'last state → undo list']],
  ['call', 4, 5, 526, '9 POST /save'],
  ['call', 5, 6, 570, '10 write .tmp, rename'],
  ['call', 7, 6, 622, '11 reload every 1 s'],
  ['loop', 7, 640, ['12 redraw if it changed'], -1],
];

function cDrawEdit(svg) {
  const g = cAdd(svg, 'g', { class: 'lane' });
  C_LANES.forEach(([name, c = 'file', box = '', sub], i) =>
    cBox(g, cLane(i) - 55, 10, 110, 44, box, sub ? [[name, c], [sub, 'note small']] : [[name, c]]));
  cAdd(svg, 'path', { class: 'life', d: C_LANES.map((_, i) => `M ${cLane(i)} 54 V 680`).join(' ') });
  C_STEPS.forEach(([kind, ...s]) => {
    if (kind === 'call') {
      const [from, to, y, label] = s, a = cLane(from), b = cLane(to);
      cAdd(svg, 'path', { class: 'edge', d: `M ${a} ${y} H ${b - 2 * Math.sign(b - a)}` });
      cAdd(svg, 'text', { x: (a + b) / 2, y: y - 10, class: 'msg' }, label);
    } else {
      const [lane, y, lines, side = 1] = s, x = cLane(lane);
      cAdd(svg, 'path', { class: 'edge', d: `M ${x} ${y} h ${36 * side} v 26 h ${-34 * side}` });
      lines.forEach((t, i) => cAdd(svg, 'text', { x: x + 44 * side, y: y + 8 + 16 * i, class: `msg ${side > 0 ? 'l' : 'r'}` }, t));
    }
  });
}

cDrawFiles(document.getElementById('code-files'));
cDrawEdit(document.getElementById('code-edit'));
