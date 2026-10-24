const bName = k => (k < 26 ? 'abcdefghijklmnopqrstuvwxyz'[k] : `n${k}`);
const bNameAt = name => (name.length === 1 ? name.charCodeAt(0) - 97 : +name.slice(1));

class BBuilder {
  constructor({ s, rho, Tphi, Ttheta, cr, made, item, methods = {}, step, open, seqs, lock, locking }) {
    Object.assign(this, { s, rho, Tphi, Ttheta, cr, made, item, methods, step, open, seqs, lock, locking });
    this.made ??= 1 + Math.max(-1, ...seqs.flatMap(q => q.steps.flatMap(st => [...st.nodes].map(n => bNameAt(n.name)))).filter(Number.isFinite));
    this.fixItem();
  }

  openSeq() { return this.seqs[this.open]; }
  openStep() { return this.openSeq().open; }
  shown() { return [...this.openStep().nodes]; }

  fixItem() {
    const shown = this.shown();
    if (!shown.includes(this.item)) this.item = shown[0];
  }

  cell() { return { r: this.cr, phi: 2 * Math.PI / this.Tphi, theta: 2 * Math.PI / this.Ttheta }; }

  place(I) {
    const c = this.cell();
    return bAdd(B_O, bCart(I.r * c.r, I.phi * c.phi, I.theta * c.theta));
  }

  indexOf(p) {
    const c = this.cell(), w = bSub(p, B_O), r = bLen(w);
    if (r < 1e-9) return { r: 0, phi: 0, theta: 0 };
    return { r: r / c.r, phi: Math.acos(Math.max(-1, Math.min(1, w[1] / r))) / c.phi, theta: Math.atan2(w[2], w[0]) / c.theta };
  }

  gridAt(p) {
    const J = this.indexOf(p);
    return { ...J, theta: J.theta < -1e-9 ? J.theta + this.Ttheta : Math.max(0, J.theta) };
  }

  vec(s, t) {
    const ang = x => x * 2 * Math.PI / (12 * s);
    const R = t.r * this.rho, ph = ang(t.phi), th = ang(t.theta);
    return [R * Math.sin(ph) * Math.cos(th), R * Math.cos(ph), R * Math.sin(ph) * Math.sin(th)];
  }

  local(s, v) {
    const len = bLen(v), tick = 2 * Math.PI / (12 * s);
    if (len < 1e-12) return { phi: 0, theta: 0, r: 0 };
    return { phi: Math.atan2(Math.hypot(v[0], v[2]), v[1]) / tick, theta: bWrap(s, Math.atan2(v[2], v[0]) / tick), r: len / this.rho };
  }

  addNode() {
    return this.openStep().addNode(new BNode({ name: bName(this.made++), at: { r: 0, phi: 0, theta: 0 } }));
  }

  dropNode(n) {
    this.openStep().dropNode(this, n);
    this.fixItem();
  }

  addSeq() {
    this.seqs = [...this.seqs, this.openSeq().copy(`sequence ${this.seqs.length + 1}`)];
    return this.seqs.length - 1;
  }

  dropSeq() {
    this.seqs = this.seqs.filter((q, i) => i !== this.open);
    this.renumber();
    return Math.max(0, this.open - 1);
  }

  renumber() {
    this.seqs.forEach((q, i) => { if (/^sequence \d+$/.test(q.name)) q.name = `sequence ${i + 1}`; });
    return this;
  }

  toJSON() {
    const { s, rho, Tphi, Ttheta, cr, made, item, methods, step, open, seqs, lock, locking } = this;
    return { s, rho, Tphi, Ttheta, cr, made, item: item && item.name, methods, step, open, seqs,
      ...(lock && (lock.phi || lock.theta) ? { lock } : {}), ...(locking ? { locking } : {}) };
  }

  static from(record) {
    const seqs = record.seqs.map(q => ({ ...q }));
    if (record.picked) seqs[record.open].picked ??= record.picked;
    const made = seqs.map(BSeq.from), item = made[record.open].open.find(record.item);
    return new BBuilder({ ...record, seqs: made, item });
  }

  static load() {
    try {
      const saved = window.B_SAVED && JSON.parse(JSON.stringify(window.B_SAVED));
      if (saved && saved.seqs[saved.open]) return BBuilder.from(saved).renumber();
    } catch { }
    return null;
  }
}
