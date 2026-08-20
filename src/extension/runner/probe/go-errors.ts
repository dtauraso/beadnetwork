import { logfmt } from "../../probe/logfmt";
import * as fs from "fs";

export function goErrorLine(message: string): string {
  return logfmt({ ts_ms: Date.now(), src: "go", kind: "error", message }) + "\n";
}

export function appendGoError(goErrorsFile: string | undefined, message: string): void {
  if (!goErrorsFile) return;
  try {
    fs.appendFileSync(goErrorsFile, goErrorLine(message), "utf8");
  } catch { /* eslint-disable-line no-empty */ }
}
