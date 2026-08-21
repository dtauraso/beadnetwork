import { logfmt } from "./probe/logfmt";
import { IN_KIND_RAW_INPUT } from "../Input/input-layout-gen";
import { writeInputFile } from "./runner/input-file";
import { resolveScenePath } from "./runner/counts";
import * as fs from "fs";
import * as path from "path";
import * as vscode from "vscode";
import { BuildAndRunRunner } from "./runCommand";
import {
  parseWebviewToHost,
  type HostToWebviewMsg,
  type WebviewToHostMsg,
} from "../Input/messages";
import { appendWebviewLog } from "./webview-log";
import { PROBE_DIR, PROBE_FILES } from "./probe-files";
import { resolveRepoRoot } from "./repo-root";
import { BUF_BLOCK_TAG_COLUMN, BUF_BLOCK_TAG_VIEW, BUF_BLOCK_TAG_EDGE_STREAM, BUF_BLOCK_TAG_NODE_STREAM, BUF_BLOCK_TAG_INTERIOR_STREAM, BUF_BLOCK_TAG_BEAD_STREAM } from "../Buffer/frame-tags";

export type MessageCtx = {
  logUri: vscode.Uri | undefined;
  runner: BuildAndRunRunner;
  post: (msg: HostToWebviewMsg) => void;
  scenePath: string;
  anchorPath?: string;
};

function assertNever(msg: never): never {
  throw new Error(`handle-message: unhandled webview message kind ${String((msg as { type?: unknown }).type)}`);
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
        const entry = logfmt({
          ts_ms: Date.now(),
          msgType: msg.type,
          message: error.message,
          stack: error.stack ?? "",
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
        for (const { row, buffer } of runner.getLastBeadFrames()) {
          ctx.post({ type: "buffer-snapshot", buffer, tag: BUF_BLOCK_TAG_BEAD_STREAM, row, gen: runner.currentGen() });
        }

        for (const { col, buffer } of runner.getLastColumnValues()) {
          ctx.post({ type: "buffer-snapshot", buffer, tag: BUF_BLOCK_TAG_COLUMN, row: col, gen: runner.currentGen() });
        }
      }
      return;
    }
    case "webview-log":
      await appendWebviewLog(msg.entry, logUri);
      return;
    case "go-record": {
      if (!runner.isRunning()) return;
      // Raw input is the CURRENT input, so it goes to the file the gesture
      // goroutine reads when it wakes. Sending it down the pipe queued it, and
      // a queue of input replays history after the fingers stop. Edits and save
      // are one-shot commands, not state, and stay on stdin.
      const first = new Uint8Array(msg.record instanceof Uint8Array ? msg.record : new Uint8Array(msg.record))[0];
      if (first === IN_KIND_RAW_INPUT) {
        // No fallback to the pipe: Go no longer reads raw input from it, so a
        // silent fallback would drop every gesture while looking like it works.
        writeInputFile(
          ctx.anchorPath === undefined ? ctx.scenePath : resolveScenePath(ctx.anchorPath),
          msg.record,
        );
        return;
      }
      runner.writeStdin(msg.record);
      return;
    }
    // LIVE_CASES_END

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
  return resolveRepoRoot(vscode.workspace.workspaceFolders?.[0]?.uri.fsPath);
}
