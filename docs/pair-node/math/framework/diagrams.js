const NS = 'http://www.w3.org/2000/svg';

function tag(name, attrs) {
  const e = document.createElementNS(NS, name);
  for (const k in attrs) e.setAttribute(k, attrs[k]);
  return e;
}

function svg(w, h) {
  return tag('svg', { viewBox: `0 0 ${w} ${h}`, width: '100%', role: 'img' });
}

function at(cx, cy, r, i, n) {
  const a = (i / n) * 2 * Math.PI - Math.PI / 2;
  return [cx + r * Math.cos(a), cy + r * Math.sin(a)];
}

function fold(d, m) {
  const u = ((d % m) + m) % m;
  return Math.min(u, m - u);
}

function head(cx, cy, rad, i, n, cls) {
  const a = (i / n) * 2 * Math.PI - Math.PI / 2;
  const dx = Math.cos(a), dy = Math.sin(a);
  const tip = [cx + rad * dx, cy + rad * dy];
  const back = rad - 9;
  const b1 = [cx + back * dx - 3.6 * dy, cy + back * dy + 3.6 * dx];
  const b2 = [cx + back * dx + 3.6 * dy, cy + back * dy - 3.6 * dx];
  return tag('polygon', { points: [tip, b1, b2].join(' '), class: cls });
}

function tauParts(k, n) {
  if (k === 0) return ['0', ''];
  let a = k, b = n;
  while (b) [a, b] = [b, a % b];
  const num = k / a, den = n / a;
  return [(num === 1 ? '' : num) + 'τ', den === 1 ? '' : String(den)];
}

function fraction(g, x, y, num, den) {
  const put = (s, dy, cls) => {
    const t = tag('text', { x: x, y: y + dy, class: cls });
    t.textContent = s;
    g.appendChild(t);
  };
  if (!den) {
    put(num, 4, 'ring-frac');
    return;
  }
  const w = 3.6 * Math.max(num.length, den.length) + 3;
  put(num, -2, 'ring-frac');
  g.appendChild(tag('line', { x1: x - w, y1: y + 1.5, x2: x + w, y2: y + 1.5, class: 'ring-bar' }));
  put(den, 12, 'ring-frac');
}

function ring(spec) {
  const n = spec.points, m = n / 2, S = spec.size || 240, c = S / 2, r = S * 0.29;
  const numOff = r + S * 0.055, labOff = r + S * 0.14;
  const g = svg(S, S);
  g.appendChild(tag('circle', { cx: c, cy: c, r: r, class: 'ring-line' }));

  for (let i = 0; i < n; i++) {
    const [x, y] = at(c, c, r, i, n);
    g.appendChild(tag('circle', { cx: x, cy: y, r: 3, class: 'ring-dot' }));
    if (spec.tau) {
      const [nx, ny] = at(c, c, numOff + 6, i, n);
      const [num, den] = tauParts(i, n);
      fraction(g, nx, ny, num, den);
    } else if (spec.numbers) {
      const [nx, ny] = at(c, c, numOff, i, n);
      const t = tag('text', { x: nx, y: ny + 3.5, class: 'ring-num' });
      t.textContent = i;
      g.appendChild(t);
    }
  }

  if (spec.marks) {
    g.appendChild(tag('circle', { cx: c, cy: c, r: 3, class: 'ring-mark' }));
    for (const m of spec.marks) {
      const [x, y] = at(c, c, r - 8, m.i, n);
      g.appendChild(tag('line', {
        x1: c, y1: c, x2: x, y2: y, class: m.end ? 'ring-vector end' : 'ring-vector'
      }));
      g.appendChild(head(c, c, r, m.i, n, m.end ? 'ring-arrow end' : 'ring-arrow'));
      if (m.label) {
        const [lx, ly] = at(c, c, labOff, m.i, n);
        const t = tag('text', { x: lx, y: ly + 4, class: 'ring-radian' });
        t.textContent = m.label;
        g.appendChild(t);
      }
      if (m.end) {
        const [ex, ey] = at(c, c, r * 0.5, m.i, n);
        const e = tag('text', { x: ex - 22, y: ey + 3, class: 'ring-end-label' });
        e.textContent = m.end;
        g.appendChild(e);
      }
    }
  }

  if (spec.pairs) {
    for (let i = 0; i < m; i++) {
      const [x1, y1] = at(c, c, r, i, n);
      const [x2, y2] = at(c, c, r, i + m, n);
      g.appendChild(tag('line', { x1, y1, x2, y2, class: 'ring-pair' }));
    }
  }

  if (spec.rests) {
    for (let i = 0; i < m; i++) {
      const v = phi(i, spec.rests, m), resting = v === 0;
      const [x1, y1] = at(c, c, r, i, n);
      const [x2, y2] = at(c, c, r, i + m, n);
      g.appendChild(tag('line', {
        x1, y1, x2, y2, class: resting ? 'ring-rest' : 'ring-pair'
      }));
      for (const end of [i, i + m]) {
        const [lx, ly] = at(c, c, numOff + 3, end, n);
        const t = tag('text', {
          x: lx, y: ly + 3.5, class: resting ? 'ring-steps rest' : 'ring-steps'
        });
        t.textContent = resting ? 'rest' : v;
        g.appendChild(t);
      }
    }
  }

  if (spec.axis !== undefined) {
    const [x1, y1] = at(c, c, r, spec.axis, n);
    const [x2, y2] = at(c, c, r, spec.axis + m, n);
    g.appendChild(tag('line', { x1, y1, x2, y2, class: 'ring-axis' }));
    for (const [i, label] of [[spec.axis, 'top'], [spec.axis + m, 'bottom']]) {
      const [x, y] = at(c, c, r, i, n);
      g.appendChild(tag('circle', { cx: x, cy: y, r: 5.5, class: 'ring-end' }));
      const [lx, ly] = at(c, c, labOff, i, n);
      const t = tag('text', { x: lx, y: ly + 4, class: 'ring-label' });
      t.textContent = label;
      g.appendChild(t);
    }
  }

  if (spec.arrival !== undefined) {
    const [x, y] = at(c, c, r, spec.arrival, n);
    g.appendChild(tag('line', { x1: c, y1: c, x2: x, y2: y, class: 'ring-arrival' }));
    g.appendChild(tag('circle', { cx: x, cy: y, r: 5.5, class: 'ring-arrival-dot' }));
    const [lx, ly] = at(c, c, labOff, spec.arrival, n);
    const t = tag('text', { x: lx, y: ly + 4, class: 'ring-label arrival' });
    t.textContent = spec.arrivalLabel || 'arrived';
    g.appendChild(t);
  }
  return g;
}

function phi(x, rests, m) {
  return Math.min(...rests.map((s) => fold(x - s, m)));
}

for (const host of document.querySelectorAll('[data-ring]')) {
  host.appendChild(ring(JSON.parse(host.dataset.ring)));
}
