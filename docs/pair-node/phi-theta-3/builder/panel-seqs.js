function bSeqPicker(builder, h, open) {
  const load = text => open(BBuilder.from(JSON.parse(text)));
  const go = i => { builder.open = i; load(JSON.stringify(builder)); };
  bUndo = h('button', { textContent: `undo (${B_PAST.length})`, disabled: !B_PAST.length });
  bUndo.onclick = () => { bNow = B_PAST.pop(); load(bNow); };
  const pick = h('select', {}, builder.seqs.map((q, i) => h('option', { value: i, textContent: q.name, selected: i === builder.open })));
  pick.onchange = () => go(+pick.value);
  const add = h('button', { textContent: '+ new sequence' });
  add.onclick = () => go(builder.addSeq());
  const del = h('button', { textContent: 'delete sequence', disabled: builder.seqs.length < 2 });
  del.onclick = () => go(builder.dropSeq());
  return h('div', { className: 'row' }, [h('label', {}, ['sequence ', pick]), add, del, bUndo]);
}
