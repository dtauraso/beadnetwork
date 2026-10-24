const B_NODE = {
  place(builder) { return bPlace(builder, this.idx); },

  take(builder, p) {
    const J = bIndexOf(builder, p);
    this.idx = { ...J, theta: J.theta < 0 ? J.theta + builder.Ttheta : J.theta };
  },

  send(builder, t) {
    const d = bDir(builder, t.ab.row);
    if (!d) return false;
    t.ab.b.take(builder, bAdd(this.place(builder), bVec(builder, bS(builder, this), d)));
    return true;
  },

  hear(builder, t, T) {
    const row = t.ab.row, s = bS(builder, this), v = bMeasure(builder, row, this.place(builder), T);
    if (v) row.start = { phi: bWrap(s, v.phi - row.delta.phi), theta: bWrap(s, v.theta - row.delta.theta), r: v.r };
  },

  moved(builder, keep = []) {
    const at = this.place(builder);
    this.in.filter(t => !keep.includes(t)).forEach(t => t.ab.a.hear(builder, t, at));
    this.out.filter(t => !keep.includes(t)).forEach(t => this.hear(builder, t, t.ab.b.place(builder)));
  },

  setIdx(builder, k, v) {
    this.idx[k] = v;
    this.moved(builder);
  },

  setStart(builder, t, k, v) {
    t.ab.row.start[k] = k === 'r' ? v : bWrap(bS(builder, this), v);
    bMove(builder, t);
    return '';
  },

  step(builder, t, k, sign) {
    const r = t.ab.row;
    r.delta[k] = bWrap(bS(builder, this), r.delta[k] + sign * r.increment[k]);
  },

  setS(builder, value) {
    const was = bS(builder, this), k = value / was;
    if (!bWhole(value) || value < 1) return `${this.name}'s s ${bN(value)}: not a whole number of at least 1, so it stays at ${was}`;
    this.s = value;
    this.out.forEach(t => {
      const r = t.ab.row;
      r.start = { ...r.start, phi: r.start.phi * k, theta: r.start.theta * k };
    });
    return '';
  },

  addOut(t) { this.out = [...this.out, t]; },
  addIn(t) { this.in = [...this.in, t]; },
  drop(t) {
    this.out = this.out.filter(x => x !== t);
    this.in = this.in.filter(x => x !== t);
  },
};
