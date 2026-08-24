
export interface RowLeafValues<N extends string> {
  bytes: (row: number, name: N) => DataView | undefined;
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

export function makeRowLeafValues<N extends string>(
  pathsDir: string,
  names: readonly N[],
): RowLeafValues<N> {
  const latest = new Map<number, Map<string, DataView>>();
  const observed = new Set<number>();
  let started = false;
  let template: string | undefined;

  const split = (row: number, buf: ArrayBuffer): void => {
    const dv = new DataView(buf);
    const into = new Map<string, DataView>();
    let off = 0;
    for (const name of names) {
      if (off + 4 > buf.byteLength) return;
      const len = dv.getUint32(off, true);
      off += 4;
      if (off + len > buf.byteLength) return;
      into.set(name, new DataView(buf, off, len));
      off += len;
    }
    latest.set(row, into);
  };

  const start = (): void => {
    if (started || typeof window === "undefined") return;
    started = true;
    const pump = async () => {
      for (;;) {
        await new Promise((r) => requestAnimationFrame(() => r(undefined)));
        const scene = window.BEADNETWORK_SCENE_BASE;
        const src = window.BEADNETWORK_SRC_BASE;
        if (!scene || !src) continue;
        if (template === undefined) {
          const p = await readUrl(`${src}/${pathsDir}/block.bin`, "default");
          if (!p) continue;
          template = new TextDecoder().decode(p);
        }
        const rows = [...observed];
        const tag = ++seq;
        await Promise.all(rows.map(async (row) => {
          const rel = template!.replace("{row}", String(row));
          const buf = await readUrl(`${scene}/${rel}?r=${tag}`, "no-store");
          if (buf) split(row, buf);
        }));
      }
    };
    void pump();
  };

  return {
    bytes: (row, name) => {
      observed.add(row);
      start();
      return latest.get(row)?.get(name);
    },
  };
}
