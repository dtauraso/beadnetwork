import * as fs from "fs";
import * as path from "path";
import * as vscode from "vscode";
import { resolveScenePath } from "./runner/scene-path";

type Want = {
  pathsDir: string;
  rows?: number[];
  rel?: string;
  once: boolean;
  sent: Set<string>;
  stamps: Map<string, Buffer>;
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
  anchorPath: string,
): vscode.Disposable {
  const wants = new Map<string, Want>();
  const watchers = new Map<string, fs.FSWatcher>();

  let sceneRoot = resolveScenePath(anchorPath);

  const push = (want: Want, rel: string, row?: number): void => {
    let bytes: Buffer;
    try {
      bytes = fs.readFileSync(path.join(sceneRoot, rel));
    } catch {
      return;
    }
    const last = want.stamps.get(rel);
    if (last !== undefined && last.equals(bytes)) return;

    want.stamps.set(rel, bytes);
    want.sent.add(rel);

    try {
      fs.appendFileSync(
        path.join(path.dirname(anchorPath), ".probe", "blocks.log"),
        `${new Date().toISOString().slice(11, 23)} ${want.pathsDir} ${rel} ${String(bytes.length)}B\n`,
      );
    } catch { /* eslint-disable-line no-empty */ }

    void panel.webview.postMessage({
      type: "block",
      pathsDir: want.pathsDir,
      rel,
      row,
      b64: bytes.toString("base64"),
    });
  };

  const watchDirOf = (rel: string, onChanged: (rel: string) => void): void => {
    const dir = path.dirname(path.join(sceneRoot, rel));
    if (watchers.has(dir)) return;
    try {
      watchers.set(
        dir,
        fs.watch(dir, (_event, name) => {
          if (typeof name !== "string") return;
          onChanged(path.join(path.dirname(rel), name));
        }),
      );
    } catch { /* eslint-disable-line no-empty */ }
  };

  const relsOf = (want: Want): { rel: string; row?: number }[] => {
    if (want.rel === undefined) {
      try {
        want.rel = fs.readFileSync(path.join(srcRoot, want.pathsDir, "block.bin"), "utf8").trim();
      } catch {
        return [];
      }
    }
    if (want.rows === undefined) return [{ rel: want.rel }];
    return want.rows.map((row) => ({ rel: want.rel!.replace("{row}", String(row)), row }));
  };

  const changed = (changedRel: string): void => {
    for (const want of wants.values()) {
      if (want.once && want.sent.size > 0) continue;
      for (const { rel, row } of relsOf(want)) {
        if (rel === changedRel) push(want, rel, row);
      }
    }
  };

  const arm = (want: Want): void => {
    for (const { rel, row } of relsOf(want)) {
      push(want, rel, row);
      if (!want.once) watchDirOf(rel, changed);
    }
  };

  const rearm = (): void => {
    const next = resolveScenePath(anchorPath);
    if (next === sceneRoot) return;
    sceneRoot = next;

    for (const w of watchers.values()) w.close();
    watchers.clear();
    for (const want of wants.values()) {
      want.sent.clear();
      want.stamps.clear();
      arm(want);
    }
  };

  let selectionWatcher: fs.FSWatcher | undefined;
  try {
    selectionWatcher = fs.watch(path.join(anchorPath, "view", "scene"), () => { rearm(); });
  } catch { /* eslint-disable-line no-empty */ }

  const sub = panel.webview.onDidReceiveMessage((raw: unknown) => {
    const msg = raw as WantMsg | undefined;
    if (msg?.type !== "want-block" || typeof msg.pathsDir !== "string") return;

    const existing = wants.get(msg.pathsDir);
    if (existing) {
      if (msg.rows) existing.rows = msg.rows;
      arm(existing);
      return;
    }
    const want: Want = {
      pathsDir: msg.pathsDir,
      rows: msg.rows,
      once: msg.cadenceMs === 0,
      sent: new Set(),
      stamps: new Map(),
    };
    wants.set(msg.pathsDir, want);
    arm(want);
  });

  return new vscode.Disposable(() => {
    selectionWatcher?.close();
    for (const w of watchers.values()) w.close();
    watchers.clear();
    sub.dispose();
  });
}
