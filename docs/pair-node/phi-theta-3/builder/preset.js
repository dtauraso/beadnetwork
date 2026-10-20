function bPreset() {
  const a = bMake({ name: 'a', s: 1, idx: { r: 2, phi: 3, theta: 0 } });
  const b = bMake({ name: 'b', s: 1, idx: { r: 0, phi: 0, theta: 0 } });
  const c = bMake({ name: 'c', s: 1, idx: { r: 0, phi: 0, theta: 0 } });
  const p = {
    s: 1, rho: 0.5, Tphi: 12, Ttheta: 12, cr: 0.5,
    picked: [], made: 3, item: c, methods: { atob: true, pair: true }, step: 1,
    open: 0, seqs: [{ name: 'sequence 1', steps: [{ edges: new Set(), nodes: new Set([a, b, c]) }] }],
  };
  const rows = [
    { from: a, to: b, start: { phi: 2, theta: 0, r: 2 } },
    { from: b, to: c, start: { phi: 4, theta: 0, r: 2 } },
  ];
  const [ab, bc] = rows.map(r => bAddTree(p, r));
  [ab, bc].forEach(t => bPut(p, t.ab.row));
  return p;
}
