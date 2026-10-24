class BCards {
  constructor({ item, methods = {}, card = 1 } = {}) {
    this.item = item;
    this.methods = { ...methods };
    this.card = card;
  }

  fix(nodes) { if (!nodes.has(this.item)) this.item = [...nodes][0]; }

  toJSON() {
    const { item, methods, card } = this;
    return { item: item && item.name, methods, card };
  }

  static from({ item, methods, card } = {}, by) { return new BCards({ item: by.get(item), methods, card }); }
}
