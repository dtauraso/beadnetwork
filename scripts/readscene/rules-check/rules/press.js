function stepDelta(link, axis, sign, s) {
  const delta = { ...link.delta };
  delta[axis] = delta[axis] + sign * link.increment[axis];
  if (axis === "r") {
    if (link.start.r + delta.r < 0) return null;
  } else {
    delta[axis] = wrap(delta[axis], turn(s));
  }
  return delta;
}

function pressLink(link, button, s) {
  const axes = button.k === "both" ? AXES : [button.k];
  let delta = link.delta;
  for (const axis of axes) {
    const lock = button.k === "both" && link.lock ? link.lock[axis] : 0;
    const next = stepDelta({ ...link, delta }, axis, button.sign * (lock || 1), s);
    if (next !== null) delta = next;
  }
  return delta;
}

function predict(before, presses) {
  const nodes = {};
  for (const n of before.nodes) nodes[n.id] = n;
  const stepped = {};
  for (const link of before.links) {
    const to = nodes[link.to];
    if (!link.picked || !to || stepped[link.to]) continue;
    let delta = link.delta;
    for (const button of presses) delta = pressLink({ ...link, delta }, button, to.s);
    stepped[link.to] = { ...link, delta };
  }
  const links = before.links.map(l => stepped[l.to] && stepped[l.to].from === l.from ? stepped[l.to] : l);
  const positions = {};
  const find = id => {
    if (id in positions) return positions[id];
    const n = nodes[id];
    positions[id] = n && n.place ? nodePosition(before.frame, n.place) : null;
    const link = stepped[id];
    if (link) {
      const a = find(link.from);
      if (a) positions[id] = add(a, linkVector(before.frame, link, n.s));
    }
    return positions[id];
  };
  for (const n of before.nodes) find(n.id);
  return { links, positions };
}

function compare(before, after, presses) {
  const expected = predict(before, presses);
  const got = checkSequence(after);
  const links = expected.links.map((l, i) => {
    const r = after.links.find(a => a.from === l.from && a.to === l.to);
    const same = Boolean(r) && sameTicks(r.delta, l.delta);
    return { name: l.from + "→" + l.to, was: before.links[i].delta, expected: l.delta, recorded: r ? r.delta : null, same };
  });
  const nodes = before.nodes.map(n => {
    const e = expected.positions[n.id];
    const r = got.positions[n.id] || null;
    const gap = e && r ? distance(e, r) : null;
    return { id: n.id, expected: e, recorded: r, gap, same: e && r ? samePlace(e, r) : !e && !r };
  });
  const closes = got.links.every(row => row.gap === undefined || row.closes);
  return { links, nodes, closes, ok: closes && links.every(l => l.same) && nodes.every(n => n.same) };
}
