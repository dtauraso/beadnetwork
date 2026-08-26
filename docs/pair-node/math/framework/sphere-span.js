const SPAN_HEAD_LEN = 13;

function spanArrow(g, from, to, lineCls, headCls) {
  const dx = to[0] - from[0], dy = to[1] - from[1];
  const len = Math.hypot(dx, dy);
  if (len < 1e-9) return;
  const ux = dx / len, uy = dy / len;

  const back = [to[0] - ux * SPAN_HEAD_LEN, to[1] - uy * SPAN_HEAD_LEN];
  const wing = SPAN_HEAD_LEN * 0.42;
  g.appendChild(tag('line', {
    x1: from[0], y1: from[1], x2: back[0], y2: back[1], class: lineCls,
  }));
  g.appendChild(tag('polygon', {
    points: [
      `${to[0]},${to[1]}`,
      `${back[0] - uy * wing},${back[1] + ux * wing}`,
      `${back[0] + uy * wing},${back[1] - ux * wing}`,
    ].join(' '),
    class: headCls,
  }));
}

function sphereSpan(g, view, c, origins) {
  const ends = origins.map((o) => viewProject(view, c, o));
  spanArrow(g, ends[0], ends[1], 'ring-span', 'ring-span-head');
  spanArrow(g, ends[1], ends[0], 'ring-span', 'ring-span-head');
}

function sphereReach(g, view, c, origins, anchorIndex) {
  const from = origins[anchorIndex];
  const seated = origins[1 - anchorIndex];
  const step = {
    x: seated.x - from.x,
    y: seated.y - from.y,
    z: seated.z - from.z,
  };
  const tip = { x: seated.x + step.x, y: seated.y + step.y, z: seated.z + step.z };
  spanArrow(g, viewProject(view, c, seated), viewProject(view, c, tip),
    'ring-reach', 'ring-reach-head');
}
