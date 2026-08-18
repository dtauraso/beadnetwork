import { splitLengthPrefixed } from "./framing-columns";

export class ColumnStore {
  private readonly latest = new Map<number, Buffer>();
  private readonly partial = new Map<number, Buffer>();

  private version = 0;

  handle(col: number, chunk: Buffer): boolean {
    const { values, rest } = splitLengthPrefixed(this.partial.get(col) ?? Buffer.alloc(0), chunk);
    this.partial.set(col, rest);
    if (values.length === 0) return false;

    this.latest.set(col, values[values.length - 1]!);
    this.version++;
    return true;
  }

  get(col: number): Buffer | undefined {
    return this.latest.get(col);
  }

  entries(): ReadonlyMap<number, Buffer> {
    return this.latest;
  }

  getVersion(): number {
    return this.version;
  }
}
