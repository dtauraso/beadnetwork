function bPreset() {
  const a = new BNode({ name: 'a', s: 1, at: { r: 2, phi: 3, theta: 0 } });
  const b = new BNode({ name: 'b', s: 1, at: { r: 0, phi: 0, theta: 0 } });
  const c = new BNode({ name: 'c', s: 1, at: { r: 0, phi: 0, theta: 0 } });
  const edges = [
    new BEdge({ from: a, to: b, start: { phi: 2, theta: 0, r: 2 } }),
    new BEdge({ from: b, to: c, start: { phi: 4, theta: 0, r: 2 } }),
  ];
  const p = new BBuilder({
    s: 1, rho: 0.5, Tphi: 12, Ttheta: 12, cr: 0.5,
    made: 3, item: c, methods: { atob: true, pair: true }, step: 1,
    open: 0, seqs: [new BSeq({ name: 'sequence 1', steps: [new BStep({ nodes: [a, b, c], edges })] })],
  });
  edges.forEach(e => e.from.send(p, e));
  return p;
}
