import * as fs from "fs";
import * as path from "path";
import * as vscode from "vscode";
import { BuildAndRunRunner } from "./runCommand";
import type { HostToWebviewMsg } from "./messages";
import { buildWebviewHtml } from "./extension/html";
import { handleMessage } from "./extension/handle-message";
import { serveDocsOpen } from "./extension/docs-open";
import { armHostReloadWatcher } from "./extension/host-reload-watcher";
import { armBundleWatcher } from "./extension/bundle-watcher";
import { armGoWatcher } from "./extension/go-watcher";
import { PROBE_FILES } from "./probe-files";
import { resolveRepoRoot } from "./repo-root";

export function activate(context: vscode.ExtensionContext) {
  context.subscriptions.push(
    vscode.commands.registerCommand("topology.openEditor", (uri?: vscode.Uri) => {
      openTopologyEditor(context, uri);
    }),
  );
  armHostReloadWatcher(context);

  serveDocsOpen(context);
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

  const root = resolveRepoRoot(vscode.workspace.workspaceFolders?.[0]?.uri.fsPath);
  if (root) {
    const candidate = path.join(root, "topology");
    if (fs.existsSync(candidate)) return candidate;
  }
  return undefined;
}

function wireMessageHandler(
  panel: vscode.WebviewPanel,
  folderUri: vscode.Uri | undefined,
  runner: BuildAndRunRunner,
  post: (msg: HostToWebviewMsg) => void,
): void {
  panel.webview.onDidReceiveMessage((raw) => {
    const workspaceFolder = folderUri ? vscode.workspace.getWorkspaceFolder(folderUri) : undefined;

    const repoRootUri = (() => {
      const root = resolveRepoRoot(vscode.workspace.workspaceFolders?.[0]?.uri.fsPath);
      return root ? vscode.Uri.file(root) : undefined;
    })();
    const logUri = workspaceFolder?.uri ?? folderUri ?? repoRootUri;
    void handleMessage(raw, { logUri, runner, post }).catch((err: unknown) => {
      console.error("topology: handleMessage failed", err);
    });
  });
}

function openTopologyEditor(context: vscode.ExtensionContext, folderUri?: vscode.Uri): void {
  const topologyPath = resolveTopologyPath(folderUri);

  const probeRoot = resolveRepoRoot(vscode.workspace.workspaceFolders?.[0]?.uri.fsPath);
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

  const repoRoot = resolveRepoRoot(vscode.workspace.workspaceFolders?.[0]?.uri.fsPath);
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
