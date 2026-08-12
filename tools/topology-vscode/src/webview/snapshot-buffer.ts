import { decodeViewFrame } from "./three/decode/buffer-decode-view";

let latestViewFrame: ArrayBuffer | null = null;

type SnapshotListener = () => void;
const viewListeners = new Set<SnapshotListener>();

export function setLatestViewFrame(buf: ArrayBuffer, gen = 0): void {
  latestViewFrame = buf;
  noteGen(gen);
  for (const fn of viewListeners) fn();
}

function genTable<T>(tables: Map<number, Map<number, T>>, gen: number): Map<number, T> {
  let t = tables.get(gen);
  if (!t) {
    t = new Map<number, T>();
    tables.set(gen, t);

    for (const key of tables.keys()) {
      if (key < gen - 1) tables.delete(key);
    }
  }
  return t;
}

let latestGen = 0;

function noteGen(gen: number): void {
  if (gen > latestGen) latestGen = gen;
}

export function resetSceneIdentityForTest(): void {
  latestGen = 0;
  edgeStream.clear();
  nodeStream.clear();
  interiorStream.clear();
}

export function getLatestViewFrame(): ArrayBuffer | null {
  return latestViewFrame;
}

export function subscribeViewFrame(fn: SnapshotListener): () => void {
  viewListeners.add(fn);
  return () => {
    viewListeners.delete(fn);
  };
}

function makeRowStreamTable(withVersion: boolean) {
  const tables: Map<number, Map<number, ArrayBuffer>> = new Map();
  const listeners = new Set<SnapshotListener>();
  let version = 0;
  return {
    set(row: number, buf: ArrayBuffer, gen: number): void {
      noteGen(gen);
      genTable(tables, gen).set(row, buf);
      if (withVersion) version++;
      for (const fn of listeners) fn();
    },
    get(): ReadonlyMap<number, ArrayBuffer> {
      return genTable(tables, latestGen);
    },
    subscribe(fn: SnapshotListener): () => void {
      listeners.add(fn);
      return () => {
        listeners.delete(fn);
      };
    },
    getVersion(): number {
      return version;
    },
    clear(): void {
      tables.clear();
    },
  };
}

const edgeStream = makeRowStreamTable(false);

export function setLatestEdgeStreamFrame(row: number, buf: ArrayBuffer, gen = 0): void {
  edgeStream.set(row, buf, gen);
}

export function getLatestEdgeStreamFrames(): ReadonlyMap<number, ArrayBuffer> {
  return edgeStream.get();
}

export function subscribeEdgeStreamFrame(fn: SnapshotListener): () => void {
  return edgeStream.subscribe(fn);
}

const nodeStream = makeRowStreamTable(true);
const interiorStream = makeRowStreamTable(true);

export function setLatestNodeStreamFrame(row: number, buf: ArrayBuffer, gen = 0): void {
  nodeStream.set(row, buf, gen);
}

export function getLatestNodeStreamFrames(): ReadonlyMap<number, ArrayBuffer> {
  return nodeStream.get();
}

export function getNodeStreamVersion(): number {
  return nodeStream.getVersion();
}

export function subscribeNodeStreamFrame(fn: SnapshotListener): () => void {
  return nodeStream.subscribe(fn);
}

export function setLatestInteriorStreamFrame(row: number, buf: ArrayBuffer, gen = 0): void {
  interiorStream.set(row, buf, gen);
}

export function getLatestInteriorStreamFrames(): ReadonlyMap<number, ArrayBuffer> {
  return interiorStream.get();
}

export function getInteriorStreamVersion(): number {
  return interiorStream.getVersion();
}

export function subscribeInteriorStreamFrame(fn: SnapshotListener): () => void {
  return interiorStream.subscribe(fn);
}
