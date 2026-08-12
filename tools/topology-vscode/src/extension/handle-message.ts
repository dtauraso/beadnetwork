import * as fs from "fs";
import * as path from "path";
import * as vscode from "vscode";
import { BuildAndRunRunner } from "../runCommand";
import {
  parseWebviewToHost,
  type HostToWebviewMsg,
  type WebviewToHostMsg,
} from "../messages";
import { appendWebviewLog } from "./webview-log";
import { PROBE_DIR, PROBE_FILES } from "../probe-files";
import { BUF_BLOCK_TAG_VIEW, BUF_BLOCK_TAG_EDGE_STREAM, BUF_BLOCK_TAG_NODE_STREAM, BUF_BLOCK_TAG_INTERIOR_STREAM } from "../schema/buffer-layout/frame-tags";

export type MessageCtx = {
  logUri: vscode.Uri | undefined;
  runner: BuildAndRunRunner;
  post: (msg: HostToWebviewMsg) => void;
};

function assertNever(msg: never): never {
  throw new Error(`handle-message: unhandled webview message kind ${JSON.stringify(msg)}`);
}

export async function handleMessage(raw: unknown, ctx: MessageCtx): Promise<void> {
  const msg = parseWebviewToHost(raw);
  if (!msg) {
    console.warn("topology editor: ignoring malformed webview message", raw);
    return;
  }
  try {
    await dispatch(msg, ctx);
  } catch (err) {
    const error = err instanceof Error ? err : new Error(String(err));
    console.error("topology editor: unhandled message handler error", error);

    const repoRoot = workspaceRoot();
    if (repoRoot) {
      try {
        const probeDir = path.join(repoRoot, PROBE_DIR);
        fs.mkdirSync(probeDir, { recursive: true });
        const probeFile = path.join(probeDir, PROBE_FILES.handlerErrorLast);
        const entry = JSON.stringify({
          timestamp: new Date().toISOString(),
          msgType: msg.type,
          message: error.message,
          stack: error.stack ?? null,
        });
        fs.appendFileSync(probeFile, entry + "\n", "utf8");
      } catch (probeErr) {
        console.error("topology editor: could not write probe file", probeErr);
      }
    }
  }
}

async function dispatch(msg: WebviewToHostMsg, ctx: MessageCtx): Promise<void> {
  const { logUri, runner } = ctx;
  switch (msg.type) {

    // LIVE_CASES_START
    case "ready": {

      const wasRunning = runner.isRunning();
      runner.run();
      if (wasRunning) {
        const viewFrame = runner.getLastViewFrame();
        if (viewFrame) {
          ctx.post({ type: "buffer-snapshot", buffer: viewFrame, tag: BUF_BLOCK_TAG_VIEW, gen: runner.currentGen() });
        }

        for (const { row, buffer } of runner.getLastEdgeFrames()) {
          ctx.post({ type: "buffer-snapshot", buffer, tag: BUF_BLOCK_TAG_EDGE_STREAM, row, gen: runner.currentGen() });
        }

        for (const { row, buffer } of runner.getLastNodeFrames()) {
          ctx.post({ type: "buffer-snapshot", buffer, tag: BUF_BLOCK_TAG_NODE_STREAM, row, gen: runner.currentGen() });
        }
        for (const { row, buffer } of runner.getLastInteriorFrames()) {
          ctx.post({ type: "buffer-snapshot", buffer, tag: BUF_BLOCK_TAG_INTERIOR_STREAM, row, gen: runner.currentGen() });
        }
      }
      return;
    }
    case "webview-log":
      await appendWebviewLog(msg.entry, logUri);
      return;
    case "go-record":

      if (!runner.isRunning()) return;
      runner.writeStdin(msg.record);
      return;
    // LIVE_CASES_END

    // kinds stdin_reader.go's msg.Type switch dispatches (its MSG_TYPES fence). A kind here

    // DECLARED_NOT_SENT_START
    case "raw-input":
    case "save":
    case "edit":
      console.warn(`topology editor: unexpected direct "${msg.type}" message (expected via go-record)`, msg);
      return;
    // DECLARED_NOT_SENT_END
    default:
      assertNever(msg);
  }
}

function workspaceRoot(): string | undefined {
  return vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
}
