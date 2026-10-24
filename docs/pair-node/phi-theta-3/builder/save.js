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
    const builder = window.B_SAVED ? JSON.parse(JSON.stringify(window.B_SAVED)) : bCarry();
    if (builder && builder.seqs[builder.open]) { bWire(builder); return bSeqNumber(builder); }
  } catch { }
  return null;
}

function bCarry() {
  try { return JSON.parse(localStorage.getItem(B_KEY)); } catch { return null; }
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
  bSend();
}

const B_FILE = 'phi-theta-3/builder/sequences.js';
let bSent = null, bSending = Promise.resolve();

function bSend() {
  const at = window.BEADNETWORK_DOCS_OPEN;
  if (!at || bNow === bSent) return;
  const text = `window.B_SAVED = ${JSON.stringify(JSON.parse(bNow), null, 2)};\n`;
  bSent = bNow;
  bSending = bSending.then(() => fetch(`http://localhost:${at.port}/save?token=${encodeURIComponent(at.token)}&file=${encodeURIComponent(B_FILE)}`, { method: 'POST', body: text }))
    .then(r => { if (!r.ok) throw new Error(`status ${r.status}`); })
    .catch(e => { bSent = null; console.warn(`builder: ${B_FILE} not saved: ${e.message}`); });
}
