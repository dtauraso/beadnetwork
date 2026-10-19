function bEdgePicker(st, h, picked, redraw) {
  const trees = [...st.trees], abName = t => `${t.ab.a.name}→${t.ab.b.name}`;
  picked.splice(0, picked.length, ...picked.filter(t => trees.includes(t)));
  if (!trees.length) return [];
  const left = trees.filter(t => !picked.includes(t));
  const pick = h('select', {}, [h('option', { value: -1, textContent: left.length ? 'pick an edge' : 'all edges picked' }),
    ...left.map((t, i) => h('option', { value: i, textContent: abName(t) }))]);
  pick.onchange = () => { if (left[+pick.value]) picked.push(left[+pick.value]); redraw(); };
  const items = picked.map(t => {
    const drop = h('button', { textContent: 'unpick' });
    drop.onclick = () => { picked.splice(picked.indexOf(t), 1); redraw(); };
    return h('li', {}, [h('div', { className: 'row' }, [h('span', { className: 't', textContent: abName(t) }), drop])]);
  });
  return [h('div', { className: 'row' }, [h('label', {}, ['edge ', pick])]),
    items.length ? h('ul', { className: 'abs' }, items) : h('div', { className: 'hint', textContent: 'no edges picked' })];
}
