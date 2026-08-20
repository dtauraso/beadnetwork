import * as vscode from "vscode";

export const PROBE_DIR = ".probe";

export const PROBE_FILES = {

  go: "go.jsonl",
  goErrors: "go-errors.jsonl",
  ts: "ts.jsonl",
  tsErrors: "ts-errors.jsonl",
  handlerErrorLast: "handler-error-last.json",
} as const;

export const PROBE_OWNER_DIRS = ["node", "edge", "interior", "bead"] as const;

export type ProbeOwner = (typeof PROBE_OWNER_DIRS)[number];

export function probeOwnerFile(probeDir: string, owner: ProbeOwner, row: number): string {
  return `${probeDir}/${owner}/${row}.jsonl`;
}

export const PROBE_TRACE_FILES = [
  PROBE_FILES.go,
  PROBE_FILES.ts,
] as const;

export const PROBE_TRACE_SETTING_SECTION = "wirefold";
export const PROBE_TRACE_SETTING_KEY = "probe.trace";

export function isProbeTraceEnabled(): boolean {
  return vscode.workspace
    .getConfiguration(PROBE_TRACE_SETTING_SECTION)
    .get<boolean>(PROBE_TRACE_SETTING_KEY, false);
}
