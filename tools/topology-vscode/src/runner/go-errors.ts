import * as fs from "fs";

/** Format a Go-side error as a probe JSONL line (src="go", kind="error"). */
export function goErrorLine(message: string): string {
  return JSON.stringify({ ts_ms: Date.now(), src: "go", kind: "error", message }) + "\n";
}

/** Append a Go-side error line to goErrorsFile, swallowing write failures (the same
 *  best-effort append repeated at every stderr/build-error/close/error call site below). */
export function appendGoError(goErrorsFile: string | undefined, message: string): void {
  if (!goErrorsFile) return;
  try {
    fs.appendFileSync(goErrorsFile, goErrorLine(message), "utf8");
  } catch { /* swallow */ }
}
