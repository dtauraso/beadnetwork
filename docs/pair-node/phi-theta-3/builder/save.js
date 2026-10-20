const B_KEY = 'pair-builder-v6';

function bPack(k, v) {
  if (['gone', 'out', 'in'].includes(k)) return undefined;
  if (k === 'edges') return [...v].map(t => t.ab.row);
  if (k === 'picked') return v.map(t => `${t.ab.a.name}→${t.ab.b.name}`);
  if (v instanceof Set) return [...v];
  if (k === 'from' || k === 'to' || k === 'item') return v.name;
  return v;
}

function bWireStep(step) {
  const by = new Map(step.nodes.map(({ name, s, idx }) => [name, bMake({ name, s, idx })])), rows = step.edges.map(({ from, to, start, delta, increment }) => ({ from, to, start, delta, increment }));
  step.nodes = new Set(by.values());
  step.edges = new Set();
  rows.forEach(r => {
    r.from = by.get(r.from); r.to = by.get(r.to);
    if (r.from && r.to) bLink(step, r);
  });
}

function bWire(builder) {
  builder.seqs.forEach(q => q.steps.forEach(bWireStep));
  const step = bOpenStep(builder), by = new Map([...step.nodes].map(n => [n.name, n])), abs = new Map([...step.edges].map(t => [`${t.ab.a.name}→${t.ab.b.name}`, t]));
  builder.picked = (builder.picked || []).map(k => abs.get(k)).filter(Boolean);
  builder.methods ??= {};
  builder.item = by.get(builder.item) || bFirst(step.nodes);
  builder.made ??= 1 + Math.max(-1, ...builder.seqs.flatMap(q => q.steps.flatMap(st => [...st.nodes].map(n => bNameAt(n.name)))).filter(Number.isFinite));
}

function bLoad() {
  try {
    const builder = JSON.parse(localStorage.getItem(B_KEY)) || bOld();
    if (builder && builder.seqs[builder.open]) { bWire(builder); return bSeqNumber(builder); }
  } catch { }
  return null;
}

function bOld() {
  const get = k => JSON.parse(localStorage.getItem(k)), book = get('pair-builder-seqs-v1'), v5 = get('pair-builder-v5');
  const datas = book && book.seqs.length ? book.seqs.map(q => (q.data ? JSON.parse(q.data) : v5)) : [v5];
  const open = book && datas[book.at] ? book.at : 0;
  if (!datas[open]) return null;
  const { nodes, trees, ...rest } = datas[open];
  return { ...rest, open, seqs: datas.map((d, i) => ({ name: book ? book.seqs[i].name : 'sequence 1', steps: [{ edges: d ? d.trees : [], nodes: d ? d.nodes : [] }] })) };
}

function bSeqNumber(builder) {
  builder.seqs.forEach((q, i) => { if (/^sequence \d+$/.test(q.name)) q.name = `sequence ${i + 1}`; });
  return builder;
}

const B_PAST = [];
let bNow = null, bUndo = null;

function bKeep(builder) {
  const now = JSON.stringify(builder, bPack);
  if (bNow !== null && now !== bNow) { B_PAST.push(bNow); if (B_PAST.length > 100) B_PAST.shift(); }
  bNow = now;
  if (bUndo) { bUndo.disabled = !B_PAST.length; bUndo.textContent = `undo (${B_PAST.length})`; }
  try { localStorage.setItem(B_KEY, now); } catch { }
}
