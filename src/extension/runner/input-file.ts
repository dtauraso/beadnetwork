import * as fs from "fs";
import * as path from "path";
import { IN_EVENT_KINDS } from "../../Input/Codec/input-layout-gen";

export const INPUT_DIR_REL = path.join("view", "input");

export function inputSlotRel(kind: string): string {
  return path.join(INPUT_DIR_REL, `${kind}.bin`);
}

export function writeInputFile(scenePath: string, record: ArrayBuffer | Uint8Array): void {
  const bytes = record instanceof Uint8Array ? record : new Uint8Array(record);

  const kind = IN_EVENT_KINDS[bytes[1] ?? -1];
  if (kind === undefined) return;

  const dest = path.join(scenePath, inputSlotRel(kind));
  const tmp = `${dest}.tmp`;
  try {
    fs.mkdirSync(path.dirname(dest), { recursive: true });
    fs.writeFileSync(tmp, bytes);
    fs.renameSync(tmp, dest);
  } catch {
    /* eslint-disable-line no-empty */
  }
}
