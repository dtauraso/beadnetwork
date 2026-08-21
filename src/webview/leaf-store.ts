export interface LeafStore<N extends string> {
  bytes: (name: N) => DataView | undefined;
  f32: (name: N, fallback?: number) => number;
  i32: (name: N, fallback?: number) => number;
  u32: (name: N, fallback?: number) => number;
  u8: (name: N, fallback?: number) => number;
  f32Run: (name: N) => Float32Array | null;
  i32Run: (name: N) => Int32Array | null;
  u32Run: (name: N) => Uint32Array | null;
  text: (name: N) => Uint8Array | null;
}

let seq = 0;

async function readUrl(url: string, cache: RequestCache): Promise<ArrayBuffer | undefined> {
  try {
    const res = await fetch(url, { cache });
    return res.ok ? await res.arrayBuffer() : undefined;
  } catch {
    return undefined;
  }
}

const READ_INTERVAL_MS = 100;

export function makeLeafStore<N extends string>(
  pathsDir: string,
  names: readonly N[],
): LeafStore<N> {
  const latest = new Map<string, DataView>();
  let started = false;
  let paths: Map<string, string> | undefined;

  const loadPaths = async (src: string): Promise<Map<string, string> | undefined> => {
    const bufs = await Promise.all(
      names.map((n) => readUrl(`${src}/${pathsDir}/${n}.bin`, "default")),
    );
    const out = new Map<string, string>();
    for (const [i, name] of names.entries()) {
      const buf = bufs[i];
      if (buf === undefined) return undefined;
      out.set(name, new TextDecoder().decode(buf));
    }
    return out;
  };

  const start = (): void => {
    if (started || typeof window === "undefined") return;
    started = true;
    const pump = async () => {
      for (;;) {
        const scene = window.WIREFOLD_SCENE_BASE;
        const src = window.WIREFOLD_SRC_BASE;
        if (scene && src) {
          paths ??= await loadPaths(src);
          if (paths) {
            const bufs = await Promise.all(
              names.map((n) => readUrl(`${scene}/${paths?.get(n) ?? ""}?r=${++seq}`, "no-store")),
            );
            for (const [i, name] of names.entries()) {
              const b = bufs[i];
              if (b === undefined) continue;
              latest.set(name, new DataView(b));
            }
          }
        }
        await new Promise((r) => setTimeout(r, READ_INTERVAL_MS));
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
