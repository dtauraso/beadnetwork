class BEdge {
  constructor({ from, to, start, delta = { phi: 0, theta: 0 }, increment = { phi: 1, theta: 1 }, lock, locking }) {
    this.from = from;
    this.to = to;
    this.start = start;
    this.delta = delta;
    this.increment = increment;
    this.lock = lock;
    this.locking = locking;
  }

  get name() { return `${this.from.name}→${this.to.name}`; }

  dir(builder, s = this.from.scale(builder)) {
    return { ...this.start, phi: bWrap(s, this.start.phi + this.delta.phi), theta: bWrap(s, this.start.theta + this.delta.theta) };
  }

  back(builder, a = this.dir(builder)) {
    const sa = this.from.scale(builder), sb = this.to.scale(builder);
    return { phi: bWrap(sb, a.phi * sb / sa + 6 * sb), theta: bWrap(sb, a.theta * sb / sa), r: a.r };
  }

  measure(builder, P = this.from.place(builder), T = this.to.place(builder)) {
    const s = this.from.scale(builder), v = bSub(T, P), len = bLen(v), tick = 2 * Math.PI / (12 * s), half = 6 * s;
    const was = this.start ? this.dir(builder) : { phi: 0, theta: 0 };
    if (len < 1e-9) return { phi: was.phi, theta: was.theta, r: 0 };
    const side = Math.hypot(v[0], v[2]), phi = Math.atan2(side, v[1]) / tick;
    const theta = side < 1e-9 * len ? was.theta : Math.atan2(v[2], v[0]) / tick;
    const gap = (a, b) => { const d = bWrap(s, a - b); return Math.min(d, 12 * s - d); };
    const off = (p, t) => gap(p, was.phi) + gap(t, was.theta);
    const [p, t] = off(phi, theta) <= off(-phi, theta + half) ? [phi, theta] : [-phi, theta + half];
    return { phi: bWrap(s, p), theta: bWrap(s, t), r: len / builder.rho };
  }

  remeasure(builder) {
    const s = this.from.scale(builder), v = this.measure(builder);
    this.start = { phi: bWrap(s, v.phi - this.delta.phi), theta: bWrap(s, v.theta - this.delta.theta), r: v.r };
  }

  move(builder) {
    this.from.send(builder, this);
    this.to.moved(builder, [this]);
  }

  setStart(builder, k, v) {
    this.start[k] = k === 'r' ? v : bWrap(this.from.scale(builder), v);
    this.move(builder);
    return '';
  }

  step(builder, k, sign) {
    this.delta[k] = bWrap(this.from.scale(builder), this.delta[k] + sign * this.increment[k]);
  }

  scaleAngles(k) {
    this.start = { ...this.start, phi: this.start.phi * k, theta: this.start.theta * k };
  }

  ends(builder) {
    const t = this.dir(builder), P = this.from.place(builder);
    return { edge: this, P, T: this.to.place(builder), t, d: builder.vec(this.from.scale(builder), t) };
  }

  links(builder) {
    const { P, T, t } = this.ends(builder);
    return [
      { from: this.from, to: this.to, P, T, t, s: this.from.scale(builder) },
      { from: this.to, to: this.from, P: T, T: P, t: this.back(builder, t), s: this.to.scale(builder) },
    ];
  }

  unlink() {
    this.from.drop(this);
    this.to.drop(this);
  }

  toJSON() {
    const { start, delta, increment, lock, locking } = this;
    return { from: this.from.name, to: this.to.name, start, delta, increment,
      ...(lock && (lock.phi || lock.theta) ? { lock } : {}), ...(locking ? { locking } : {}) };
  }

  static from(record, by) {
    const from = by.get(record.from), to = by.get(record.to);
    return from && to ? new BEdge({ ...record, from, to }) : null;
  }
}
