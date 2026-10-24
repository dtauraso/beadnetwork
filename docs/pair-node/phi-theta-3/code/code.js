const cNS = 'http://www.w3.org/2000/svg';
const C_W = 200, C_H = 70, C_GX = 50, C_GY = 14, C_PAD = 16;

function cAdd(parent, tag, attrs, text) {
  const el = document.createElementNS(cNS, tag);
  Object.entries(attrs).forEach(([k, v]) => el.setAttribute(k, v));
  if (text !== undefined) el.textContent = text;
  parent.append(el);
  return el;
}

function cPlaces(cols) {
  const at = {};
  cols.forEach((col, i) => col.forEach((id, j) => {
    if (id) at[id] = { x: C_PAD + i * (C_W + C_GX), y: C_PAD + j * (C_H + C_GY) };
  }));
  return at;
}

function cEdge(g, a, b, fy, ty) {
  if (a.x === b.x && Math.abs(b.y - a.y) > C_H + C_GY) {
    const x = a.x + C_W;
    return cAdd(g, 'path', { class: 'edge', d: `M ${x} ${a.y + C_H / 2} h 18 V ${b.y + C_H / 2} H ${x + 2}` });
  }
  if (a.x === b.x) {
    const x = a.x + C_W / 2, down = b.y > a.y;
    const y1 = down ? a.y + C_H : a.y, y2 = down ? b.y - 2 : b.y + C_H + 2;
    return cAdd(g, 'path', { class: 'edge', d: `M ${x} ${y1} V ${y2}` });
  }
  const x1 = a.x + C_W, y1 = a.y + fy, x2 = b.x - 2, y2 = b.y + ty, k = C_GX / 1.4;
  cAdd(g, 'path', { class: 'edge', d: `M ${x1} ${y1} C ${x1 + k} ${y1} ${x2 - k} ${y2} ${x2} ${y2}` });
}

const cSpread = (list, i) => C_H * (i + 1) / (list.length + 1);

function cTrunk(g, a, targets) {
  const xm = a.x + C_W + C_GX / 2, ys = targets.map(([b, ty]) => b.y + ty), y0 = a.y + C_H / 2;
  cAdd(g, 'path', { class: 'trunk', d: `M ${a.x + C_W} ${y0} H ${xm} M ${xm} ${Math.min(y0, ...ys)} V ${Math.max(y0, ...ys)}` });
  targets.forEach(([b], i) => cAdd(g, 'path', { class: 'edge', d: `M ${xm} ${ys[i]} H ${b.x - 2}` }));
}

function cBox(g, id, { x, y }) {
  const [file, role, kind = 'fn', name = id] = C_FN[id];
  cAdd(g, 'rect', { class: `box ${kind}`, x, y, width: C_W, height: C_H, rx: 6 });
  cAdd(g, 'text', { x: x + 12, y: y + 19, class: 'name' }, name);
  cAdd(g, 'text', { x: x + 12, y: y + 40, class: 'role' }, role);
  cAdd(g, 'text', { x: x + 12, y: y + 58, class: 'file' }, file);
}

function cMap(svg, { cols, calls }) {
  const at = cPlaces(cols);
  const rows = Math.max(...cols.map(c => c.length));
  const w = 2 * C_PAD + 20 + cols.length * C_W + (cols.length - 1) * C_GX, h = 2 * C_PAD + rows * (C_H + C_GY) - C_GY;
  Object.entries({ viewBox: `0 0 ${w} ${h}`, width: w, height: h }).forEach(([k, v]) => svg.setAttribute(k, v));
  const edges = cAdd(svg, 'g', {}), boxes = cAdd(svg, 'g', {});
  const pairs = calls.flatMap(line => {
    const [from, to] = line.split('>').map(s => s.trim());
    return to.split(/\s+/).map(t => [from, t]).filter(([f, t]) => at[f].x !== at[t].x);
  });
  const ends = (side, other) => id => pairs.filter(p => p[side] === id).sort((p, q) => at[p[other]].y - at[q[other]].y);
  const outs = ends(0, 1), ins = ends(1, 0);
  const into = (from, t) => { const i = ins(t); return cSpread(i, i.findIndex(p => p[0] === from)); };
  calls.forEach(line => {
    const [from, to] = line.split('>').map(s => s.trim()), ts = to.split(/\s+/), a = at[from];
    if (ts.length >= 3 && ts.every(t => at[t].x === a.x + C_W + C_GX)) return cTrunk(edges, a, ts.map(t => [at[t], into(from, t)]));
    const o = outs(from);
    ts.forEach(t => cEdge(edges, a, at[t], cSpread(o, o.findIndex(p => p[1] === t)), into(from, t)));
  });
  Object.entries(at).forEach(([id, p]) => cBox(boxes, id, p));
}

Object.entries(C_MAPS).forEach(([id, map]) => cMap(document.getElementById(id), map));
