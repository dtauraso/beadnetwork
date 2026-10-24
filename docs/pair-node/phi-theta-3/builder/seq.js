class BSeq {
  constructor({ name, steps = [new BStep()] }) {
    this.name = name;
    this.steps = steps;
  }

  get open() { return this.steps[0]; }

  copy(name) { return BSeq.from({ ...JSON.parse(JSON.stringify(this)), name }); }

  toJSON() { return { name: this.name, steps: this.steps }; }

  static from({ name, picked, steps, lock, locking }) {
    return new BSeq({ name, steps: steps.map((st, i) => BStep.from(i ? st : { picked, lock, locking, ...st })) });
  }
}
