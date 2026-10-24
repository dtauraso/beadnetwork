const T_HEAD = 30, T_LINE = 17, T_CH = 7.2, T_GX = 18, T_GY = 44;
const tNum = v => `${+v.toFixed(3)}`;
const tAng = (o, ks = ['phi', 'theta', 'r']) => (o ? ks.filter(k => k in o).map(k => `${{ phi: 'φ', theta: 'θ', r: 'r' }[k]} ${tNum(o[k])}`).join(' | ') : '—');
const tNames = (ts, name = x => x.name) => (ts.length ? ts.map(name).join(', ') : '—');
const tOn = o => tNames(Object.keys(o || {}).filter(k => o[k]), k => k);
const tEsc =s => s.replace(/&/g, '&amp;').replace(/</g, '&lt;');

function tLoad() {
  const builder = BBuilder.load();
  if (builder) return { builder, from: `the builder's save, ${B_FILE}` };
  return { builder: bPreset(), from: `the three-node preset — the builder has not written ${B_FILE} yet` };
}

function tState(builder) {
  return { kind: 'builder', label: 'builder', lines: [['s', tNum(builder.s)], ['ρ', tNum(builder.rho)], ['T_φ', tNum(builder.Tphi)], ['T_θ', tNum(builder.Ttheta)], ['c_r', tNum(builder.cr)],
    ['made', `${builder.made}`], ['sequences', tNames(builder.seqs)], ['open', builder.openSeq().name, 'ref']] };
}

const tLock = o => (o.lock && (o.lock.phi || o.lock.theta) ? tAng(o.lock, ['phi', 'theta']) : '—');

function tSeq(seq) {
  return { kind: 'seq', label: seq.name, lines: [['steps', `${seq.steps.length}`]] };
}

function tStep(st, i) {
  return { kind: 'step', label: `step ${i + 1}`, lines: [['cards', 'method cards', 'ref'], ['edges', tNames([...st.edges]), 'ref'], ['nodes', tNames([...st.nodes]), 'ref'],
    ['picked', tNames(st.picked), 'ref'], ['all lock', tLock(st)], ['all locking', st.locking ? 'on' : 'off'],
    ...tWorked([['~recalculated', tNames(st.recalc()), 'ref']])] };
}

function tCards(c) {
  return { kind: 'cards', label: 'method cards', lines: [['item', c.item ? c.item.name : '—', 'ref'], ['methods', tOn(c.methods)], ['card step', `${c.card}`]] };
}

const tWorked = rows => [['', 'not stored', 'note'], ...rows];

function tEdge(builder, e) {
  return { kind: 'edge', label: `edge ${e.name}`, lines: [['from', e.from.name, 'ref'], ['to', e.to.name, 'ref'], ['start', tAng(e.start)], ['Δ', tAng(e.delta, ['phi', 'theta'])],
    ['increment', tAng(e.increment, ['phi', 'theta'])], ['lock', tLock(e)], ['locking', e.locking ? 'on' : 'off'],
    ...tWorked([['~start + Δ', tAng(e.dir(builder))], [`~measured C_${e.from.name} → C_${e.to.name}`, tAng(e.measure(builder))]])] };
}

function tNode(builder, n) {
  const C = n.place(builder);
  return { kind: 'node', label: `node ${n.name}`, lines: [['s', n.s ? tNum(n.s) : `— (the scene's s, ${tNum(builder.s)})`],
    ['at', n.at.from ? `from ${n.at.from.name}`: 'from O', n.at.from ? 'ref' : undefined], ['', tAng(n.at)],
    ['out', tNames(n.out), 'ref'], ['in', tNames(n.in), 'ref'],
    ...tWorked([['~index on O\'s grid', tAng(n.index(builder), ['r', 'phi', 'theta'])], ['~C', ['x', 'y', 'z'].map((k, i) => `${k} ${tNum(C[i])}`).join(' | ')]])] };
}

function tGraph(builder) {
  const seq = builder.openSeq(), parts = [], nodes = [];
  const box = new Map([['builder', tState(builder)], ['seq', tSeq(seq)]]), kids = new Map([['builder', ['seq']], ['seq', seq.steps]]);
  const add = (k, b, cs = []) => { box.set(k, b); kids.set(k, cs); };
  seq.steps.forEach((st, i) => {
    const edges = [...st.edges], loose = [...st.nodes].filter(n => !n.in.length && !n.out.length);
    add(st, tStep(st, i), [st.cards, ...edges, ...loose]);
    add(st.cards, tCards(st.cards));
    edges.forEach(e => add(e, tEdge(builder, e), [e.from, e.to]));
    st.nodes.forEach(n => add(n, tNode(builder, n)));
    parts.push(st.cards, ...edges);
    nodes.push(...st.nodes);
  });
  return { box, kids, levels: [['builder'], ['seq'], seq.steps, parts, nodes].filter(l => l.length) };
}

function tLayout({ box, levels }) {
  const at = new Map();
  levels.forEach(l => l.forEach(k => {
    const { label, lines, steps = [{ cols: [lines] }] } = box.get(k);
    let top = T_HEAD;
    const bands = steps.map(({ name, cols }) => {
      let x = 0;
      const cs = cols.map(lines => {
        const kw = Math.max(...lines.map(([n]) => n.replace(/^~/, '').length)) + 2;
        const w = Math.max(...lines.map(([, v, kind]) => (kind === 'note' || kind === 'part' ? 0 : kw) + v.length)) * T_CH + 24;
        x += w;
        return { lines, x: x - w, w, kw: kw * T_CH };
      });
      const head = name ? T_LINE + 6 : 0, h = head + Math.max(...cols.map(c => c.length)) * T_LINE;
      top += h;
      return { name, top: top - h, head, h, w: x, cs };
    });
    at.set(k, { w: Math.max(...bands.map(b => b.w), label.length * T_CH + 24), h: top + 8, bands });
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
    const { label, kind } = g.box.get(k), box = n.x - n.w / 2;
    parts.push(`<rect class="box k-${kind}" x="${box}" y="${n.y}" width="${n.w}" height="${n.h}" rx="${g.kids.get(k).length ? 8 : 14}"/>`);
    parts.push(`<text class="name" x="${n.x}" y="${n.y + T_HEAD / 2 + 2}">${tEsc(label)}</text>`);
    n.bands.forEach(({ name: step, top, head, h, cs }, b) => {
      const ceil = n.y + top + head - 3, last = b === n.bands.length - 1;
      if (step) {
        parts.push(`<text class="l data part" x="${box + 12}" y="${n.y + top + T_LINE / 2}">${tEsc(step)}</text>`);
        parts.push(`<line class="steprule" x1="${box + 8}" x2="${box + n.w - 8}" y1="${ceil}" y2="${ceil}"/>`);
      }
      cs.forEach(({ lines, x, w, kw }, c) => {
        const left = box + x, right = c === cs.length - 1 ? box + n.w : left + w;
        const y1 = step ? ceil + 8 : n.y + 8, y2 = last ? n.y + n.h - 8 : n.y + top + h, dots = Math.round((y2 - y1) / 6), gap = (y2 - y1) / dots;
        const title = !step && Math.abs(left - n.x) < label.length * 4.5 + 6;
        if (c) for (let d = 0; d <= dots; d++) {
          const y = y1 + d * gap;
          if (!(title && y < n.y + T_HEAD - 2)) parts.push(`<circle class="rule" cx="${left}" cy="${y}" r="1.25"/>`);
        }
        lines.forEach(([name, v, kind], i) => {
          const calc = name.startsWith('~'), ref = kind === 'ref', y = n.y + top + head + i * T_LINE + T_LINE / 2;
          if (kind === 'note') {
            parts.push(`<line class="split" x1="${left + 8}" x2="${right - 8}" y1="${y - T_LINE / 2}" y2="${y - T_LINE / 2}"/>`);
            parts.push(`<text class="l data note" x="${left + 12}" y="${y}">${tEsc(v)}</text>`);
            return;
          }
          if (kind === 'part') {
            parts.push(`<text class="l data part" x="${left + 12}" y="${y}">${tEsc(v)}</text>`);
            return;
          }
          parts.push(`<text class="l data${calc ? ' calc' : ''}" y="${y}"><tspan class="k" x="${left + 12}">${tEsc(calc ? name.slice(1) : name)}</tspan>`
            + `<tspan class="v" x="${left + 12 + kw}">${ref && v !== '—' ? v.split(', ').map(r => `<tspan class="ref">${tEsc(r)}</tspan>`).join(', ') : tEsc(v)}</tspan></text>`);
        });
      });
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

function tReread() {
  const was = JSON.stringify(window.B_SAVED), s = document.createElement('script');
  s.src = `../builder/sequences.js?t=${Date.now()}`;
  s.onload = s.onerror = () => { s.remove(); if (JSON.stringify(window.B_SAVED) !== was) tDraw(); };
  document.head.append(s);
}

setInterval(tReread, 1000);
tDraw();
