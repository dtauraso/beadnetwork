const B_SEQS = 'pair-builder-seqs-v1';

function bSeqBook() {
  try { const b = JSON.parse(localStorage.getItem(B_SEQS)); if (b && b.seqs.length && b.seqs[b.at]) return bSeqNumber(bSeqClean(b)); } catch { }
  return { seqs: [{ name: 'sequence 1' }], at: 0 };
}

function bSeqClean(book) {
  book.seqs.forEach(q => {
    if (!q.data) return;
    const s = JSON.parse(q.data);
    bWire(s);
    q.data = JSON.stringify(s, bPack);
  });
  try { localStorage.setItem(B_SEQS, JSON.stringify(book)); } catch { }
  return book;
}

function bSeqNumber(book) {
  book.seqs.forEach((q, i) => { if (/^sequence \d+$/.test(q.name)) q.name = `sequence ${i + 1}`; });
  return book;
}

const B_PAST = [];
let bNow = null, bUndo = null;

function bSeqKeep(book, st) {
  book.seqs[book.at].data = JSON.stringify(st, bPack);
  const now = JSON.stringify(book);
  if (bNow !== null && now !== bNow) { B_PAST.push(bNow); if (B_PAST.length > 100) B_PAST.shift(); }
  bNow = now;
  if (bUndo) { bUndo.disabled = !B_PAST.length; bUndo.textContent = `undo (${B_PAST.length})`; }
  try { localStorage.setItem(B_SEQS, now); } catch { }
}

function bSeqPicker(book, h, open) {
  const go = i => { book.at = i; const s = JSON.parse(book.seqs[i].data); bWire(s); open(s); };
  bUndo = h('button', { textContent: `undo (${B_PAST.length})`, disabled: !B_PAST.length });
  bUndo.onclick = () => { bNow = B_PAST.pop(); Object.assign(book, JSON.parse(bNow)); go(book.at); };
  const pick = h('select', {}, book.seqs.map((q, i) => h('option', { value: i, textContent: q.name, selected: i === book.at })));
  pick.onchange = () => go(+pick.value);
  const add = h('button', { textContent: '+ new sequence' });
  add.onclick = () => { book.seqs.push({ name: `sequence ${book.seqs.length + 1}`, data: book.seqs[book.at].data }); go(book.seqs.length - 1); };
  const del = h('button', { textContent: 'delete sequence', disabled: book.seqs.length < 2 });
  del.onclick = () => {
    book.seqs.splice(book.at, 1);
    bSeqNumber(book);
    go(Math.max(0, book.at - 1));
  };
  return h('div', { className: 'row' }, [h('label', {}, ['sequence ', pick]), add, del, bUndo]);
}
