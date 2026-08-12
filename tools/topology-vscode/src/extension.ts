import * as fs from "fs";
import * as path from "path";
import * as vscode from "vscode";
import { BuildAndRunRunner } from "./runCommand";
import { buildBinary } from "./goBuild";
import { shouldRestartAfterBuild, TrailingDebouncer } from "./hotRestart";
import { hashBundle, isHostReloadEnabled, shouldReloadHost } from "./hostReload";
import type { HostToWebviewMsg } from "./messages";
import { buildWebviewHtml } from "./extension/html";
import { handleMessage } from "./extension/handle-message";
import { serveDocsOpen } from "./extension/docs-open";
import { PROBE_FILES } from "./probe-files";

export function activate(context: vscode.ExtensionContext) {
  context.subscriptions.push(
    vscode.commands.registerCommand("topology.openEditor", (uri?: vscode.Uri) => {
      openTopologyEditor(context, uri);
    }),
  );
  armHostReloadWatcher(context);

  serveDocsOpen(context);
}

function armHostReloadWatcher(context: vscode.ExtensionContext): void {
  const hostBundlePath = path.join(context.extensionPath, "out", "extension.js");
  const hostChannel = vscode.window.createOutputChannel("topology host-reload");
  context.subscriptions.push(hostChannel);

  let loadedHash: string | undefined;
  try {
    loadedHash = hashBundle(fs.readFileSync(hostBundlePath));
  } catch { /* eslint-disable-line no-empty */ }
  const hostWatcher = vscode.workspace.createFileSystemWatcher(
    new vscode.RelativePattern(
      vscode.Uri.file(path.join(context.extensionPath, "out")),
      "extension.js",
    ),
  );
  context.subscriptions.push(hostWatcher);
  const debouncer = new TrailingDebouncer(250);
  context.subscriptions.push({ dispose: () => debouncer.dispose() });

  let reloading = false;
  const maybeReload = () => {
    debouncer.schedule(() => {
      if (reloading) return;

      if (!isHostReloadEnabled()) return;
      let newHash: string;
      try {
        newHash = hashBundle(fs.readFileSync(hostBundlePath));
      } catch {
        return; 
      }
      if (!shouldReloadHost(loadedHash, newHash)) return;

      reloading = true;
      hostChannel.appendLine("[topology] extension host bundle changed — reloading window");
      vscode.window.setStatusBarMessage(
        "Topology: extension updated, reloading window…",
        3000,
      );

      void vscode.commands.executeCommand("workbench.action.reloadWindow");
    });
  };
  hostWatcher.onDidChange(maybeReload);
  hostWatcher.onDidCreate(maybeReload);
}

function resetProbeLogs(repoRoot: string): void {
  try {
    const probeDir = path.join(repoRoot, ".probe");
    fs.mkdirSync(probeDir, { recursive: true });

    for (const name of Object.values(PROBE_FILES)) {
      fs.writeFileSync(path.join(probeDir, name), "");
    }
  } catch { /* eslint-disable-line no-empty */ }
}

function resolveTopologyPath(folderUri?: vscode.Uri): string | undefined {
  if (folderUri) return folderUri.fsPath;

  const folder = vscode.workspace.workspaceFolders?.[0];
  if (folder) {
    const candidate = path.join(folder.uri.fsPath, "topology");
    if (fs.existsSync(candidate)) return candidate;
  }
  return undefined;
}

function armBundleWatcher(panel: vscode.WebviewPanel, context: vscode.ExtensionContext): vscode.FileSystemWatcher {
  const bundleWatcher = vscode.workspace.createFileSystemWatcher(
    new vscode.RelativePattern(
      vscode.Uri.file(path.join(context.extensionPath, "out")),
      "webview.js",
    ),
  );
  console.log("[topology] bundleWatcher armed for", path.join(context.extensionPath, "out", "webview.js"));
  let pending: NodeJS.Timeout | undefined;
  const reload = (kind: string) => () => {
    console.log("[topology] bundleWatcher fired:", kind);
    if (pending) clearTimeout(pending);
    pending = setTimeout(() => {
      console.log("[topology] hot-reload: re-rendering webview.html");
      panel.webview.html = buildWebviewHtml(panel.webview, context.extensionPath);
    }, 150);
  };
  bundleWatcher.onDidChange(reload("change"));
  bundleWatcher.onDidCreate(reload("create"));
  return bundleWatcher;
}

function armGoWatcher(
  repoRoot: string | undefined,
  runner: BuildAndRunRunner,
  panel: vscode.WebviewPanel,
): vscode.FileSystemWatcher | undefined {
  if (!repoRoot) return undefined;
  const binPath = path.join(repoRoot, ".wirefold-cache", "wirefold");
  const goErrorsFile = path.join(repoRoot, ".probe", "go-errors.jsonl");
  const goChannel = vscode.window.createOutputChannel("topology go-build");
  const goWatcher = vscode.workspace.createFileSystemWatcher(
    new vscode.RelativePattern(repoRoot, "**/*.go"),
  );
  const debouncer = new TrailingDebouncer(250);
  const rebuild = () => {
    debouncer.schedule(() => {
      const res = buildBinary(repoRoot, binPath);
      if (shouldRestartAfterBuild(res)) {
        goChannel.appendLine("[go] rebuilt wirefold");

        if (runner.restart()) {
          goChannel.appendLine("[go] hot-restarting sim");
        }
      } else if (!res.ok) {
        goChannel.appendLine(`[go] build error: ${res.error}`);
        try {
          fs.mkdirSync(path.dirname(goErrorsFile), { recursive: true });
          fs.appendFileSync(
            goErrorsFile,
            JSON.stringify({ ts_ms: Date.now(), src: "go", kind: "error", message: res.error }) + "\n",
            "utf8",
          );
        } catch { /* eslint-disable-line no-empty */ }
      }

    });
  };
  goWatcher.onDidChange(rebuild);
  goWatcher.onDidCreate(rebuild);
  goWatcher.onDidDelete(rebuild);

  panel.onDidDispose(() => {
    debouncer.dispose();
    goChannel.dispose();
  });
  return goWatcher;
}

function wireMessageHandler(
  panel: vscode.WebviewPanel,
  folderUri: vscode.Uri | undefined,
  runner: BuildAndRunRunner,
  post: (msg: HostToWebviewMsg) => void,
): void {
  panel.webview.onDidReceiveMessage((raw) => {
    const workspaceFolder = folderUri ? vscode.workspace.getWorkspaceFolder(folderUri) : undefined;

    const logUri = workspaceFolder?.uri ?? folderUri ?? vscode.workspace.workspaceFolders?.[0]?.uri;
    void handleMessage(raw, { logUri, runner, post }).catch((err: unknown) => {
      console.error("topology: handleMessage failed", err);
    });
  });
}

function openTopologyEditor(context: vscode.ExtensionContext, folderUri?: vscode.Uri): void {
  const topologyPath = resolveTopologyPath(folderUri);

  const probeRoot = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
  if (probeRoot) resetProbeLogs(probeRoot);

  const panel = vscode.window.createWebviewPanel(
    "topology.editor",
    "Topology Editor",
    vscode.ViewColumn.One,
    {
      enableScripts: true,
      retainContextWhenHidden: true,
      localResourceRoots: [vscode.Uri.file(path.join(context.extensionPath, "out"))],
    },
  );
  panel.webview.html = buildWebviewHtml(panel.webview, context.extensionPath);

  const post = (msg: HostToWebviewMsg): void => {
    void panel.webview.postMessage(msg);
  };
  const runner = new BuildAndRunRunner(

    (snapshot) => post(snapshot),
  );

  const bundleWatcher = armBundleWatcher(panel, context);

  const repoRoot = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
  const goWatcher = armGoWatcher(repoRoot, runner, panel);

  context.subscriptions.push(runner);

  panel.onDidDispose(() => {
    bundleWatcher?.dispose();
    goWatcher?.dispose();
    runner.dispose();
  });

  wireMessageHandler(panel, folderUri, runner, post);

  runner.run(topologyPath);
}
