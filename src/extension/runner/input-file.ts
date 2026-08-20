import * as fs from "fs";
import * as path from "path";

export const INPUT_FILE_REL = path.join("view", "input", "current.bin");

export function writeInputFile(scenePath: string, record: ArrayBuffer | Uint8Array): void {
  const bytes = record instanceof Uint8Array ? record : new Uint8Array(record);
  const dest = path.join(scenePath, INPUT_FILE_REL);
  const tmp = `${dest}.tmp`;
  try {
    fs.mkdirSync(path.dirname(dest), { recursive: true });
    fs.writeFileSync(tmp, bytes);
    fs.renameSync(tmp, dest);
  } catch {
    /* eslint-disable-line no-empty */
  }
}
