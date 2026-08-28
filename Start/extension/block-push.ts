import * as fs from "fs";
import * as path from "path";
import * as vscode from "vscode";

const TICK_MS = 16;

type Want = {
  pathsDir: string;
  cadenceMs: number;
  rows?: number[];
  rel?: string;
  dueAt: number;
  seen: Map<string, string>;
  done: boolean;
};

type WantMsg = {
  type?: string;
  pathsDir?: string;
  rows?: number[];
  cadenceMs?: number;
};

export function armBlockPush(
  panel: vscode.WebviewPanel,
  srcRoot: string,
  sceneRoot: string,
): vscode.Disposable {
  const wants = new Map<string, Want>();

  const relOf = (want: Want): string | undefined => {
    if (want.rel !== undefined) return want.rel;
    try {
      want.rel = fs.readFileSync(path.join(srcRoot, want.pathsDir, "block.bin"), "utf8").trim();
      return want.rel;
    } catch {
      return undefined;
    }
  };

  const push = (want: Want, rel: string, row?: number): void => {
    const file = path.join(sceneRoot, rel);
    let stamp: string;
    try {
      const st = fs.statSync(file);
      stamp = `${String(st.mtimeMs)}:${String(st.size)}`;
    } catch {
      return;
    }
    if (want.seen.get(rel) === stamp) return;

    let bytes: Buffer;
    try {
      bytes = fs.readFileSync(file);
    } catch {
      return;
    }
    want.seen.set(rel, stamp);

    void panel.webview.postMessage({
      type: "block",
      pathsDir: want.pathsDir,
      rel,
      row,
      b64: bytes.toString("base64"),
    });
  };

  const timer = setInterval(() => {
    const now = Date.now();
    for (const want of wants.values()) {
      if (want.done || now < want.dueAt) continue;
      want.dueAt = now + Math.max(want.cadenceMs, TICK_MS);

      const rel = relOf(want);
      if (rel === undefined) continue;

      if (want.rows === undefined) {
        push(want, rel);
      } else {
        for (const row of want.rows) push(want, rel.replace("{row}", String(row)), row);
      }

      if (want.cadenceMs === 0 && want.seen.size > 0) want.done = true;
    }
  }, TICK_MS);

  const sub = panel.webview.onDidReceiveMessage((raw: unknown) => {
    const msg = raw as WantMsg | undefined;
    if (msg?.type !== "want-block" || typeof msg.pathsDir !== "string") return;

    const existing = wants.get(msg.pathsDir);
    if (existing) {
      if (msg.rows) existing.rows = msg.rows;
      existing.done = false;
      existing.dueAt = 0;
      return;
    }
    wants.set(msg.pathsDir, {
      pathsDir: msg.pathsDir,
      cadenceMs: typeof msg.cadenceMs === "number" ? msg.cadenceMs : 100,
      rows: msg.rows,
      dueAt: 0,
      seen: new Map(),
      done: false,
    });
  });

  return new vscode.Disposable(() => {
    clearInterval(timer);
    sub.dispose();
  });
}
