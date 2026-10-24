class BNode {
  constructor({ name, s, at }) {
    this.name = name;
    this.s = s;
    this.at = at;
    this.out = [];
    this.in = [];
    this.kids = [];
  }

  scale(builder) { return this.s || builder.s; }

  place(builder) {
    const { from } = this.at;
    return from ? bAdd(from.place(builder), builder.vec(from.scale(builder), this.at)) : builder.place(this.at);
  }

  index(builder) {
    const { from, r, phi, theta } = this.at;
    return from ? builder.gridAt(this.place(builder)) : { r, phi, theta };
  }

  sit(at) {
    const was = this.at && this.at.from;
    if (was) was.kids = was.kids.filter(k => k !== this);
    this.at = at;
    if (at.from) at.from.kids = [...at.from.kids, this];
  }

  hold(builder, change) {
    const held = this.kids.map(k => [k, k.place(builder)]);
    change();
    held.forEach(([k, p]) => k.rest(builder, p));
  }

  rest(builder, p) {
    const { from } = this.at;
    this.at = { from, ...builder.local(from.scale(builder), bSub(p, from.place(builder))) };
  }

  root(builder) { this.sit(this.index(builder)); }

  take(builder, from, d) {
    if (from === this) {
      const p = bAdd(this.place(builder), builder.vec(this.scale(builder), d));
      return this.hold(builder, () => this.sit(builder.gridAt(p)));
    }
    for (let x = from; x.at.from; x = x.at.from) if (x.at.from === this) { x.root(builder); break; }
    this.hold(builder, () => this.sit({ from, phi: d.phi, theta: d.theta, r: d.r }));
  }

  send(builder, e) { e.to.take(builder, this, e.dir(builder)); }

  moved(builder, keep = []) {
    [...this.in, ...this.out].filter(e => !keep.includes(e)).forEach(e => e.remeasure(builder));
  }

  setIdx(builder, k, v) {
    const I = { ...this.index(builder), [k]: v };
    this.hold(builder, () => this.sit(I));
    this.moved(builder);
  }

  setS(builder, value) {
    const was = this.scale(builder), k = value / was;
    if (!bWhole(value) || value < 1) return `${this.name}'s s ${bN(value)}: not a whole number of at least 1, so it stays at ${was}`;
    this.hold(builder, () => {
      this.s = value;
      this.out.forEach(e => e.scaleAngles(k));
    });
    return '';
  }

  addOut(e) { this.out = [...this.out, e]; }
  addIn(e) { this.in = [...this.in, e]; }
  drop(e) {
    this.out = this.out.filter(x => x !== e);
    this.in = this.in.filter(x => x !== e);
  }

  leave(builder) {
    this.kids.forEach(k => k.root(builder));
    this.root(builder);
  }

  toJSON() {
    const { from, ...at } = this.at;
    return { name: this.name, s: this.s, at: from ? { from: from.name, ...at } : at };
  }

  static from({ name, s }) { return new BNode({ name, s }); }

  seat({ at, idx }, by) {
    const { from, ...place } = at || idx;
    this.sit(by.get(from) ? { from: by.get(from), ...place } : place);
  }
}
