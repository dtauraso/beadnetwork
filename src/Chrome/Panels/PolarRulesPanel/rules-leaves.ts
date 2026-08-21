import { RULES_VALUE_NAMES, type RulesValueName } from "./rules-values-gen";

const latest = new Map<string, DataView>();

let seq = 0;
async function readBytes(url: string): Promise<ArrayBuffer | undefined> {
  return readUrl(`${url}?r=${++seq}`, "no-store");
}

async function readGenerated(url: string): Promise<ArrayBuffer | undefined> {
  return readUrl(url, "default");
}

async function readUrl(url: string, cache: RequestCache): Promise<ArrayBuffer | undefined> {
  try {
    const res = await fetch(url, { cache });
    return res.ok ? await res.arrayBuffer() : undefined;
  } catch {
    return undefined;
  }
}

async function loadPaths(src: string): Promise<Map<string, string> | undefined> {
  const bufs = await Promise.all(
    RULES_VALUE_NAMES.map((n) => readGenerated(`${src}/Chrome/Panels/PolarRulesPanel/paths/${n}.bin`)),
  );
  const out = new Map<string, string>();
  for (const [i, name] of RULES_VALUE_NAMES.entries()) {
    const buf = bufs[i];
    if (buf === undefined) return undefined;
    out.set(name, new TextDecoder().decode(buf));
  }
  return out;
}

const READ_INTERVAL_MS = 100;

let started = false;
export function startRulesReads(): void {
  if (started || typeof window === "undefined") return;
  started = true;
  let paths: Map<string, string> | undefined;
  const pump = async () => {
    for (;;) {
      const scene = window.WIREFOLD_SCENE_BASE;
      const src = window.WIREFOLD_SRC_BASE;
      if (scene && src) {
        paths ??= await loadPaths(src);
        if (paths) {
          const bufs = await Promise.all(
            RULES_VALUE_NAMES.map((n) => readBytes(`${scene}/${paths?.get(n) ?? ""}`)),
          );
          for (const [i, name] of RULES_VALUE_NAMES.entries()) {
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
}

function view(name: RulesValueName): DataView | undefined {
  startRulesReads();
  return latest.get(name);
}

export function rulesF32(name: RulesValueName, fallback = 0): number {
  const v = view(name);
  return v && v.byteLength >= 4 ? v.getFloat32(0, true) : fallback;
}

export function rulesI32(name: RulesValueName, fallback = 0): number {
  const v = view(name);
  return v && v.byteLength >= 4 ? v.getInt32(0, true) : fallback;
}

export function rulesU8(name: RulesValueName, fallback = 0): number {
  const v = view(name);
  return v && v.byteLength >= 1 ? v.getUint8(0) : fallback;
}

export function rulesBytes(name: RulesValueName): DataView | undefined {
  return view(name);
}

export function rulesF32Run(name: RulesValueName): Float32Array | null {
  const v = view(name);
  if (!v || v.byteLength === 0) return null;
  const out = new Float32Array(v.byteLength / 4);
  for (let i = 0; i < out.length; i++) out[i] = v.getFloat32(i * 4, true);
  return out;
}

export function rulesU32Run(name: RulesValueName): Uint32Array | null {
  const v = view(name);
  if (!v || v.byteLength === 0) return null;
  const out = new Uint32Array(v.byteLength / 4);
  for (let i = 0; i < out.length; i++) out[i] = v.getUint32(i * 4, true);
  return out;
}

export function rulesI32Run(name: RulesValueName): Int32Array | null {
  const v = view(name);
  if (!v || v.byteLength === 0) return null;
  const out = new Int32Array(v.byteLength / 4);
  for (let i = 0; i < out.length; i++) out[i] = v.getInt32(i * 4, true);
  return out;
}

export function rulesU8Run(name: RulesValueName): Uint8Array | null {
  const v = view(name);
  if (!v || v.byteLength === 0) return null;
  return new Uint8Array(v.buffer, v.byteOffset, v.byteLength);
}

export function rulesText(name: RulesValueName): Uint8Array | null {
  const v = view(name);
  if (!v) return null;
  return new Uint8Array(v.buffer, v.byteOffset, v.byteLength);
}
