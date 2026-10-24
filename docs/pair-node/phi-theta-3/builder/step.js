class BStep {
  constructor({ nodes = [], edges = [], item, methods = {}, card = 1 } = {}) {
    this.nodes = new Set();
    this.edges = new Set();
    nodes.forEach(n => this.addNode(n));
    edges.forEach(e => this.addEdge(e));
    this.item = item;
    this.methods = { ...methods };
    this.card = card;
    this.fixItem();
  }

  find(name) { return [...this.nodes].find(n => n.name === name); }

  fixItem() { if (!this.nodes.has(this.item)) this.item = [...this.nodes][0]; }

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
  }

  dropNode(builder, n) {
    n.leave(builder);
    [...n.out, ...n.in].forEach(e => this.dropEdge(e));
    this.nodes.delete(n);
    n.gone = true;
    this.fixItem();
  }

  moveAll(builder, edges = [...this.edges]) {
    edges.forEach(() => edges.forEach(e => e.from.send(builder, e)));
    edges.forEach(e => e.to.moved(builder, edges));
  }

  toJSON() {
    const { item, methods, card } = this;
    return { item: item && item.name, methods, card, edges: [...this.edges], nodes: [...this.nodes] };
  }

  static from({ nodes, edges, item, methods, card }) {
    const made = nodes.map(BNode.from), by = new Map(made.map(n => [n.name, n]));
    made.forEach((n, i) => n.seat(nodes[i], by));
    const rows = edges || nodes.flatMap(({ name, out = [] }) => out.map(r => ({ ...r, from: name })));
    return new BStep({ nodes: made, edges: rows.map(r => BEdge.from(r, by)).filter(Boolean), item: by.get(item), methods, card });
  }
}
