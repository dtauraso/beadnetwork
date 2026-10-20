const B_KEY = 'pair-builder-v6';

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

function bSeqPicker(builder, h, open) {
  const load = text => { const s = JSON.parse(text); bWire(s); open(s); };
  const go = i => { builder.open = i; load(JSON.stringify(builder, bPack)); };
  bUndo = h('button', { textContent: `undo (${B_PAST.length})`, disabled: !B_PAST.length });
  bUndo.onclick = () => { bNow = B_PAST.pop(); load(bNow); };
  const pick = h('select', {}, builder.seqs.map((q, i) => h('option', { value: i, textContent: q.name, selected: i === builder.open })));
  pick.onchange = () => go(+pick.value);
  const add = h('button', { textContent: '+ new sequence' });
  add.onclick = () => {
    const q = JSON.parse(JSON.stringify(builder.seqs[builder.open], bPack));
    q.steps.forEach(bWireStep);
    builder.seqs.push({ ...q, name: `sequence ${builder.seqs.length + 1}` });
    go(builder.seqs.length - 1);
  };
  const del = h('button', { textContent: 'delete sequence', disabled: builder.seqs.length < 2 });
  del.onclick = () => {
    builder.seqs.splice(builder.open, 1);
    bSeqNumber(builder);
    go(Math.max(0, builder.open - 1));
  };
  return h('div', { className: 'row' }, [h('label', {}, ['sequence ', pick]), add, del, bUndo]);
}
