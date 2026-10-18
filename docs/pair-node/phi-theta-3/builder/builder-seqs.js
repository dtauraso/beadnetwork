const B_SEQS = 'pair-builder-seqs-v1';

function bSeqBook() {
  try { const b = JSON.parse(localStorage.getItem(B_SEQS)); if (b && b.seqs.length && b.seqs[b.at]) return b; } catch { }
  return { seqs: [{ name: 'sequence 1' }], at: 0 };
}

function bSeqKeep(book, st) {
  book.seqs[book.at].data = JSON.stringify(st, bPack);
  try { localStorage.setItem(B_SEQS, JSON.stringify(book)); } catch { }
}

function bSeqPicker(book, h, open) {
  const go = i => { book.at = i; const s = JSON.parse(book.seqs[i].data); bWire(s); open(s); };
  const pick = h('select', {}, book.seqs.map((q, i) => h('option', { value: i, textContent: q.name, selected: i === book.at })));
  pick.onchange = () => go(+pick.value);
  const add = h('button', { textContent: '+ new sequence' });
  add.onclick = () => {
    const n = 1 + Math.max(0, ...book.seqs.map(q => +(q.name.match(/^sequence (\d+)$/) || [])[1] || 0));
    book.seqs.push({ name: `sequence ${n}`, data: book.seqs[book.at].data }); go(book.seqs.length - 1);
  };
  const del = h('button', { textContent: 'delete sequence', disabled: book.seqs.length < 2 });
  del.onclick = () => { if (confirm(`delete ${book.seqs[book.at].name}?`)) { book.seqs.splice(book.at, 1); go(Math.max(0, book.at - 1)); } };
  return h('div', { className: 'row' }, [h('label', {}, ['sequence ', pick]), add, del]);
}
