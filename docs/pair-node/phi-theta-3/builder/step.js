class BStep {
  constructor({ nodes = [], edges = [], cards = new BCards(), picked = [], lock, locking } = {}) {
    this.nodes = new Set();
    this.edges = new Set();
    nodes.forEach(n => this.addNode(n));
    edges.forEach(e => this.addEdge(e));
    this.cards = cards;
    this.picked = picked;
    this.lock = lock;
    this.locking = locking;
    this.fixItem();
    this.prune();
  }

  allows(e, picked = this.picked) { return !picked.some(x => x.to === e.to && x.from !== e.from); }

  pick(e) { if (this.edges.has(e) && !this.picked.includes(e) && this.allows(e)) this.picked = [...this.picked, e]; }

  unpick(e) { this.picked = this.picked.filter(x => x !== e); }

  prune() {
    this.picked = this.picked.reduce((kept, e) => (this.edges.has(e) && this.allows(e, kept) ? [...kept, e] : kept), []);
    return this.picked;
  }

  offered() { return this.recalc().filter(e => this.allows(e)); }

  recalc() { return [...this.edges].filter(e => !this.picked.includes(e)); }

  find(name) { return [...this.nodes].find(n => n.name === name); }

  fixItem() { this.cards.fix(this.nodes); }

  addNode(n) {
    this.nodes.add(n);
    return n;
  }

  addEdge(e) {
    e.from.addOut(e);
    e.to.addIn(e);
    this.edges.add(e);
    return e;
  }

  connect(builder, from, to) {
    const e = this.addEdge(new BEdge({ from, to, start: { phi: 0, theta: 0, r: 2 } }));
    e.start = e.measure(builder);
    return e;
  }

  dropEdge(e) {
    e.unlink();
    this.edges.delete(e);
    this.unpick(e);
  }

  dropNode(builder, n) {
    [...n.out, ...n.in].forEach(e => this.dropEdge(e));
    this.nodes.delete(n);
    n.gone = true;
    this.fixItem();
  }

  moveAll(builder, edges = [...this.edges]) {
    const by = new Map(edges.map(e => [e.to, e])), done = new Set();
    const loop = n => { const r = [n]; for (let x = by.get(n); x && r.length <= by.size; x = by.get(x.from)) { if (x.from === n) return r; r.push(x.from); } return null; };
    const ring = r => {
      const v = r.map(n => builder.vec(by.get(n).from.scale(builder), by.get(n).dir(builder)));
      const gap = v.reduce(bAdd).map(x => x / r.length), mid = ps => ps.reduce(bAdd).map(x => x / ps.length);
      const ps = [[0, 0, 0]];
      for (let i = r.length - 1; i > 0; i--) ps.unshift(bAdd(ps[0], bSub(v[i - 1], gap)));
      const shift = bSub(mid(r.map(n => n.place(builder))), mid(ps));
      r.forEach((n, i) => { n.sit(builder.gridAt(bAdd(ps[i], shift))); done.add(n); });
    };
    const put = n => {
      const e = by.get(n);
      if (!e || done.has(n)) return;
      const r = loop(n);
      if (r) return ring(r);
      done.add(n);
      put(e.from);
      e.from.send(builder, e);
    };
    edges.forEach(e => put(e.to));
    edges.forEach(e => e.to.moved(builder, edges));
  }

  settle(builder) {
    const by = new Map([...this.nodes].map(n => [n.name, n]));
    const at = n => (n.at.from ? bAdd(at(by.get(n.at.from)), builder.vec(by.get(n.at.from).scale(builder), n.at)) : builder.place(n.at));
    [...this.nodes].filter(n => n.at.from).map(n => [n, at(n)]).forEach(([n, p]) => n.sit(builder.gridAt(p)));
    this.edges.forEach(e => { if (!e.start) e.start = e.measure(builder); });
  }

  toJSON() {
    const { cards, picked, lock, locking } = this;
    return { cards, edges: [...this.edges], nodes: [...this.nodes], picked: picked.map(e => e.name),
      ...(lock && (lock.phi || lock.theta) ? { lock } : {}), ...(locking ? { locking } : {}) };
  }

  static from({ nodes, edges, cards, item, methods, card, picked = [], lock, locking }) {
    const made = nodes.map(BNode.from), by = new Map(made.map(n => [n.name, n]));
    const rows = edges || nodes.flatMap(({ name, out = [] }) => out.map(r => ({ ...r, from: name })));
    const es = rows.map(r => BEdge.from(r, by)).filter(Boolean), named = new Map(es.map(e => [e.name, e]));
    return new BStep({ nodes: made, edges: es, cards: BCards.from(cards || { item, methods, card }, by), picked: picked.map(k => named.get(k)).filter(Boolean), lock, locking });
  }
}
