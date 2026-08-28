export interface RowLeafValues<N extends string> {
  bytes: (row: number, name: N) => DataView | undefined;
}

export interface RowBlockMsg {
  pathsDir: string;
  rel: string;
  row?: number;
  b64: string;
}

declare global {
  interface Window {
    BEADNETWORK_BLOCKS?: {
      on: (fn: (m: RowBlockMsg) => void) => void;
      want: (pathsDir: string, cadenceMs: number) => void;
      wantRows: (pathsDir: string, rows: number[], cadenceMs: number) => void;
      bytes: (b64: string) => ArrayBuffer;
    };
  }
}

const ROW_READ_INTERVAL_MS = 100;

export function makeRowLeafValues<N extends string>(
  pathsDir: string,
  names: readonly N[],
): RowLeafValues<N> {
  const latest = new Map<number, Map<string, DataView>>();
  const observed = new Set<number>();
  let started = false;
  let asked = 0;

  const split = (row: number, buf: ArrayBuffer): void => {
    const dv = new DataView(buf);
    const into = new Map<string, DataView>();
    let off = 0;
    for (const name of names) {
      if (off + 4 > buf.byteLength) break;
      const len = dv.getUint32(off, true);
      off += 4;
      if (off + len > buf.byteLength) break;
      into.set(name, new DataView(buf, off, len));
      off += len;
    }
    latest.set(row, into);
  };

  const askForRows = (): void => {
    const blocks = window.BEADNETWORK_BLOCKS;
    if (!blocks || observed.size === asked) return;
    asked = observed.size;
    blocks.wantRows(pathsDir, [...observed], ROW_READ_INTERVAL_MS);
  };

  const start = (): void => {
    if (started || typeof window === "undefined") return;
    const blocks = window.BEADNETWORK_BLOCKS;
    if (!blocks) return;
    started = true;

    blocks.on((m) => {
      if (m.pathsDir !== pathsDir || m.row === undefined) return;
      split(m.row, blocks.bytes(m.b64));
    });
    askForRows();
  };

  return {
    bytes: (row, name) => {
      observed.add(row);
      start();
      askForRows();
      return latest.get(row)?.get(name);
    },
  };
}
