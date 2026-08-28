
export interface LeafValues<N extends string> {
  bytes: (name: N) => DataView | undefined;
  f32: (name: N, fallback?: number) => number;
  i32: (name: N, fallback?: number) => number;
  u32: (name: N, fallback?: number) => number;
  u8: (name: N, fallback?: number) => number;
  f64: (name: N, fallback?: number) => number;
  i64: (name: N, fallback?: number) => number;
  f64Run: (name: N) => number[];
  f32Run: (name: N) => Float32Array | null;
  i32Run: (name: N) => Int32Array | null;
  u32Run: (name: N) => Uint32Array | null;
  text: (name: N) => Uint8Array | null;
}


async function readUrl(url: string, cache: RequestCache): Promise<ArrayBuffer | undefined> {
  try {
    const res = await fetch(url, { cache });
    return res.ok ? await res.arrayBuffer() : undefined;
  } catch {
    return undefined;
  }
}

const READ_INTERVAL_MS = 100;

export function makeLeafValues<N extends string>(
  pathsDir: string,
  names: readonly N[],

  cadence: "interval" | "frame" = "interval",
): LeafValues<N> {
  const latest = new Map<string, DataView>();
  let started = false;
  let blockPath: string | undefined;

  const split = (buf: ArrayBuffer): void => {
    const dv = new DataView(buf);
    let off = 0;
    for (const name of names) {
      if (off + 4 > buf.byteLength) return;
      const len = dv.getUint32(off, true);
      off += 4;
      if (off + len > buf.byteLength) return;
      latest.set(name, new DataView(buf, off, len));
      off += len;
    }
  };

  const start = (): void => {
    if (started || typeof window === "undefined") return;
    started = true;
    const pump = async () => {
      for (;;) {
        const scene = window.BEADNETWORK_SCENE_BASE;
        const src = window.BEADNETWORK_SRC_BASE;
        if (scene && src) {
          if (blockPath === undefined) {
            const p = await readUrl(`${src}/${pathsDir}/block.bin`, "default");
            if (p) blockPath = new TextDecoder().decode(p);
          }
          if (blockPath !== undefined) {
            const buf = await readUrl(`${scene}/${blockPath}`, "no-store");
            if (buf) split(buf);
          }
        }
        await (cadence === "frame"
          ? new Promise((r) => requestAnimationFrame(() => r(undefined)))
          : new Promise((r) => setTimeout(r, READ_INTERVAL_MS)));
      }
    };
    void pump();
  };

  const view = (name: N): DataView | undefined => {
    start();
    return latest.get(name);
  };

  const run = <T>(name: N, width: number, make: (n: number) => T, set: (out: T, i: number, v: DataView) => void): T | null => {
    const v = view(name);
    if (!v || v.byteLength === 0) return null;
    const out = make(v.byteLength / width);
    for (let i = 0; i < v.byteLength / width; i++) set(out, i, v);
    return out;
  };

  return {
    bytes: view,
    f32: (name, fallback = 0) => {
      const v = view(name);
      return v && v.byteLength >= 4 ? v.getFloat32(0, true) : fallback;
    },
    i32: (name, fallback = 0) => {
      const v = view(name);
      return v && v.byteLength >= 4 ? v.getInt32(0, true) : fallback;
    },
    f64: (name, fallback = 0) => {
      const v = view(name);
      return v && v.byteLength >= 8 ? v.getFloat64(0, true) : fallback;
    },
    i64: (name, fallback = 0) => {
      const v = view(name);
      return v && v.byteLength >= 8 ? Number(v.getBigInt64(0, true)) : fallback;
    },
    f64Run: (name) => {
      const v = view(name);
      if (!v || v.byteLength === 0) return [];
      const out: number[] = [];
      for (let i = 0; i < v.byteLength / 8; i++) out.push(v.getFloat64(i * 8, true));
      return out;
    },
    u32: (name, fallback = 0) => {
      const v = view(name);
      return v && v.byteLength >= 4 ? v.getUint32(0, true) : fallback;
    },
    u8: (name, fallback = 0) => {
      const v = view(name);
      return v && v.byteLength >= 1 ? v.getUint8(0) : fallback;
    },
    f32Run: (name) => run(name, 4, (n) => new Float32Array(n), (o, i, v) => { o[i] = v.getFloat32(i * 4, true); }),
    i32Run: (name) => run(name, 4, (n) => new Int32Array(n), (o, i, v) => { o[i] = v.getInt32(i * 4, true); }),
    u32Run: (name) => run(name, 4, (n) => new Uint32Array(n), (o, i, v) => { o[i] = v.getUint32(i * 4, true); }),
    text: (name) => {
      const v = view(name);
      if (!v) return null;
      return new Uint8Array(v.buffer, v.byteOffset, v.byteLength);
    },
  };
}
