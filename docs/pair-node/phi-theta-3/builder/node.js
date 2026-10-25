class BNode {
  constructor({ name, s, at }) {
    this.name = name;
    this.s = s;
    this.at = at;
    this.out = [];
    this.in = [];
  }

  scale(builder) { return this.s || builder.s; }

  place(builder) { return builder.place(this.at); }

  index() { return this.at; }

  sit(at) { this.at = at; }

  take(builder, from, d, P = from.place(builder)) { this.sit(builder.gridAt(bAdd(P, builder.vec(from.scale(builder), d)))); }

  send(builder, e, P) { e.to.take(builder, this, e.dir(builder), P); }

  moved(builder, keep = []) {
    [...this.in, ...this.out].filter(e => !keep.includes(e)).forEach(e => e.remeasure(builder));
  }

  setIdx(builder, k, v) {
    this.sit({ ...this.at, [k]: v });
    this.moved(builder);
  }

  setS(builder, value) {
    const was = this.scale(builder), k = value / was;
    if (!bWhole(value) || value < 1) return `${this.name}'s s ${bN(value)}: not a whole number of at least 1, so it stays at ${was}`;
    this.s = value;
    this.out.forEach(e => e.scaleAngles(k));
    return '';
  }

  addOut(e) { this.out = [...this.out, e]; }
  addIn(e) { this.in = [...this.in, e]; }
  drop(e) {
    this.out = this.out.filter(x => x !== e);
    this.in = this.in.filter(x => x !== e);
  }

  toJSON() { return { name: this.name, s: this.s, at: this.at }; }

  static from({ name, s, at, idx }) { return new BNode({ name, s, at: at || idx }); }
}
