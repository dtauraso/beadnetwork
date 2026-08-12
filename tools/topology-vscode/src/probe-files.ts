import * as vscode from "vscode";

export const PROBE_DIR = ".probe";

export const PROBE_FILES = {

  go: "go.jsonl",
  goNode: "go-node.jsonl",
  goEdge: "go-edge.jsonl",
  goInterior: "go-interior.jsonl",
  goErrors: "go-errors.jsonl",
  ts: "ts.jsonl",
  tsErrors: "ts-errors.jsonl",
  handlerErrorLast: "handler-error-last.json",
} as const;

export const PROBE_TRACE_FILES = [
  PROBE_FILES.go,
  PROBE_FILES.goNode,
  PROBE_FILES.goEdge,
  PROBE_FILES.goInterior,
  PROBE_FILES.ts,
] as const;

export const PROBE_TRACE_SETTING_SECTION = "wirefold";
export const PROBE_TRACE_SETTING_KEY = "probe.trace";

export function isProbeTraceEnabled(): boolean {
  return vscode.workspace
    .getConfiguration(PROBE_TRACE_SETTING_SECTION)
    .get<boolean>(PROBE_TRACE_SETTING_KEY, false);
}
