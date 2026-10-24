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
    bWirePicked(q);
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
