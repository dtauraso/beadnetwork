const SPAN_HEAD_LEN = 13;

function spanArrow(g, from, to) {
  const dx = to[0] - from[0], dy = to[1] - from[1];
  const len = Math.hypot(dx, dy);
  if (len < 1e-9) return;
  const ux = dx / len, uy = dy / len;

  const back = [to[0] - ux * SPAN_HEAD_LEN, to[1] - uy * SPAN_HEAD_LEN];
  const wing = SPAN_HEAD_LEN * 0.42;
  g.appendChild(tag('line', {
    x1: from[0], y1: from[1], x2: back[0], y2: back[1], class: 'ring-span',
  }));
  g.appendChild(tag('polygon', {
    points: [
      `${to[0]},${to[1]}`,
      `${back[0] - uy * wing},${back[1] + ux * wing}`,
      `${back[0] + uy * wing},${back[1] - ux * wing}`,
    ].join(' '),
    class: 'ring-span-head',
  }));
}

function sphereSpan(g, view, c, origins) {
  const ends = origins.map((o) => viewProject(view, c, o));
  spanArrow(g, ends[0], ends[1]);
  spanArrow(g, ends[1], ends[0]);
}
