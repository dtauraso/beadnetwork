class BSeq {
  constructor({ name, picked = [], steps = [new BStep()] }) {
    this.name = name;
    this.steps = steps;
    this.picked = picked;
  }

  get open() { return this.steps[0]; }

  pick(e) { if (this.open.edges.has(e) && !this.picked.includes(e)) this.picked = [...this.picked, e]; }
  unpick(e) { this.picked = this.picked.filter(x => x !== e); }
  prune() { this.picked = this.picked.filter(e => this.open.edges.has(e)); return this.picked; }

  copy(name) { return BSeq.from({ ...JSON.parse(JSON.stringify(this)), name }); }

  toJSON() { return { name: this.name, steps: this.steps, picked: this.picked.map(e => e.name) }; }

  static from({ name, picked = [], steps }) {
    const made = steps.map(BStep.from), by = new Map([...made[0].edges].map(e => [e.name, e]));
    return new BSeq({ name, steps: made, picked: picked.map(k => by.get(k)).filter(Boolean) });
  }
}
