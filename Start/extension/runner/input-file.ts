import * as fs from "fs";
import * as path from "path";
import { IN_EVENT_KINDS } from "../../../Categories/Scene/Drag/input-defs";

export const INPUT_DIR_REL = path.join("view", "input");

export function inputSlotRel(kind: string): string {
  return path.join(INPUT_DIR_REL, `${kind}.bin`);
}

const FLUSH_MS = 16;

const pending = new Map<string, Uint8Array>();
const dirReady = new Set<string>();
let scene: string | undefined;
let timer: NodeJS.Timeout | undefined;

function flush(): void {
  timer = undefined;
  const root = scene;
  if (root === undefined) {
    pending.clear();
    return;
  }
  for (const [kind, bytes] of pending) {
    const dest = path.join(root, inputSlotRel(kind));
    const tmp = `${dest}.tmp`;
    try {
      if (!dirReady.has(root)) {
        fs.mkdirSync(path.dirname(dest), { recursive: true });
        dirReady.add(root);
      }
      fs.writeFileSync(tmp, bytes);
      fs.renameSync(tmp, dest);
    } catch {
      /* eslint-disable-line no-empty */
    }
  }
  pending.clear();
}

export function writeInputFile(scenePath: string, record: ArrayBuffer | Uint8Array): void {
  const bytes = record instanceof Uint8Array ? record : new Uint8Array(record);

  const kind = IN_EVENT_KINDS[bytes[1] ?? -1];
  if (kind === undefined) return;

  if (scenePath !== scene) {
    flush();
    scene = scenePath;
  }
  pending.set(kind, bytes);
  timer ??= setTimeout(flush, FLUSH_MS);
}
