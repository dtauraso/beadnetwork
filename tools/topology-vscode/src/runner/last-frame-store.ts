export abstract class LastFrameStore {
  protected lastViewFrame: ArrayBuffer | undefined;
  protected lastEdgeFrames: Map<number, ArrayBuffer> = new Map();
  protected lastNodeFrames: Map<number, ArrayBuffer> = new Map();
  protected lastInteriorFrames: Map<number, ArrayBuffer> = new Map();
  protected lastBeadFrames: Map<number, ArrayBuffer> = new Map();

  getLastViewFrame(): ArrayBuffer | undefined {
    return this.lastViewFrame?.slice(0);
  }

  getLastEdgeFrames(): Array<{ row: number; buffer: ArrayBuffer }> {
    return Array.from(this.lastEdgeFrames, ([row, buffer]) => ({ row, buffer: buffer.slice(0) }));
  }

  getLastNodeFrames(): Array<{ row: number; buffer: ArrayBuffer }> {
    return Array.from(this.lastNodeFrames, ([row, buffer]) => ({ row, buffer: buffer.slice(0) }));
  }

  getLastInteriorFrames(): Array<{ row: number; buffer: ArrayBuffer }> {
    return Array.from(this.lastInteriorFrames, ([row, buffer]) => ({ row, buffer: buffer.slice(0) }));
  }

  getLastBeadFrames(): Array<{ row: number; buffer: ArrayBuffer }> {
    return Array.from(this.lastBeadFrames, ([row, buffer]) => ({ row, buffer: buffer.slice(0) }));
  }
}
