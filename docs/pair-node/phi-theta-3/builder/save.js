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
  bFileWrite();
}

let bFile = null, bFileOn = false, bFileSent = null, bFileBtn = null, bFileErr = '', bFileQueue = Promise.resolve();

function bFileLabel() {
  if (!window.showSaveFilePicker) return 'download file';
  if (!bFile) return 'save to file';
  return bFileOn ? `saving to ${bFile.name}` : `resume saving to ${bFile.name}`;
}

function bFileShow() {
  if (!bFileBtn) return;
  bFileBtn.textContent = bFileLabel();
  bFileBtn.title = bFileErr || (bFileOn ? 'every change is written to this file; click to choose another file' : 'write every change to a file you choose, such as builder/sequences.json');
  bFileBtn.classList.toggle('bad', !!bFileErr);
}

function bFileWrite() {
  if (!bFile || !bFileOn || bNow === null || bNow === bFileSent) return;
  const text = JSON.stringify(JSON.parse(bNow), null, 2) + '\n', file = bFile;
  bFileSent = bNow;
  bFileQueue = bFileQueue.then(async () => {
    const out = await file.createWritable();
    await out.write(text);
    await out.close();
    bFileErr = '';
  }).catch(e => { bFileErr = `could not write ${file.name}: ${e.message}`; bFileOn = false; bFileSent = null; }).then(bFileShow);
}

function bFileStore(handle) {
  return new Promise(done => {
    try {
      const req = indexedDB.open('pair-builder-file', 1);
      req.onupgradeneeded = () => req.result.createObjectStore('file');
      req.onsuccess = () => {
        const tx = req.result.transaction('file', 'readwrite'), store = tx.objectStore('file');
        const get = handle ? store.put(handle, 'file') : store.get('file');
        get.onsuccess = () => done(handle || get.result || null);
        get.onerror = () => done(null);
      };
      req.onerror = () => done(null);
    } catch { done(null); }
  });
}

async function bFileClick() {
  if (!window.showSaveFilePicker) {
    const a = document.createElement('a');
    a.href = URL.createObjectURL(new Blob([JSON.stringify(JSON.parse(bNow), null, 2) + '\n'], { type: 'application/json' }));
    a.download = 'sequences.json';
    a.click();
    return;
  }
  try {
    if (bFile && !bFileOn && await bFile.requestPermission({ mode: 'readwrite' }) === 'granted') bFileOn = true;
    else {
      bFile = await window.showSaveFilePicker({ suggestedName: 'sequences.json', types: [{ description: 'JSON', accept: { 'application/json': ['.json'] } }] });
      bFileOn = true;
      bFileStore(bFile);
    }
    bFileErr = '';
    bFileSent = null;
  } catch (e) {
    if (e.name !== 'AbortError') bFileErr = e.message;
  }
  bFileShow();
  bFileWrite();
}

if (window.showSaveFilePicker) bFileStore().then(async handle => {
  if (!handle) return;
  bFile = handle;
  try { bFileOn = await handle.queryPermission({ mode: 'readwrite' }) === 'granted'; } catch { }
  bFileShow();
  bFileWrite();
});
