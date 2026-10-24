class BSeq {
  constructor({ name, picked = [], steps = [new BStep()], lock, locking }) {
    this.name = name;
    this.steps = steps;
    this.picked = picked;
    this.lock = lock;
    this.locking = locking;
  }

  get open() { return this.steps[0]; }

  pick(e) { if (this.open.edges.has(e) && !this.picked.includes(e)) this.picked = [...this.picked, e]; }
  unpick(e) { this.picked = this.picked.filter(x => x !== e); }
  prune() { this.picked = this.picked.filter(e => this.open.edges.has(e)); return this.picked; }

  copy(name) { return BSeq.from({ ...JSON.parse(JSON.stringify(this)), name }); }

  toJSON() {
    const { name, steps, picked, lock, locking } = this;
    return { name, steps, picked: picked.map(e => e.name),
      ...(lock && (lock.phi || lock.theta) ? { lock } : {}), ...(locking ? { locking } : {}) };
  }

  static from({ name, picked = [], steps, lock, locking }) {
    const made = steps.map(BStep.from), by = new Map([...made[0].edges].map(e => [e.name, e]));
    return new BSeq({ name, steps: made, picked: picked.map(k => by.get(k)).filter(Boolean), lock, locking });
  }
}
