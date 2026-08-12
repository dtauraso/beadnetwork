import * as fs from "fs/promises";
import * as fsSync from "fs";
import * as path from "path";
import * as vscode from "vscode";
import { PROBE_DIR, PROBE_FILES, isProbeTraceEnabled } from "../probe-files";

const ERROR_LABELS = new Set([
  "window-error", "unhandled-rejection", "render-error",
  "early-window-error", "early-unhandled-rejection",
  "load-error",
]);

let pendingTs: Promise<void> = Promise.resolve();
let pendingTsErrors: Promise<void> = Promise.resolve();

export async function appendWebviewLog(
  entry: string,
  documentUri: vscode.Uri | undefined,
): Promise<void> {

  if (documentUri === undefined) return;
  let parsed: { label?: string } | undefined;
  try {
    const raw: unknown = JSON.parse(entry);
    if (typeof raw === "object" && raw !== null) {
      const label = (raw as Record<string, unknown>).label;
      parsed = typeof label === "string" ? { label } : {};
    }
  } catch { /* eslint-disable-line no-empty */ }
  const isError = parsed?.label !== undefined && ERROR_LABELS.has(parsed.label);
  if (isError) {
    pendingTsErrors = pendingTsErrors.then(() => doAppend(entry, documentUri, PROBE_FILES.tsErrors));
    return pendingTsErrors;
  } else {
    if (!isProbeTraceEnabled()) {
      return;
    }
    pendingTs = pendingTs.then(() => doAppend(entry, documentUri, PROBE_FILES.ts));
    return pendingTs;
  }
}

async function doAppend(entry: string, documentUri: vscode.Uri, filename: string): Promise<void> {
  const folder = vscode.workspace.getWorkspaceFolder(documentUri);
  const baseDir = folder ? folder.uri.fsPath : path.dirname(documentUri.fsPath);
  const dir = path.join(baseDir, PROBE_DIR);
  const file = path.join(dir, filename);
  try {
    await fs.mkdir(dir, { recursive: true });
    await fs.appendFile(file, entry + "\n", "utf8");
  } catch (err) {
    console.warn("topology editor: webview-log append failed", err);

    try {
      const errFile = path.join(dir, PROBE_FILES.tsErrors);
      fsSync.appendFileSync(errFile, JSON.stringify({ ts_ms: Date.now(), src: "ts-ext", label: "ext.webview-log-append-failed", message: String(err) }) + "\n", "utf8");
    } catch { /* eslint-disable-line no-empty */ }
  }
}
